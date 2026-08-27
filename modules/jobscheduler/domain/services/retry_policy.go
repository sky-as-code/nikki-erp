package services

import (
	"time"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// maxBackoffShift bounds the doubling before it is capped anyway.
//
// The shift is what makes the delay grow, and 1<<63 overflows int64 nanoseconds into a negative
// duration - which would schedule a retry in the past and spin the loop. Stopping at 32 is far
// beyond any attempt count a real job reaches, and the result is capped by the configured maximum
// immediately afterwards, so the clamp is never the value anyone observes.
const maxBackoffShift = 32

// RetryDecision is what happens after an attempt ends without success.
type RetryDecision struct {
	// ShouldRetry is true only when a new attempt may actually run. When it is false the
	// execution is terminal and NO attempt record is created: a cancelled retry candidate is not
	// an attempt and must never appear in attempt history, or the history would show runs that
	// never happened.
	ShouldRetry bool

	// NextRetryAt is meaningful only when ShouldRetry is true.
	NextRetryAt time.Time

	// FailureCode is set only when ShouldRetry is false, and names why in machine-readable form.
	FailureCode string
}

// RetryInput is everything the decision depends on, passed by value.
//
// Nothing here is read from a database, a clock or a configuration service. That is deliberate:
// the retry rule is the most important logic in this module, and keeping it a pure function of
// its inputs is what lets every case be an ordinary table test rather than an integration setup.
type RetryInput struct {
	// AttemptNumber is the attempt that just failed, counting from 1.
	AttemptNumber int

	// MaxAttempts and RetryIntervalSecs come from the execution's snapshot, with the configured
	// defaults already resolved, so a configuration change cannot alter a retry chain that has
	// already begun.
	MaxAttempts       int
	RetryIntervalSecs int

	// MaxIntervalSecs comes from runtime configuration rather than the snapshot. It is a system
	// safety limit rather than per-job behaviour: an operator lowering it to shed load needs it
	// to take effect on work already queued.
	MaxIntervalSecs int

	// FinishedAt is when the failed attempt ended. The next retry is measured from here rather
	// than from when it started, so a slow attempt does not get a shorter gap before the next.
	FinishedAt time.Time

	// NextOccurrenceAt is the following occurrence of this execution's schedule, and the ceiling
	// on its retries. Nil means the schedule produces nothing further and there is therefore no
	// ceiling at all - which is not the same as "no retry".
	NextOccurrenceAt *time.Time

	// Retryable reports whether the failure is the kind repeating could fix.
	Retryable bool
}

// BackoffDelay is the delay before attempt number `attemptNumber`.
//
//	candidate = retryInterval * 2^(k-2)   for k >= 2
//	delay     = min(candidate, maxInterval)
//
// Attempt 1 is the first run and has no delay before it.
//
// The doubling uses a shift rather than math.Pow because the result must be exact: a float
// intermediate would introduce rounding into a value that decides whether a retry lands before or
// after the next occurrence, and being a nanosecond out at that boundary flips the decision.
func BackoffDelay(attemptNumber int, retryIntervalSecs int, maxIntervalSecs int) time.Duration {
	if attemptNumber < 2 {
		return 0
	}
	if retryIntervalSecs < 0 {
		retryIntervalSecs = 0
	}

	shift := attemptNumber - 2
	if shift > maxBackoffShift {
		shift = maxBackoffShift
	}

	candidateSecs := retryIntervalSecs << uint(shift)
	// The shift can still overflow for a large interval and a moderate attempt count; a negative
	// result means it wrapped, and the cap is the right answer in that case too.
	if candidateSecs < 0 || (maxIntervalSecs >= 0 && candidateSecs > maxIntervalSecs) {
		candidateSecs = maxIntervalSecs
	}
	if candidateSecs < 0 {
		candidateSecs = 0
	}
	return time.Duration(candidateSecs) * time.Second
}

// DecideRetry applies the retry policy after an attempt has failed or been abandoned.
//
// The order of the checks is itself the business rule and must not be rearranged:
//
//  1. A non-retryable failure ends the execution immediately, whatever quota remains. There is no
//     point computing a delay for a retry that will fail identically, and spending the quota only
//     delays the failure an operator needs to see.
//
//  2. An exhausted quota ends it. AttemptNumber counts the first run, so max_attempts of 3 gives
//     attempts 1, 2 and 3, and the failure of 3 is terminal.
//
//  3. The retry window. The candidate is permitted only when
//
//     candidate < NextOccurrenceAt
//
//     strictly less than. A candidate landing exactly on the next occurrence is cancelled: that
//     instant already belongs to the next occurrence's first attempt, and two runs must not share
//     it. When the window blocks the retry the execution ends as failed with RETRY_WINDOW_EXPIRED,
//     and the remaining quota is forfeited for this occurrence - a retry that cannot run before
//     the next occurrence will not become runnable by waiting longer.
//
//     A nil NextOccurrenceAt means no ceiling, and the retry proceeds under the other limits only.
//     Treating nil as the zero time would cancel every retry on a job that has just reached the
//     end of its effective period, silently.
//
// There is no exception to check 3. Exponential backoff, worker crash recovery, REST timeouts,
// command failures and rate limiting all arrive here and all pass through it.
func DecideRetry(in RetryInput) RetryDecision {
	if !in.Retryable {
		return RetryDecision{FailureCode: models.FailureNonRetryable}
	}

	if in.AttemptNumber >= in.MaxAttempts {
		return RetryDecision{FailureCode: models.FailureMaxAttemptsReached}
	}

	delay := BackoffDelay(in.AttemptNumber+1, in.RetryIntervalSecs, in.MaxIntervalSecs)
	candidate := in.FinishedAt.UTC().Add(delay)

	if in.NextOccurrenceAt != nil && !candidate.Before(in.NextOccurrenceAt.UTC()) {
		return RetryDecision{FailureCode: models.FailureRetryWindowExpired}
	}

	return RetryDecision{ShouldRetry: true, NextRetryAt: candidate}
}
