package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

func at(hour, minute, second int) time.Time {
	return time.Date(2026, time.August, 20, hour, minute, second, 0, time.UTC)
}

func ptr(t time.Time) *time.Time { return &t }

// AC-18: the delay doubles and is then capped by the configured maximum.
func TestBackoffDoublesAndIsCappedByConfiguration(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 0},                  // the first run has no delay before it
		{2, 10 * time.Second},   // base
		{3, 20 * time.Second},   // base x 2
		{4, 40 * time.Second},   // base x 4
		{5, 80 * time.Second},   // base x 8
		{6, 160 * time.Second},  // base x 16
		{7, 300 * time.Second},  // would be 320; capped
		{10, 300 * time.Second}, // still capped
		{40, 300 * time.Second}, // far past where a naive shift would overflow
	}

	for _, tc := range tests {
		assert.Equal(t, tc.want, BackoffDelay(tc.attempt, 10, 300),
			"attempt %d", tc.attempt)
	}
}

// A delay that wrapped negative would schedule a retry in the past and spin the loop, so the
// arithmetic has to saturate rather than overflow.
func TestBackoffNeverProducesANegativeDelay(t *testing.T) {
	for _, attempt := range []int{2, 32, 64, 1000} {
		assert.GreaterOrEqual(t, BackoffDelay(attempt, 86400, 300), time.Duration(0),
			"attempt %d", attempt)
	}
}

// AC-13: a candidate landing before the next occurrence is allowed.
func TestRetryIsAllowedWhenItLandsBeforeTheNextOccurrence(t *testing.T) {
	decision := DecideRetry(RetryInput{
		AttemptNumber: 1, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
		FinishedAt: at(10, 0, 5), NextOccurrenceAt: ptr(at(10, 1, 0)), Retryable: true,
	})

	require.True(t, decision.ShouldRetry)
	assert.Equal(t, at(10, 0, 15), decision.NextRetryAt)
	assert.Empty(t, decision.FailureCode)
}

// AC-14: a candidate landing exactly on the next occurrence is cancelled. That instant already
// belongs to the next occurrence's first attempt, and two runs must not share it.
func TestRetryIsCancelledWhenItLandsExactlyOnTheNextOccurrence(t *testing.T) {
	decision := DecideRetry(RetryInput{
		AttemptNumber: 1, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
		FinishedAt: at(10, 0, 50), NextOccurrenceAt: ptr(at(10, 1, 0)), Retryable: true,
	})

	assert.False(t, decision.ShouldRetry, "the comparison is strictly less-than")
	assert.Equal(t, models.FailureRetryWindowExpired, decision.FailureCode)
}

// The boundary is exact to the nanosecond, because that is what decides the case above.
func TestRetryIsAllowedOneNanosecondBeforeTheCeiling(t *testing.T) {
	finished := at(10, 1, 0).Add(-10*time.Second - time.Nanosecond)

	decision := DecideRetry(RetryInput{
		AttemptNumber: 1, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
		FinishedAt: finished, NextOccurrenceAt: ptr(at(10, 1, 0)), Retryable: true,
	})

	require.True(t, decision.ShouldRetry)
	assert.True(t, decision.NextRetryAt.Before(at(10, 1, 0)))
}

// AC-15, replaying the worked example in the requirement verbatim: a job on */1, attempt 1 fails
// at 10:00:55, base interval 10s, so the candidate is 10:01:05 while the next occurrence is
// 10:01:00. The retry is cancelled and the occurrence ends.
func TestRequirementExampleRetryOverrunsTheNextMinute(t *testing.T) {
	decision := DecideRetry(RetryInput{
		AttemptNumber: 1, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
		FinishedAt: at(10, 0, 55), NextOccurrenceAt: ptr(at(10, 1, 0)), Retryable: true,
	})

	assert.False(t, decision.ShouldRetry)
	assert.Equal(t, models.FailureRetryWindowExpired, decision.FailureCode)
}

// AC-19, the second worked example: a job on */5, attempt 3 fails at 10:03:30 with a 60s base, so
// the next delay is 240s and the candidate is 10:07:30 while the next occurrence is 10:05:00.
//
// The point of this case is that 240 is well UNDER the 300s cap: the backoff maximum is not what
// stops it. Exponential backoff must not be allowed to let an old execution bleed into the next
// occurrence, whatever the cap says.
func TestRequirementExampleBackoffBlockedBelowTheCap(t *testing.T) {
	assert.Equal(t, 240*time.Second, BackoffDelay(4, 60, 300),
		"precondition: the delay is under the cap, so the cap is not what cancels the retry")

	decision := DecideRetry(RetryInput{
		AttemptNumber: 3, MaxAttempts: 5, RetryIntervalSecs: 60, MaxIntervalSecs: 300,
		FinishedAt: at(10, 3, 30), NextOccurrenceAt: ptr(at(10, 5, 0)), Retryable: true,
	})

	assert.False(t, decision.ShouldRetry)
	assert.Equal(t, models.FailureRetryWindowExpired, decision.FailureCode)
}

