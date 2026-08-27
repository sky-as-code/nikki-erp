package external

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"
	it "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/external"
)

func newTestExecutor() *RestApiExecutor {
	return NewRestApiExecutor(&client.HttpClient{Client: http.Client{}})
}

func testInput(config map[string]any) it.ActionInput {
	return it.ActionInput{
		Config:        config,
		ExecutionKey:  "inventory:rebuild:2026-08-20T10:00:00Z",
		JobId:         "01M2JBJ0000000001000000000",
		ExecutionId:   "01M2JBE0000000001000000000",
		AttemptNumber: 2,
		Timeout:       5 * time.Second,
	}
}

func TestValidateAcceptsEverySupportedMethod(t *testing.T) {
	executor := newTestExecutor()

	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"} {
		errs := executor.Validate(nil, map[string]any{
			"method": method, "url": "https://internal.example/run",
		})
		assert.Nil(t, errs, "method %s should be accepted", method)
	}
}

// Compared exactly rather than case-insensitively: accepting "post" and sending "POST" would mean
// the stored job does not say what it actually does.
func TestValidateRejectsUnsupportedAndLowercaseMethods(t *testing.T) {
	executor := newTestExecutor()

	for _, method := range []string{"TRACE", "OPTIONS", "CONNECT", "post", "Get", ""} {
		errs := executor.Validate(nil, map[string]any{
			"method": method, "url": "https://internal.example/run",
		})
		require.NotNil(t, errs, "method %q should be rejected", method)
		assert.True(t, errs.Has("action_config.method"))
	}
}

// The URL arrives from an API caller, so restricting the scheme is a security boundary: file:// or
// gopher:// would turn an authenticated job registration into a way to read local files or reach
// services never meant to be reachable.
func TestValidateRejectsNonHttpSchemes(t *testing.T) {
	executor := newTestExecutor()

	for _, rawUrl := range []string{
		"file:///etc/passwd",
		"gopher://internal.example/1",
		"ftp://internal.example/x",
		"/relative/path",
		"internal.example/no-scheme",
		"https://",
		"",
	} {
		errs := executor.Validate(nil, map[string]any{"method": "POST", "url": rawUrl})
		require.NotNil(t, errs, "url %q should be rejected", rawUrl)
		assert.True(t, errs.Has("action_config.url"))
	}
}

func TestExecuteSendsTheSchedulerIdentityHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	outcome := newTestExecutor().Execute(context.Background(), testInput(map[string]any{
		"method": "POST", "url": server.URL,
	}))

	require.True(t, outcome.Succeeded)
	assert.Equal(t, "inventory:rebuild:2026-08-20T10:00:00Z", received.Get(HeaderIdempotencyKey))
	assert.Equal(t, "01M2JBJ0000000001000000000", received.Get(HeaderJobId))
	assert.Equal(t, "01M2JBE0000000001000000000", received.Get(HeaderExecutionId))
	assert.Equal(t, "2", received.Get(HeaderAttempt))
}

// Every attempt of the same execution presents the same idempotency key, because a retry must be
// recognizable to the receiver as the same work rather than as new work.
func TestRetriesOfOneExecutionShareTheirIdempotencyKey(t *testing.T) {
	keys := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get(HeaderIdempotencyKey))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor := newTestExecutor()
	for attempt := 1; attempt <= 3; attempt++ {
		in := testInput(map[string]any{"method": "POST", "url": server.URL})
		in.AttemptNumber = attempt
		executor.Execute(context.Background(), in)
	}

	require.Len(t, keys, 3)
	assert.Equal(t, keys[0], keys[1])
	assert.Equal(t, keys[1], keys[2])
}

// The scheduler's headers are set last, so an action config cannot rewrite them. If it could, two
// different occurrences could be made to look like the same work to the receiver, defeating the
// idempotency the headers exist to provide.
func TestActionConfigCannotOverrideSchedulerHeaders(t *testing.T) {
	var received http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	newTestExecutor().Execute(context.Background(), testInput(map[string]any{
		"method": "POST", "url": server.URL,
		"headers": map[string]any{
			HeaderIdempotencyKey: "forged",
			HeaderAttempt:        "999",
			"X-Custom":           "kept",
		},
	}))

	assert.Equal(t, "inventory:rebuild:2026-08-20T10:00:00Z", received.Get(HeaderIdempotencyKey))
	assert.Equal(t, "2", received.Get(HeaderAttempt))
	assert.Equal(t, "kept", received.Get("X-Custom"), "unrelated headers still pass through")
}

func TestExecuteClassifiesRemoteFailures(t *testing.T) {
	tests := []struct {
		status        int
		wantRetryable bool
	}{
		{http.StatusOK, false},
		{http.StatusServiceUnavailable, true},
		{http.StatusTooManyRequests, true},
		{http.StatusRequestTimeout, true},
		{http.StatusUnauthorized, false},
		{http.StatusUnprocessableEntity, false},
	}

	for _, tc := range tests {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
		}))

		outcome := newTestExecutor().Execute(context.Background(), testInput(map[string]any{
			"method": "GET", "url": server.URL,
		}))

		if tc.status == http.StatusOK {
			assert.True(t, outcome.Succeeded)
		} else {
			assert.False(t, outcome.Succeeded, "status %d", tc.status)
			assert.Equal(t, tc.wantRetryable, outcome.Retryable, "status %d", tc.status)
		}
		server.Close()
	}
}

// The per-attempt timeout comes from configuration and must bound the call, rather than the
// client's own much longer ceiling. Without it an attempt could outlive the lease protecting it,
// and another instance would start the same work while this one was still running.
func TestExecuteIsBoundedByThePerAttemptTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	in := testInput(map[string]any{"method": "GET", "url": server.URL})
	in.Timeout = 150 * time.Millisecond

	started := time.Now()
	outcome := newTestExecutor().Execute(context.Background(), in)
	elapsed := time.Since(started)

	assert.False(t, outcome.Succeeded)
	assert.Equal(t, ErrorCodeTimeout, outcome.ErrorCode)
	assert.Less(t, elapsed, 2*time.Second, "the attempt must not outlive its timeout")
}

// A row can outlive a validation change, so a bad config fails the attempt rather than panicking
// a worker goroutine - and it is not retried, because it will be exactly as bad next time.
func TestExecuteRefusesAnUnusableConfigWithoutRetrying(t *testing.T) {
	executor := newTestExecutor()

	for _, config := range []map[string]any{
		{"method": "TRACE", "url": "https://internal.example"},
		{"method": "GET", "url": "file:///etc/passwd"},
		{"method": "GET"},
		{},
	} {
		outcome := executor.Execute(context.Background(), testInput(config))

		assert.False(t, outcome.Succeeded)
		assert.False(t, outcome.Retryable, "a permanently bad config must not be retried")
	}
}

func TestExecuteSendsAConfiguredBody(t *testing.T) {
	var body []byte
	var contentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body = make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	outcome := newTestExecutor().Execute(context.Background(), testInput(map[string]any{
		"method": "POST", "url": server.URL, "body": `{"full":true}`,
	}))

	require.True(t, outcome.Succeeded)
	assert.Equal(t, `{"full":true}`, string(body))
	assert.Equal(t, "application/json", contentType)
}
