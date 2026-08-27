package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

func baseInput(scheduledFor time.Time) MaterializeInput {
	return MaterializeInput{
		CronExpression:    "*/5 * * * *",
		NextRunAt:         &scheduledFor,
		IsEnabled:         true,
		MisfirePolicy:     models.MisfireRunOnce,
		ConcurrencyPolicy: models.ConcurrencyForbidOverlap,
		MisfireThreshold:  120 * time.Second,
		Now:               scheduledFor.Add(time.Second),
	}
}

func TestMaterializeProducesAnExecutionAndAdvancesTheSchedule(t *testing.T) {
	scheduled := at(10, 0, 0)

	result := Materialize(baseInput(scheduled))

	require.True(t, result.ShouldCreateExecution())
	assert.Equal(t, scheduled, *result.ExecutionScheduledFor)
	require.NotNil(t, result.NextOccurrenceAt)
	assert.Equal(t, at(10, 5, 0), *result.NextOccurrenceAt)
	require.NotNil(t, result.NextRunAt)
	assert.Equal(t, at(10, 5, 0), *result.NextRunAt)
	assert.Empty(t, result.SkipReason)
}

// The ceiling on an execution's retries is the NEXT occurrence of its schedule. Getting this from
// anywhere else - the resumed position, or what actually ran - would make the retry window depend
// on scheduler timing rather than on the schedule.
func TestNextOccurrenceComesFromTheScheduleNotFromWhatRan(t *testing.T) {
	scheduled := at(10, 0, 0)
	in := baseInput(scheduled)
	// Notice the occurrence very late, so the resumed position is far from the occurrence itself.
	in.Now = at(11, 47, 0)

	result := Materialize(in)

	require.True(t, result.ShouldCreateExecution())
	require.NotNil(t, result.NextOccurrenceAt)
	assert.Equal(t, at(10, 5, 0), *result.NextOccurrenceAt,
		"the ceiling is the occurrence after the one that ran, regardless of when it was noticed")
}

// A disabled job holds no next_run_at: leaving one would have the timer wake for work that will
// never be created.
func TestDisabledJobProducesNothingAndParks(t *testing.T) {
	in := baseInput(at(10, 0, 0))
	in.IsEnabled = false

	result := Materialize(in)

	assert.False(t, result.ShouldCreateExecution())
	assert.Nil(t, result.NextRunAt)
	assert.Equal(t, SkipJobNotEnabled, result.SkipReason)
}

// One broken row must not stop every other job from running, so an unparseable expression parks
// that job rather than crashing the tick.
func TestUnparseableCronParksTheJobRatherThanFailing(t *testing.T) {
	in := baseInput(at(10, 0, 0))
	in.CronExpression = "@daily"

	result := Materialize(in)

	assert.False(t, result.ShouldCreateExecution())
	assert.Nil(t, result.NextRunAt)
	assert.Equal(t, SkipInvalidCron, result.SkipReason)
}

// The effective period is [from, until). An occurrence exactly on `until` does not run.
func TestEffectivePeriodExcludesItsEndInstant(t *testing.T) {
	scheduled := at(10, 0, 0)
	in := baseInput(scheduled)
	in.EffectiveUntil = ptr(scheduled)

	result := Materialize(in)

	assert.False(t, result.ShouldCreateExecution())
	assert.Equal(t, SkipOutsideWindow, result.SkipReason)
	assert.Nil(t, result.NextRunAt, "past the window there is nothing further")
}

func TestEffectivePeriodIncludesItsStartInstant(t *testing.T) {
	scheduled := at(10, 0, 0)
	in := baseInput(scheduled)
	in.EffectiveFrom = ptr(scheduled)

	result := Materialize(in)

	assert.True(t, result.ShouldCreateExecution(), "[from, until) includes from")
}

func TestOccurrenceBeforeTheWindowOpensIsSkippedButTheScheduleAdvances(t *testing.T) {
	scheduled := at(10, 0, 0)
	in := baseInput(scheduled)
	in.EffectiveFrom = ptr(at(12, 0, 0))

	result := Materialize(in)

	assert.False(t, result.ShouldCreateExecution())
	assert.Equal(t, SkipOutsideWindow, result.SkipReason)
	require.NotNil(t, result.NextRunAt, "the job is not parked; its window simply has not opened")
	assert.Equal(t, at(10, 5, 0), *result.NextRunAt)
}