// With no further occurrence there is no ceiling, and the retry proceeds under the other limits
// only. Treating nil as the zero time would cancel every retry on a job that has just passed its
// effective period, silently.
func TestNilCeilingMeansNoLimitRatherThanNoRetry(t *testing.T) {
	decision := DecideRetry(RetryInput{
		AttemptNumber: 1, MaxAttempts: 3, RetryIntervalSecs: 3600, MaxIntervalSecs: 86400,
		FinishedAt: at(10, 0, 0), NextOccurrenceAt: nil, Retryable: true,
	})

	require.True(t, decision.ShouldRetry, "no further occurrence means no ceiling, not no retry")
	assert.Equal(t, at(11, 0, 0), decision.NextRetryAt)
}

// A non-retryable failure ends the execution immediately, whatever quota remains: repeating a 401
// or a 422 spends attempts to no purpose and delays the failure an operator needs to see.
func TestNonRetryableFailureEndsTheExecutionDespiteRemainingQuota(t *testing.T) {
	decision := DecideRetry(RetryInput{
		AttemptNumber: 1, MaxAttempts: 10, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
		FinishedAt: at(10, 0, 0), NextOccurrenceAt: ptr(at(23, 0, 0)), Retryable: false,
	})

	assert.False(t, decision.ShouldRetry)
	assert.Equal(t, models.FailureNonRetryable, decision.FailureCode)
}

// AttemptNumber counts the first run, so max_attempts of 3 yields attempts 1, 2 and 3, and the
// failure of 3 is terminal. An off-by-one here would silently grant every job an extra attempt.
func TestQuotaCountsTheFirstRun(t *testing.T) {
	ceiling := ptr(at(23, 0, 0))

	for attempt := 1; attempt <= 2; attempt++ {
		decision := DecideRetry(RetryInput{
			AttemptNumber: attempt, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
			FinishedAt: at(10, 0, 0), NextOccurrenceAt: ceiling, Retryable: true,
		})
		assert.True(t, decision.ShouldRetry, "attempt %d of 3 may still retry", attempt)
	}

	decision := DecideRetry(RetryInput{
		AttemptNumber: 3, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
		FinishedAt: at(10, 0, 0), NextOccurrenceAt: ceiling, Retryable: true,
	})
	assert.False(t, decision.ShouldRetry, "the failure of attempt 3 of 3 is terminal")
	assert.Equal(t, models.FailureMaxAttemptsReached, decision.FailureCode)
}

// AC-20: an attempt abandoned because its lease expired goes through the same window check as any
// other failure. Crash recovery gets no exemption - a recovered attempt that would run past the
// next occurrence is exactly as unwelcome as one that failed normally.
func TestCrashRecoveryObeysTheRetryWindow(t *testing.T) {
	reapedAt := at(10, 0, 58)

	decision := DecideRetry(RetryInput{
		AttemptNumber: 1, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
		FinishedAt: reapedAt, NextOccurrenceAt: ptr(at(10, 1, 0)), Retryable: true,
	})

	assert.False(t, decision.ShouldRetry)
	assert.Equal(t, models.FailureRetryWindowExpired, decision.FailureCode)
}

// The checks are ordered, and the order is the rule. A non-retryable failure is reported as such
// even when the quota is also spent and the window has also closed, because that is the finding
// that explains what actually happened.
func TestCheckOrderIsPreservedWhenSeveralLimitsApplyAtOnce(t *testing.T) {
	decision := DecideRetry(RetryInput{
		AttemptNumber: 9, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
		FinishedAt: at(10, 9, 0), NextOccurrenceAt: ptr(at(10, 1, 0)), Retryable: false,
	})

	assert.Equal(t, models.FailureNonRetryable, decision.FailureCode,
		"non-retryable is checked first and is the most specific explanation")
}

// AC-16 and AC-17: whenever a retry is refused the decision carries a machine-readable reason and
// no retry instant, which is what stops the caller writing a fake attempt row for a run that never
// happened.
func TestARefusedRetryCarriesAReasonAndNoInstant(t *testing.T) {
	refusals := []RetryInput{
		{AttemptNumber: 1, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
			FinishedAt: at(10, 0, 0), Retryable: false},
		{AttemptNumber: 3, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
			FinishedAt: at(10, 0, 0), NextOccurrenceAt: ptr(at(23, 0, 0)), Retryable: true},
		{AttemptNumber: 1, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
			FinishedAt: at(10, 0, 55), NextOccurrenceAt: ptr(at(10, 1, 0)), Retryable: true},
	}

	for _, in := range refusals {
		decision := DecideRetry(in)
		assert.False(t, decision.ShouldRetry)
		assert.NotEmpty(t, decision.FailureCode)
		assert.True(t, decision.NextRetryAt.IsZero(),
			"a refused retry has no instant, so nothing can schedule it anyway")
	}
}

// The decision is computed in UTC whatever zone the input carried, so a retry instant can be
// compared with a ceiling read from the database without either being converted first.
func TestDecisionIsComputedInUtc(t *testing.T) {
	jakarta := time.FixedZone("WIB", 7*60*60)
	finished := time.Date(2026, time.August, 20, 17, 0, 0, 0, jakarta)

	decision := DecideRetry(RetryInput{
		AttemptNumber: 1, MaxAttempts: 3, RetryIntervalSecs: 10, MaxIntervalSecs: 300,
		FinishedAt: finished, NextOccurrenceAt: ptr(at(11, 0, 0)), Retryable: true,
	})

	require.True(t, decision.ShouldRetry)
	assert.Equal(t, time.UTC, decision.NextRetryAt.Location())
	assert.Equal(t, at(10, 0, 10), decision.NextRetryAt)
}
