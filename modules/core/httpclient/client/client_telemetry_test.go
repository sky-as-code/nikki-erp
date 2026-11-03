package client_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/httpclienttest"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

type syncBuffer struct {
	mutex  sync.Mutex
	buffer bytes.Buffer
}

func (this *syncBuffer) Write(data []byte) (int, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.buffer.Write(data)
}

func (this *syncBuffer) String() string {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.buffer.String()
}

func get(t *testing.T, shared *client.HttpClient, ctx context.Context, target string) *http.Response {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	require.NoError(t, err)
	response, err := shared.Do(request)
	require.NoError(t, err)
	_ = response.Body.Close()
	return response
}

// AC-21: the first hop is allowlisted and gets the headers; the redirect target is not and gets
// none of them, although net/http copies the original request's headers onto the redirect.
func TestRedirectToNonAllowlistedHostCarriesNoTelemetry(t *testing.T) {
	httpclienttest.InstallTelemetry(t)
	target := httpclienttest.NewHeaderServer(t, nil)
	origin := httpclienttest.NewHeaderServer(t, func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL+"/landing", http.StatusFound)
	})
	shared := httpclienttest.NewClient(t, nil, origin.HostPort(t))
	ctx, id := httpclienttest.TracedContext(t)

	get(t, shared, ctx, origin.URL+"/start")

	first := origin.Received()
	require.Len(t, first, 1)
	assert.NotEmpty(t, first[0].Get("traceparent"))
	assert.Contains(t, first[0].Get("baggage"), httpclienttest.BaggageKey)
	assert.Equal(t, id, first[0].Get(telemetry.HeaderIdempotencyId))
	target.AssertNoTelemetryHeaders(t)
}

// Caller-set copies of the headers are stripped too, so a hand-built request cannot leak them.
func TestHandSetTelemetryHeadersAreStrippedForNonAllowlistedHost(t *testing.T) {
	httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, nil)
	shared := httpclienttest.NewClient(t, nil)
	ctx, id := httpclienttest.TracedContext(t)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	request.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	request.Header.Set(telemetry.HeaderIdempotencyId, id)
	response, err := shared.Do(request)
	require.NoError(t, err)
	_ = response.Body.Close()

	server.AssertNoTelemetryHeaders(t)
}

// ID-AC-10: concurrent calls for one request share its idempotency id and trace, each in its own
// CLIENT span.
func TestFanOutCallsShareIdempotencyIdAndTrace(t *testing.T) {
	recorder := httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, nil)
	shared := httpclienttest.NewClient(t, nil, server.HostPort(t))
	ctx, id := httpclienttest.TracedContext(t)
	const calls = 4

	var group sync.WaitGroup
	for i := 0; i < calls; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			get(t, shared, ctx, server.URL)
		}()
	}
	group.Wait()

	parents := map[string]bool{}
	traceId := trace.SpanContextFromContext(ctx).TraceID().String()
	for _, header := range server.Received() {
		assert.Equal(t, id, header.Get(telemetry.HeaderIdempotencyId))
		parts := strings.Split(header.Get("traceparent"), "-")
		require.Len(t, parts, 4)
		assert.Equal(t, traceId, parts[1])
		parents[parts[2]] = true
	}
	assert.Len(t, parents, calls, "each call must be its own CLIENT span")
	assert.Len(t, clientSpans(recorder.Ended()), calls)
}

// ID-AC-17: a downstream answering with another id is reported; ours is neither replaced in ctx
// nor on the next call.
func TestDownstreamIdMismatchIsLoggedAndOursKept(t *testing.T) {
	httpclienttest.InstallTelemetry(t)
	foreign, err := corectx.NewIdempotencyId()
	require.NoError(t, err)
	server := httpclienttest.NewHeaderServer(t, func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set(telemetry.HeaderIdempotencyId, foreign)
	})
	logs := &syncBuffer{}
	shared := httpclienttest.NewClient(t, logging.NewLoggerTo(logs), server.HostPort(t))
	ctx, id := httpclienttest.TracedContext(t)

	get(t, shared, ctx, server.URL)
	get(t, shared, ctx, server.URL)

	assert.Contains(t, logs.String(), "downstream answered with a different idempotency id")
	assert.Contains(t, logs.String(), `"idempotency_id":"`+id+`"`)
	assert.Contains(t, logs.String(), foreign)
	current, _ := corectx.IdempotencyIdOf(ctx)
	assert.Equal(t, id, current)
	for _, header := range server.Received() {
		assert.Equal(t, id, header.Get(telemetry.HeaderIdempotencyId))
	}
}

// The client never mints an id: a context without one sends none and says so in the log.
func TestMissingIdempotencyIdIsNotMinted(t *testing.T) {
	httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, nil)
	logs := &syncBuffer{}
	shared := httpclienttest.NewClient(t, logging.NewLoggerTo(logs), server.HostPort(t))

	get(t, shared, context.Background(), server.URL)

	assert.Empty(t, server.Received()[0].Get(telemetry.HeaderIdempotencyId))
	assert.NotEmpty(t, server.Received()[0].Get("traceparent"))
	assert.Contains(t, logs.String(), "outbound http call has no idempotency id in context")
}

// A host off the allowlist still gets a CLIENT span; it just receives no header.
func TestNonAllowlistedHostIsTracedWithoutPropagation(t *testing.T) {
	recorder := httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, nil)
	shared := httpclienttest.NewClient(t, nil, "internal.example")
	ctx, _ := httpclienttest.TracedContext(t)

	get(t, shared, ctx, server.URL)

	server.AssertNoTelemetryHeaders(t)
	spans := clientSpans(recorder.Ended())
	require.Len(t, spans, 1)
	assert.Equal(t, trace.SpanContextFromContext(ctx).TraceID(), spans[0].SpanContext().TraceID())
}

// WithTimeout keeps the instrumented transport, so a derived client still strips and traces.
func TestWithTimeoutKeepsInstrumentedTransport(t *testing.T) {
	recorder := httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, nil)
	shared := httpclienttest.NewClient(t, nil)
	derived := shared.WithTimeout(1500 * time.Millisecond)
	ctx, _ := httpclienttest.TracedContext(t)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	response, err := derived.Do(request)
	require.NoError(t, err)
	_ = response.Body.Close()

	assert.Equal(t, shared.Transport, derived.Transport)
	assert.NotEqual(t, shared.Timeout, derived.Timeout)
	server.AssertNoTelemetryHeaders(t)
	assert.Len(t, clientSpans(recorder.Ended()), 1)
}
