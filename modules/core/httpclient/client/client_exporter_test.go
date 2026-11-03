package client_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/httpclienttest"
)

func clientSpans(spans []sdktrace.ReadOnlySpan) []sdktrace.ReadOnlySpan {
	out := []sdktrace.ReadOnlySpan{}
	for _, span := range spans {
		if span.SpanKind() == trace.SpanKindClient {
			out = append(out, span)
		}
	}
	return out
}

func installProvider(t *testing.T, options ...sdktrace.TracerProviderOption) *sdktrace.TracerProvider {
	httpclienttest.InstallTelemetry(t)
	provider := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		_ = provider.Shutdown(ctx)
	})
	return provider
}

func newExporter(t *testing.T, endpoint string) sdktrace.SpanExporter {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpointURL(endpoint+"/v1/traces"),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{Enabled: false}),
		otlptracehttp.WithTimeout(500*time.Millisecond),
	)
	require.NoError(t, err)
	return exporter
}

// AC-22: the exporter's own requests go out on its own client. They make no span (no export
// loop) and carry no trace or idempotency header, even with the collector host allowlisted.
func TestExporterRequestsAreNotInstrumented(t *testing.T) {
	collector := httpclienttest.NewHeaderServer(t, nil)
	recorder := tracetest.NewSpanRecorder()
	provider := installProvider(t,
		sdktrace.WithSpanProcessor(recorder),
		sdktrace.WithSyncer(newExporter(t, collector.URL)),
	)
	backend := httpclienttest.NewHeaderServer(t, nil)
	shared := httpclienttest.NewClient(t, nil, backend.HostPort(t), collector.HostPort(t))
	ctx, _ := httpclienttest.TracedContext(t)

	get(t, shared, ctx, backend.URL)
	require.NoError(t, provider.ForceFlush(context.Background()))

	collector.AssertNoTelemetryHeaders(t)
	assert.Len(t, clientSpans(recorder.Ended()), 1, "only the backend call is a CLIENT span")
}

// AC-18 at unit level: a receiver that never answers fills the bounded queue, which then drops;
// outbound calls are not held by it.
func TestStalledReceiverDoesNotBlockCalls(t *testing.T) {
	release := make(chan struct{})
	stalled := httpclienttest.NewHeaderServer(t, func(_ http.ResponseWriter, request *http.Request) {
		select {
		case <-release:
		case <-request.Context().Done():
		}
	})
	defer close(release)
	installProvider(t, sdktrace.WithBatcher(newExporter(t, stalled.URL),
		sdktrace.WithMaxQueueSize(4), sdktrace.WithMaxExportBatchSize(2), sdktrace.WithBatchTimeout(time.Millisecond)))
	backend := httpclienttest.NewHeaderServer(t, nil)
	shared := httpclienttest.NewClient(t, nil, backend.HostPort(t))
	ctx, _ := httpclienttest.TracedContext(t)

	started := time.Now()
	for i := 0; i < 40; i++ {
		get(t, shared, ctx, backend.URL)
	}

	assert.Less(t, time.Since(started), 3*time.Second)
	assert.Len(t, backend.Received(), 40)
}
