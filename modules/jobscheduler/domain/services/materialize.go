package services

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/cronexpr"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// Reasons an occurrence produced no execution. They are recorded rather than merely returned so
// that "nothing ran" can be explained without re-deriving the decision from the schedule.
const (
	SkipNone          = ""
	SkipMisfire       = "misfire_skip"
	SkipOverlap       = "forbid_overlap"
	SkipOutsideWindow = "outside_effective_period"
	SkipNoOccurrence  = "no_occurrence"
	SkipInvalidCron   = "invalid_cron"
	SkipJobNotEnabled = "job_not_enabled"
)

// MaterializeResult is what one due job produced on one tick.
type MaterializeResult struct {
	// ExecutionScheduledFor is the occurrence an execution should be created for. Nil when the
	// occurrence was skipped.
	ExecutionScheduledFor *time.Time

	// NextOccurrenceAt is the occurrence after the one being materialized, and becomes the new
	// execution's retry ceiling. Nil means the schedule produces nothing further.
	NextOccurrenceAt *time.Time

	// NextRunAt is what the job's next_run_at should be set to. Nil parks the job: either the
	// schedule is exhausted, or its effective period has ended.
	NextRunAt *time.Time

	// SkipReason explains an absent execution, and is empty when one was produced.
	SkipReason string
}

// ShouldCreateExecution reports whether this tick produced work.
func (this MaterializeResult) ShouldCreateExecution() bool {
	return this.ExecutionScheduledFor != nil
}

// MaterializeInput is one job's state as the engine found it, plus what it needs to decide.
type MaterializeInput struct {
	CronExpression string
	NextRunAt      *time.Time
	EffectiveFrom  *time.Time
	EffectiveUntil *time.Time
	IsEnabled      bool

	// MisfirePolicy and ConcurrencyPolicy are already resolved against configuration.
	MisfirePolicy     string
	ConcurrencyPolicy string

	// HasOpenExecution reports whether this job already has an execution that has not finished.
	// It must be read inside the same transaction as the insert that follows, or two ticks can
	// both see "no open execution" and both create one.
	HasOpenExecution bool

	// MisfireThreshold is how late an occurrence may be noticed and still run normally.
	MisfireThreshold time.Duration

	Now time.Time
}

// Materialize decides what one due job should do on this tick.
//
// It is a pure function of its input so that every branch - late occurrences, exhausted schedules,
// windows that have closed, jobs already running - is an ordinary table test rather than a database
// fixture. The engine does the reading and writing; this decides.
func Materialize(in MaterializeInput) MaterializeResult {
	if !in.IsEnabled {
		// A disabled job holds no next_run_at: leaving one would have the timer wake for work
		// that will never be created.
		return MaterializeResult{SkipReason: SkipJobNotEnabled}
	}

	schedule, err := cronexpr.Parse(in.CronExpression)
	if err != nil {
		// The expression was validated when the job was created, so reaching here means a row
		// predates a parser change. Parking the job is better than crashing the loop: one broken
		// job must not stop every other job from running.
		return MaterializeResult{SkipReason: SkipInvalidCron}
	}

	if in.NextRunAt == nil {
		return MaterializeResult{SkipReason: SkipNoOccurrence}
	}
	scheduledFor := in.NextRunAt.UTC()

	// The effective period is [from, until): an occurrence exactly on `until` does not run, which
	// is what makes the two ends composable when one job's window abuts another's.
	if in.EffectiveFrom != nil && scheduledFor.Before(in.EffectiveFrom.UTC()) {
		return MaterializeResult{
			SkipReason: SkipOutsideWindow,
			NextRunAt:  advance(schedule, scheduledFor, in.EffectiveUntil),
		}
	}
	if in.EffectiveUntil != nil && !scheduledFor.Before(in.EffectiveUntil.UTC()) {
		// Past the end of the window there is no further occurrence at all, so the job parks.
		return MaterializeResult{SkipReason: SkipOutsideWindow}
	}

	// A missed occurrence is one noticed more than the threshold late. Without a threshold every
	// occurrence would count as missed, since a tick always notices one some moments after its
	// instant.
	isMisfire := in.Now.UTC().Sub(scheduledFor) > in.MisfireThreshold

	if isMisfire && in.MisfirePolicy == models.MisfireSkip {
		return MaterializeResult{
			SkipReason: SkipMisfire,
			NextRunAt:  advance(schedule, in.Now.UTC(), in.EffectiveUntil),
		}
	}

	if in.ConcurrencyPolicy == models.ConcurrencyForbidOverlap && in.HasOpenExecution {
		// The occurrence is dropped but the schedule still moves on. That advance is deliberate:
		// the next occurrence is what bounds the RUNNING execution's retries, and it is derived
		// from the schedule rather than from what actually ran.
		return MaterializeResult{
			SkipReason: SkipOverlap,
			NextRunAt:  advance(schedule, scheduledFor, in.EffectiveUntil),
		}
	}

	// A late occurrence under run_once runs once, and the schedule then resumes from now rather
	// than from the occurrence that was missed. Resuming from the missed one would replay every
	// occurrence in the gap, one tick at a time - which is catch-up under another name, and is
	// exactly what the policy exists to avoid after an outage.
	resumeFrom := scheduledFor
	if isMisfire {
		resumeFrom = in.Now.UTC()
	}

	// The ceiling comes from the schedule, not from the resumed position: it is the instant the
	// NEXT occurrence of this job is nominally due, and that is what the retry window compares to.
	nextOccurrence := advance(schedule, scheduledFor, in.EffectiveUntil)

	return MaterializeResult{
		ExecutionScheduledFor: &scheduledFor,
		NextOccurrenceAt:      nextOccurrence,
		NextRunAt:             advance(schedule, resumeFrom, in.EffectiveUntil),
	}
}

// advance returns the first occurrence strictly after `from`, or nil when the schedule produces
// none or the next one would fall outside the effective period.
//
// Nil is the single representation of "nothing further", whether the cron has no more occurrences
// or the window has closed. Collapsing the two means callers have one case to handle rather than
// two that behave identically.
func advance(schedule *cronexpr.CronExpr, from time.Time, effectiveUntil *time.Time) *time.Time {
	next, ok := schedule.Next(from)
	if !ok {
		return nil
	}
	if effectiveUntil != nil && !next.Before(effectiveUntil.UTC()) {
		return nil
	}
	return &next
}

// BuildExecutionKey derives the idempotency key for one occurrence.
//
// It is deterministic on purpose: two instances computing the same occurrence produce the same
// key, so the unique constraint on it turns a duplicate materialization into a no-op rather than
// a second execution of the same work. It is also what the action sends as its idempotency header,
// so the guarantee reaches the system being called.
func BuildExecutionKey(moduleName string, jobKey string, scheduledFor time.Time) string {
	return moduleName + ":" + jobKey + ":" + scheduledFor.UTC().Format("2006-01-02T15:04:05Z")
}