// The next occurrence must not be scheduled past the end of the window either.
func TestScheduleParksWhenTheNextOccurrenceWouldFallOutsideTheWindow(t *testing.T) {
	scheduled := at(10, 0, 0)
	in := baseInput(scheduled)
	in.EffectiveUntil = ptr(at(10, 3, 0))

	result := Materialize(in)

	require.True(t, result.ShouldCreateExecution(), "this occurrence is still inside the window")
	assert.Nil(t, result.NextRunAt, "but the following one is not, so the job parks")
	assert.Nil(t, result.NextOccurrenceAt, "and there is therefore no retry ceiling")
}

// An occurrence noticed a moment late is not a misfire. Without a threshold, every occurrence
// would be one, since a tick always notices an instant some moments after it passes.
func TestSlightlyLateOccurrenceIsNotAMisfire(t *testing.T) {
	scheduled := at(10, 0, 0)
	in := baseInput(scheduled)
	in.Now = scheduled.Add(90 * time.Second) // under the 120s threshold

	result := Materialize(in)

	require.True(t, result.ShouldCreateExecution())
	assert.Equal(t, at(10, 5, 0), *result.NextRunAt,
		"a punctual occurrence resumes from the occurrence, not from now")
}

// run_once after an outage runs the missed occurrence once and then resumes from now. Resuming
// from the missed occurrence would replay every one in the gap, a tick at a time - catch-up under
// another name, which the policy exists to avoid.
func TestMisfireRunOnceRunsOnceAndResumesFromNow(t *testing.T) {
	scheduled := at(10, 0, 0)
	in := baseInput(scheduled)
	in.Now = at(12, 2, 0) // a two-hour outage: 24 occurrences of */5 were missed

	result := Materialize(in)

	require.True(t, result.ShouldCreateExecution())
	assert.Equal(t, scheduled, *result.ExecutionScheduledFor, "the missed occurrence runs once")
	require.NotNil(t, result.NextRunAt)
	assert.Equal(t, at(12, 5, 0), *result.NextRunAt,
		"the schedule resumes from now, not from 10:05, so the backlog is not replayed")
}

func TestMisfireSkipProducesNothingAndResumesFromNow(t *testing.T) {
	scheduled := at(10, 0, 0)
	in := baseInput(scheduled)
	in.MisfirePolicy = models.MisfireSkip
	in.Now = at(12, 2, 0)

	result := Materialize(in)

	assert.False(t, result.ShouldCreateExecution())
	assert.Equal(t, SkipMisfire, result.SkipReason)
	require.NotNil(t, result.NextRunAt)
	assert.Equal(t, at(12, 5, 0), *result.NextRunAt)
}

// forbid_overlap drops the occurrence but still advances the schedule. The advance matters: the
// occurrence it moves to is what bounds the RUNNING execution's retry window.
func TestForbidOverlapDropsTheOccurrenceButStillAdvances(t *testing.T) {
	scheduled := at(10, 0, 0)
	in := baseInput(scheduled)
	in.HasOpenExecution = true

	result := Materialize(in)

	assert.False(t, result.ShouldCreateExecution())
	assert.Equal(t, SkipOverlap, result.SkipReason)
	require.NotNil(t, result.NextRunAt)
	assert.Equal(t, at(10, 5, 0), *result.NextRunAt,
		"the schedule moves on even though nothing ran")
}

func TestAllowOverlapRunsAlongsideAnOpenExecution(t *testing.T) {
	in := baseInput(at(10, 0, 0))
	in.ConcurrencyPolicy = models.ConcurrencyAllowOverlap
	in.HasOpenExecution = true

	result := Materialize(in)

	assert.True(t, result.ShouldCreateExecution())
	assert.Empty(t, result.SkipReason)
}

func TestJobWithNoNextRunProducesNothing(t *testing.T) {
	in := baseInput(at(10, 0, 0))
	in.NextRunAt = nil

	result := Materialize(in)

	assert.False(t, result.ShouldCreateExecution())
	assert.Equal(t, SkipNoOccurrence, result.SkipReason)
}

// The key is deterministic so that two instances computing the same occurrence produce the same
// string, which is what lets the unique constraint collapse a duplicate into a no-op.
func TestExecutionKeyIsDeterministicAndUtc(t *testing.T) {
	jakarta := time.FixedZone("WIB", 7*60*60)
	scheduled := time.Date(2026, time.August, 20, 17, 0, 0, 0, jakarta)

	fromLocal := BuildExecutionKey("inventory", "rebuild", scheduled)
	fromUtc := BuildExecutionKey("inventory", "rebuild", scheduled.UTC())

	assert.Equal(t, fromUtc, fromLocal, "the zone of the input must not change the key")
	assert.Equal(t, "inventory:rebuild:2026-08-20T10:00:00Z", fromUtc)
}
