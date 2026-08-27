package external

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessfulStatusesSucceed(t *testing.T) {
	for _, status := range []int{200, 201, 202, 204, 299} {
		outcome := ClassifyRestResult(status, nil)

		assert.True(t, outcome.Succeeded, "status %d", status)
		assert.Empty(t, outcome.ErrorCode)
		require.NotNil(t, outcome.HttpStatusCode)
		assert.Equal(t, int32(status), *outcome.HttpStatusCode)
	}
}

func TestRetryableStatusesAreRetried(t *testing.T) {
	// 408 and 429 are the two 4xx that mean "not now" rather than "not ever", which is why they
	// are singled out from the rest of the 4xx range.
	for _, status := range []int{408, 429, 500, 502, 503, 504, 599} {
		outcome := ClassifyRestResult(status, nil)

		assert.False(t, outcome.Succeeded, "status %d", status)
		assert.True(t, outcome.Retryable, "status %d must be retryable", status)
	}
}

// A 4xx will be rejected identically next time, so retrying spends the attempt budget to no
// purpose and delays the failure the operator needs to see.
func TestNonRetryableClientErrorsEndTheExecution(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 409, 422} {
		outcome := ClassifyRestResult(status, nil)

		assert.False(t, outcome.Succeeded, "status %d", status)
		assert.False(t, outcome.Retryable, "status %d must not be retried", status)
		assert.Equal(t, "HTTP_"+itoa(status), outcome.ErrorCode)
	}
}

// The single easiest thing to get wrong here. The HTTP client returns BOTH a non-nil response and
// a non-nil error for a non-2xx status. A classifier that tested the error first would treat every
// 404 as a transport failure and retry it forever, quietly converting a permanent misconfiguration
// into an infinite retry loop.
func TestStatusWinsOverAnAccompanyingError(t *testing.T) {
	outcome := ClassifyRestResult(404, errors.New("404 Not Found"))

	assert.False(t, outcome.Retryable,
		"a status must be classified as a status, even when an error accompanies it")
	assert.Equal(t, "HTTP_404", outcome.ErrorCode)
}

func TestTimeoutIsRetryable(t *testing.T) {
	outcome := ClassifyRestResult(0, context.DeadlineExceeded)

	assert.False(t, outcome.Succeeded)
	assert.True(t, outcome.Retryable)
	assert.Equal(t, ErrorCodeTimeout, outcome.ErrorCode)
	assert.Nil(t, outcome.HttpStatusCode, "there was no response to take a status from")
}

// A cancellation is this process stopping, not the job failing. The attempt is left for the lease
// reaper so a rolling deploy does not spend the retry budget of every job that happened to be
// running at the time.
func TestCancellationIsDistinguishedFromTimeout(t *testing.T) {
	outcome := ClassifyRestResult(0, context.Canceled)

	assert.Equal(t, ErrorCodeCancelled, outcome.ErrorCode)
	assert.NotEqual(t, ErrorCodeTimeout, outcome.ErrorCode,
		"a shutdown must not be recorded as the job timing out")
	assert.True(t, IsShutdownOutcome(outcome))
}

func TestTimeoutIsNotTreatedAsShutdown(t *testing.T) {
	assert.False(t, IsShutdownOutcome(ClassifyRestResult(0, context.DeadlineExceeded)))
	assert.False(t, IsShutdownOutcome(ClassifyRestResult(503, nil)))
}

// A wrapped error must still be recognized: the HTTP stack wraps transport errors on the way out.
func TestWrappedTimeoutIsStillATimeout(t *testing.T) {
	wrapped := errors.Join(errors.New("Get \"https://example\": "), context.DeadlineExceeded)

	outcome := ClassifyRestResult(0, wrapped)

	assert.Equal(t, ErrorCodeTimeout, outcome.ErrorCode)
}

func TestNetworkFailureIsRetryable(t *testing.T) {
	outcome := ClassifyRestResult(0, &net.OpError{Op: "dial", Err: errors.New("connection refused")})

	assert.True(t, outcome.Retryable)
	assert.Equal(t, ErrorCodeNetwork, outcome.ErrorCode)
}

// Neither a status nor an error should not happen, but silently succeeding would be the one
// reading that loses information.
func TestAbsentStatusAndErrorIsRecordedRatherThanAssumedSuccessful(t *testing.T) {
	outcome := ClassifyRestResult(0, nil)

	assert.False(t, outcome.Succeeded)
	assert.True(t, outcome.Retryable)
}

// The message reaches the attempt history, where a credential would be persisted and shown.
func TestErrorMessagesCarryNoRequestDetail(t *testing.T) {
	secret := "https://internal.example/run?token=s3cr3t"
	outcome := ClassifyRestResult(0, errors.New("Get \""+secret+"\": dial tcp: refused"))

	assert.NotContains(t, outcome.ErrorMessage, "s3cr3t")
	assert.NotContains(t, outcome.ErrorMessage, "token")
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return digits
}
