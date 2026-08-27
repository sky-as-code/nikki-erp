package external

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/external"
)

// Error codes an attempt can record. They are stable strings rather than an enum because they end
// up in the database and in operator dashboards, where a renumbered constant would silently break
// every saved filter.
const (
	ErrorCodeTimeout   = "TIMEOUT"
	ErrorCodeCancelled = "CANCELLED"
	ErrorCodeNetwork   = "NETWORK"
	ErrorCodeCommand   = "COMMAND_ERROR"
)

// ClassifyRestResult turns a transport error and a status code into a retry verdict.
//
//	2xx                        -> success
//	network failure, timeout   -> retryable
//	408, 429, 5xx              -> retryable
//	every other 4xx            -> NOT retryable; the execution ends immediately
//
// A non-retryable 4xx is terminal because the request will be rejected identically next time.
// Retrying a 401 or a 422 spends the attempt budget to no purpose and delays the failure the
// operator actually needs to see - and for a job on a short schedule, it can burn the whole retry
// window on an answer that was never going to change.
//
// The order of the checks matters and is the easiest thing to get wrong here. The HTTP client
// returns BOTH a non-nil response and a non-nil error for a non-2xx status, so a classifier that
// tested the error first would see every 404 as a transport failure and retry it forever. The
// status is therefore consulted whenever there is one, and the error only when there is not.
func ClassifyRestResult(statusCode int, err error) it.ActionOutcome {
	if statusCode > 0 {
		return classifyStatus(statusCode)
	}
	return classifyTransportError(err)
}

func classifyStatus(statusCode int) it.ActionOutcome {
	code := int32(statusCode)

	if statusCode >= 200 && statusCode < 300 {
		return it.ActionOutcome{Succeeded: true, HttpStatusCode: &code}
	}

	retryable := statusCode >= 500 ||
		statusCode == http.StatusRequestTimeout ||
		statusCode == http.StatusTooManyRequests

	return it.ActionOutcome{
		Retryable:      retryable,
		ErrorCode:      "HTTP_" + strconv.Itoa(statusCode),
		ErrorMessage:   http.StatusText(statusCode),
		HttpStatusCode: &code,
	}
}

// classifyTransportError distinguishes a timeout from a shutdown, which look alike but must be
// handled differently.
//
// A deadline exceeded is the attempt genuinely taking too long: it failed, and it is worth
// retrying. A cancellation means this process is stopping mid-attempt; the work may well still be
// running on the other end, and the lease reaper will recover the attempt on whichever instance
// picks it up. Counting a shutdown against the job's retry budget would let a rolling deploy
// exhaust the attempts of every job that happened to be running.
func classifyTransportError(err error) it.ActionOutcome {
	switch {
	case err == nil:
		// No status and no error should not happen, but treating it as a silent success would be
		// the one interpretation that loses information: the attempt is recorded as failed and
		// retryable so the next run reveals what is going on.
		return it.ActionOutcome{
			Retryable:    true,
			ErrorCode:    ErrorCodeNetwork,
			ErrorMessage: "the action returned neither a response nor an error",
		}

	case errors.Is(err, context.DeadlineExceeded):
		return it.ActionOutcome{
			Retryable:    true,
			ErrorCode:    ErrorCodeTimeout,
			ErrorMessage: "the attempt exceeded its timeout",
		}

	case errors.Is(err, context.Canceled):
		return it.ActionOutcome{
			Retryable:    true,
			ErrorCode:    ErrorCodeCancelled,
			ErrorMessage: "the attempt was cancelled while the scheduler was stopping",
		}

	default:
		return it.ActionOutcome{
			Retryable:    true,
			ErrorCode:    ErrorCodeNetwork,
			ErrorMessage: "the action could not reach its target",
		}
	}
}

// IsShutdownOutcome reports whether an outcome represents this process stopping rather than the
// action failing. The caller leaves such an attempt running so the lease reaper recovers it,
// instead of consuming a retry for something that was never the job's fault.
func IsShutdownOutcome(outcome it.ActionOutcome) bool {
	return outcome.ErrorCode == ErrorCodeCancelled
}
