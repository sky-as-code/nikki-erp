package models

// Job types. Only JobTypeTechnical is accepted in the current scope; JobTypeUser is reserved
// for the user-managed scheduling surface and is rejected at the application service.
const (
	JobTypeTechnical = "technical"
	JobTypeUser      = "user"
)

// Action types.
const (
	ActionTypeCommandBus = "command_bus"
	ActionTypeRestApi    = "rest_api"
)

// Concurrency policies. forbid_overlap prevents a second execution of the same job while one
// is still open, where "open" includes an execution waiting to retry: its retry chain still
// belongs to the earlier occurrence.
const (
	ConcurrencyForbidOverlap = "forbid_overlap"
	ConcurrencyAllowOverlap  = "allow_overlap"
)

// Misfire policies for an occurrence discovered late. catch_up_all is deliberately not
// offered: replaying every missed occurrence after an outage is a stampede, not a recovery.
const (
	MisfireRunOnce = "run_once"
	MisfireSkip    = "skip"
)

// Execution lifecycle.
const (
	ExecutionStatusQueued       = "queued"
	ExecutionStatusRunning      = "running"
	ExecutionStatusWaitingRetry = "waiting_retry"
	ExecutionStatusSucceeded    = "succeeded"
	ExecutionStatusFailed       = "failed"
	ExecutionStatusCancelled    = "cancelled"
)

// Attempt lifecycle. AttemptStatusAbandoned means the lease expired without the attempt
// reporting an outcome, which is how the death of an instance becomes visible to the others.
const (
	AttemptStatusRunning   = "running"
	AttemptStatusSucceeded = "succeeded"
	AttemptStatusFailed    = "failed"
	AttemptStatusAbandoned = "abandoned"
)

// Terminal failure codes written to an execution's failure_code.
const (
	// FailureRetryWindowExpired means the next retry would have landed at or after the
	// following occurrence, so it was cancelled and the remaining attempt quota forfeited.
	FailureRetryWindowExpired = "RETRY_WINDOW_EXPIRED"

	// FailureMaxAttemptsReached means the attempt budget was spent.
	FailureMaxAttemptsReached = "MAX_ATTEMPTS_REACHED"

	// FailureNonRetryable means the action failed in a way that repeating would not fix.
	FailureNonRetryable = "NON_RETRYABLE"
)

// IsExecutionTerminal reports whether an execution has reached a state it will never leave.
// Retention deletes only terminal executions, and only a terminal execution stops blocking a
// forbid_overlap job.
func IsExecutionTerminal(status string) bool {
	switch status {
	case ExecutionStatusSucceeded, ExecutionStatusFailed, ExecutionStatusCancelled:
		return true
	default:
		return false
	}
}
