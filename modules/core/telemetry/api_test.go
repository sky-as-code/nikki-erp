package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const testIdempotencyId = "019a46c2-2a40-7b35-9e0f-8e2fd22b9b6c"

func installTestProvider(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(NewPropagator([]string{"tracecontext", "baggage"}, nil))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	return exporter
}

func TestStartDetached_NewRootLinkedToSourceKeepsIdempotencyId(t *testing.T) {
	exporter := installTestProvider(t)
	parentCtx, cancel := context.WithCancel(corectx.WithIdempotencyId(context.Background(), testIdempotencyId))
	parentCtx, parent := Tracer("t").Start(parentCtx, "request")

	detachedCtx, detached := StartDetached(parentCtx, "publish")
	cancel()
	parent.End()
	detached.End()

	assert.NoError(t, detachedCtx.Err())
	id, ok := corectx.IdempotencyIdOf(detachedCtx)
	assert.True(t, ok)
	assert.Equal(t, testIdempotencyId, id)

	spans := exporter.GetSpans()
	require.Len(t, spans, 2)
	detachedStub := spans[1]
	assert.NotEqual(t, parent.SpanContext().TraceID(), detachedStub.SpanContext.TraceID())
	assert.False(t, detachedStub.Parent.IsValid())
	require.Len(t, detachedStub.Links, 1)
	assert.Equal(t, parent.SpanContext().SpanID(), detachedStub.Links[0].SpanContext.SpanID())
}

func TestStartDetached_NoSourceSpanHasNoLink(t *testing.T) {
	exporter := installTestProvider(t)
	_, span := StartDetached(context.Background(), "cron")
	span.End()
	require.Len(t, exporter.GetSpans(), 1)
	assert.Empty(t, exporter.GetSpans()[0].Links)
}

func TestMetadata_RoundTrip(t *testing.T) {
	installTestProvider(t)
	ctx := corectx.WithIdempotencyId(context.Background(), testIdempotencyId)
	ctx, span := Tracer("t").Start(ctx, "producer")
	defer span.End()

	metadata := map[string]string{}
	InjectToMetadata(ctx, metadata)
	assert.Equal(t, testIdempotencyId, metadata[MetadataIdempotencyId])
	assert.NotEmpty(t, metadata["traceparent"])

	consumerCtx := ExtractFromMetadata(context.Background(), metadata)
	assert.Equal(t, span.SpanContext().TraceID(), trace.SpanContextFromContext(consumerCtx).TraceID())
	id, ok := corectx.IdempotencyIdOf(consumerCtx)
	assert.True(t, ok)
	assert.Equal(t, testIdempotencyId, id)
}

func TestExtractFromMetadata_DropsInvalidIdempotencyId(t *testing.T) {
	installTestProvider(t)
	ctx := ExtractFromMetadata(context.Background(), map[string]string{MetadataIdempotencyId: "not-a-uuid"})
	_, ok := corectx.IdempotencyIdOf(ctx)
	assert.False(t, ok)
	assert.False(t, trace.SpanContextFromContext(ctx).IsValid())
}

type capturedRequest struct {
	header http.Header
}

func newCaptureServer(t *testing.T, captured *[]capturedRequest, redirectTo string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		*captured = append(*captured, capturedRequest{header: request.Header.Clone()})
		if redirectTo != "" {
			http.Redirect(writer, request, redirectTo, http.StatusFound)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	return server
}

func doRequest(t *testing.T, transport http.RoundTripper, ctx context.Context, target string) {
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	require.NoError(t, err)
	request.Header.Set("traceparent", sampleTraceparent)
	request.Header.Set(HeaderIdempotencyId, "inherited")
	response, err := (&http.Client{Transport: transport}).Do(request)
	require.NoError(t, err)
	_ = response.Body.Close()
}

func TestTransport_InjectsOnlyForAllowlistedHosts(t *testing.T) {
	exporter := installTestProvider(t)
	captured := []capturedRequest{}
	server := newCaptureServer(t, &captured, "")
	ctx := corectx.WithIdempotencyId(context.Background(), testIdempotencyId)

	doRequest(t, NewTransport(nil, []string{"127.0.0.1"}), ctx, server.URL)
	doRequest(t, NewTransport(nil, []string{"other.host"}), ctx, server.URL)

	require.Len(t, captured, 2)
	allowed, denied := captured[0].header, captured[1].header
	assert.Equal(t, testIdempotencyId, allowed.Get(HeaderIdempotencyId))
	assert.NotEqual(t, sampleTraceparent, allowed.Get("traceparent"))
	assert.Contains(t, allowed.Get("traceparent"), exporter.GetSpans()[0].SpanContext.SpanID().String())
	assert.Empty(t, denied.Get("traceparent"))
	assert.Empty(t, denied.Get(HeaderIdempotencyId))
	assert.Len(t, exporter.GetSpans(), 2)
}

// AC-21: a redirect to a host outside the allowlist receives no propagation headers.
func TestTransport_RedirectReevaluatesAllowlist(t *testing.T) {
	installTestProvider(t)
	captured := []capturedRequest{}
	external := newCaptureServer(t, &captured, "")
	externalUrl, err := url.Parse(external.URL)
	require.NoError(t, err)
	externalUrl.Host = "localhost:" + externalUrl.Port()
	internal := newCaptureServer(t, &captured, externalUrl.String())
	ctx := corectx.WithIdempotencyId(context.Background(), testIdempotencyId)

	doRequest(t, NewTransport(nil, []string{"127.0.0.1"}), ctx, internal.URL)

	require.Len(t, captured, 2)
	assert.NotEmpty(t, captured[0].header.Get("traceparent"))
	assert.Equal(t, testIdempotencyId, captured[0].header.Get(HeaderIdempotencyId))
	assert.Empty(t, captured[1].header.Get("traceparent"))
	assert.Empty(t, captured[1].header.Get(HeaderIdempotencyId))
}

func TestHostMatcher(t *testing.T) {
	matcher := newHostMatcher([]string{"api.internal", "svc.local:8443", "*.corp.example"})
	cases := map[string]bool{
		"http://api.internal:9000/x":     true,
		"https://svc.local:8443/":        true,
		"https://svc.local:9443/":        false,
		"https://a.corp.example/":        true,
		"https://corp.example/":          false,
		"https://evil-api.internal.com/": false,
	}
	for raw, want := range cases {
		target, err := url.Parse(raw)
		require.NoError(t, err)
		assert.Equal(t, want, matcher.matches(target), raw)
	}
}
