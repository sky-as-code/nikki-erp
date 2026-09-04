package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// SchedulerExtService is Sales' port onto the scheduler, for the work that has to happen on a clock
// rather than in response to a request.
//
// Sales owns two kinds of background work and they are not interchangeable. Sweeps that only tidy up
// after Sales' own records — expiring drafts, reconciling payments whose settlement was lost — run on
// the in-process cron, where a missed tick costs nothing because the next one finds the same rows.
//
// Issuing an electronic invoice is not that. It produces a legal document through a third party, so
// a run that failed has to be visibly failed and retried on a policy someone can see and change,
// which is what the scheduler module provides and a cron loop does not.
type SchedulerExtService interface {
	// EnsureJob registers a recurring job, or returns the one already registered.
	//
	// Called on every boot. Registration is idempotent on the module and key, so the second boot is
	// a no-op rather than a conflict — but note it does NOT update a job that already exists: a
	// changed schedule needs an explicit update, not a redeploy.
	EnsureJob(ctx corectx.Context, cmd EnsureJobCommand) (*EnsureJobResult, error)
}

// EnsureJobCommand describes one recurring job.
type EnsureJobCommand struct {
	// ModuleName and JobKey together identify the job, and are what makes registering it twice safe.
	ModuleName string
	JobKey     string

	// Name is what a human sees in the scheduler's own screens.
	Name string

	// CronExpression is a five-field spec interpreted in UTC. Sales states it in UTC deliberately:
	// a job pinned to a local time would drift by an hour twice a year, and an invoice-issuing job
	// that runs twice in one hour is worse than one that runs at an odd local time.
	CronExpression string

	// CommandName is the CQRS request the scheduler dispatches when the job fires, written
	// "{module}_{submodule}.{action}". The scheduler sends it with an EMPTY body — it cannot import
	// the target module's types — so the receiving command's fields must all be optional.
	CommandName string

	// MaxAttempts counts the first try. RetryIntervalSeconds is the base for the backoff between
	// them.
	MaxAttempts          int32
	RetryIntervalSeconds int32
}

// EnsureJobResult reports the job, and whether this call is what created it.
type EnsureJobResult struct {
	JobId string

	// WasCreated distinguishes a first registration from the repeat that happens on every subsequent
	// boot, which is worth logging differently: the first is news, the rest are noise.
	WasCreated bool
}
