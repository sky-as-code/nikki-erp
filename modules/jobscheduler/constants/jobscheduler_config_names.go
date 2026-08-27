package constants

import core "github.com/sky-as-code/nikki-erp/modules/core/constants"

// Application Configuration keys for the scheduler.
//
// Every default the scheduler runs on lives here rather than as a literal in business logic,
// so that an operator can retune the scheduler without a code change. A job persists its own
// value only for the settings it is allowed to override; everything else is read from here.
//
// Each key must also exist in config.default.yaml. The configuration service type-asserts its
// fallback argument to a string, so a missing key panics rather than falling back.
const (
	// DefaultMaxAttempts is the attempt budget for a job that does not override it. It counts
	// the first attempt, so 3 means attempts 1, 2 and 3.
	DefaultMaxAttempts core.ConfigName = "JOBSCHEDULER.DEFAULT_MAX_ATTEMPTS"

	// DefaultRetryIntervalSecs is the base delay that exponential backoff doubles.
	DefaultRetryIntervalSecs core.ConfigName = "JOBSCHEDULER.DEFAULT_RETRY_INTERVAL_SECS"

	// ExpBackoffMaxIntervalSecs caps a single backoff delay. It is deliberately not a job
	// field and cannot be raised through the API: it is a system safety limit, and an
	// operator lowering it to shed load must have it take effect on work already queued.
	ExpBackoffMaxIntervalSecs core.ConfigName = "JOBSCHEDULER.EXPBACKOFF_MAX_INTERVAL_SECS"

	// DefaultAttemptTimeoutSecs bounds one attempt end to end: the REST call, the command,
	// the worker context, and the upper bound of the lease.
	DefaultAttemptTimeoutSecs core.ConfigName = "JOBSCHEDULER.DEFAULT_ATTEMPT_TIMEOUT_SECS"

	// LeaseSafetyMarginSecs is added to the attempt timeout to form the lease, so a worker
	// that is merely slow is not reaped while it is still running.
	LeaseSafetyMarginSecs core.ConfigName = "JOBSCHEDULER.LEASE_SAFETY_MARGIN_SECS"

	// ReconciliationIntervalSecs is the fallback sweep that recovers from lost wake-ups and
	// dead instances. It is a safety net, not the resolution of the cron schedule: raising it
	// does not make jobs less punctual, and lowering it does not make them more so.
	ReconciliationIntervalSecs core.ConfigName = "JOBSCHEDULER.RECONCILIATION_INTERVAL_SECS"

	// WorkerConcurrency is the bounded worker pool size on each application instance. The
	// scheduler never starts an unbounded number of goroutines.
	WorkerConcurrency core.ConfigName = "JOBSCHEDULER.WORKER_CONCURRENCY"

	// ClaimBatchSize caps how many executions one instance takes in a single claim. It should
	// stay in proportion to WorkerConcurrency: claiming more than the pool can start holds
	// leases on work that is not running.
	ClaimBatchSize core.ConfigName = "JOBSCHEDULER.CLAIM_BATCH_SIZE"

	// DefaultConcurrencyPolicy applies when a job declares none.
	DefaultConcurrencyPolicy core.ConfigName = "JOBSCHEDULER.DEFAULT_CONCURRENCY_POLICY"

	// DefaultMisfirePolicy applies when a job declares none.
	DefaultMisfirePolicy core.ConfigName = "JOBSCHEDULER.DEFAULT_MISFIRE_POLICY"

	// DefaultJobEnabled applies when a job is created without an explicit is_enabled.
	DefaultJobEnabled core.ConfigName = "JOBSCHEDULER.DEFAULT_JOB_ENABLED"

	// HistoryRetentionDays is how long finished executions and their attempts are kept.
	HistoryRetentionDays core.ConfigName = "JOBSCHEDULER.HISTORY_RETENTION_DAYS"

	// MisfireThresholdSecs is how late an occurrence may be noticed and still run normally.
	// Beyond it the occurrence counts as missed and the misfire policy decides. Without a
	// threshold every occurrence would be "missed" by the moments between its instant and
	// the tick that notices it.
	MisfireThresholdSecs core.ConfigName = "JOBSCHEDULER.MISFIRE_THRESHOLD_SECS"
)
