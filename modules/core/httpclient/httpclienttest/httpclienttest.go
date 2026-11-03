// Package httpclienttest holds test helpers for code that calls out through the shared HTTP
// client: an in-memory tracer, a traced request context, and a server that records headers.
// Import it from _test files only.
package httpclienttest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

// BaggageKey is the one baggage key InstallTelemetry allowlists, so a leak is observable.
const BaggageKey = "coremart.test"

// TelemetryHeaders are the headers only an allowlisted host may receive.
var TelemetryHeaders = []string{"traceparent", "tracestate", "baggage", telemetry.HeaderIdempotencyId}

// InstallTelemetry sets the global tracer provider and propagator for the test and restores the
// previous ones afterwards. Tests using it must not run in parallel.
func InstallTelemetry(t testing.TB) *tracetest.SpanRecorder {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previousProvider, previousPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(telemetry.NewPropagator([]string{"tracecontext", "baggage"}, []string{BaggageKey}))
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
	return recorder
}

// TracedContext is a request context as ingress leaves it: an active span, an idempotency id
// and one allowlisted baggage member.
func TracedContext(t testing.TB) (context.Context, string) {
	id, err := corectx.NewIdempotencyId()
	require.NoError(t, err)
	member, err := baggage.NewMember(BaggageKey, "leak-marker")
	require.NoError(t, err)
	bag, err := baggage.New(member)
	require.NoError(t, err)

	ctx := baggage.ContextWithBaggage(context.Background(), bag)
	ctx = corectx.WithIdempotencyId(ctx, id)
	ctx, span := telemetry.Tracer("httpclienttest").Start(ctx, "test.request")
	t.Cleanup(func() { span.End() })
	return ctx, id
}

// NewClient builds the shared client as production does, with the given propagation allowlist.
func NewClient(t testing.TB, logger logging.LoggerService, allowedHosts ...string) *client.HttpClient {
	if logger == nil {
		logger = logging.NewLogger()
	}
	shared, err := client.NewHttpClient(&client.HttpClientConfig{PropagationHostAllowlist: allowedHosts}, logger)
	require.NoError(t, err)
	return shared
}

// HeaderServer is an httptest server that records the headers of every request it receives.
type HeaderServer struct {
	*httptest.Server
	mutex   sync.Mutex
	headers []http.Header
}

func NewHeaderServer(t testing.TB, handler http.HandlerFunc) *HeaderServer {
	recorder := &HeaderServer{}
	recorder.Server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		recorder.mutex.Lock()
		recorder.headers = append(recorder.headers, request.Header.Clone())
		recorder.mutex.Unlock()
		if handler != nil {
			handler(writer, request)
		}
	}))
	t.Cleanup(recorder.Close)
	return recorder
}

// HostPort is the server's "host:port", the exact allowlist entry for it alone.
func (this *HeaderServer) HostPort(t testing.TB) string {
	parsed, err := url.Parse(this.URL)
	require.NoError(t, err)
	return parsed.Host
}

func (this *HeaderServer) Received() []http.Header {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return append([]http.Header(nil), this.headers...)
}

// AssertNoTelemetryHeaders fails unless the server received at least one request and none of
// them carried a telemetry header.
func (this *HeaderServer) AssertNoTelemetryHeaders(t testing.TB) {
	received := this.Received()
	require.NotEmpty(t, received, "the server received no request")
	for _, header := range received {
		for _, name := range TelemetryHeaders {
			assert.Empty(t, header.Values(name), "header %s leaked", name)
		}
	}
}
