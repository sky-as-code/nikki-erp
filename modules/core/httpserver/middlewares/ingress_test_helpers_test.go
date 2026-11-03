package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v5"
	"go.bryk.io/pkg/errors"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

type logEntry struct {
	level   string
	message string
	data    logging.Attr
	err     error
}

type recordingLogger struct {
	logging.LoggerService
	mutex   sync.Mutex
	entries []logEntry
}

func (this *recordingLogger) record(entry logEntry) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.entries = append(this.entries, entry)
}

func (this *recordingLogger) Info(ctx context.Context, message string, data logging.Attr) {
	this.record(logEntry{level: "info", message: message, data: withHttpFields(ctx, data)})
}

func (this *recordingLogger) Warn(ctx context.Context, message string, data logging.Attr) {
	this.record(logEntry{level: "warn", message: message, data: withHttpFields(ctx, data)})
}

func (this *recordingLogger) Error(ctx context.Context, message string, err error) {
	this.record(logEntry{level: "error", message: message, err: err})
}

// withHttpFields folds the access-log fields carried by ctx into data, so assertions read one map.
func withHttpFields(ctx context.Context, data logging.Attr) logging.Attr {
	merged := logging.Attr{}
	for key, value := range data {
		merged[key] = value
	}
	if fields, ok := logging.HttpFieldsOf(ctx); ok {
		merged["http.request.method"] = fields.Method
		merged["http.route"] = fields.Route
		merged["http.response.status_code"] = fields.StatusCode
		merged["duration_ms"] = fields.DurationMs
	}
	return merged
}

func (this *recordingLogger) find(message string) []logEntry {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	found := []logEntry{}
	for _, entry := range this.entries {
		if entry.message == message {
			found = append(found, entry)
		}
	}
	return found
}

// handlerSeen is what the business handler observed in its request context.
type handlerSeen struct {
	calls         int
	spanContext   trace.SpanContext
	idempotencyId string
}

type ingressFixture struct {
	server   *echo.Echo
	exporter *tracetest.InMemoryExporter
	logger   *recordingLogger
	seen     *handlerSeen
}

func newIngressFixture(t *testing.T, sampler sdktrace.Sampler) *ingressFixture {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter), sdktrace.WithSampler(sampler))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(telemetry.NewPropagator([]string{"tracecontext", "baggage"}, nil))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	fixture := &ingressFixture{server: echo.New(), exporter: exporter, logger: &recordingLogger{}, seen: &handlerSeen{}}
	fixture.server.Use(TracingMiddleware, AccessLogMiddleware(fixture.logger),
		IdempotencyMiddleware(fixture.logger), RecoverMiddleware(fixture.logger), RequestContextMiddleware)
	fixture.registerRoutes()
	return fixture
}

func (this *ingressFixture) registerRoutes() {
	this.server.GET("/items/:id", func(echoCtx *echo.Context) error {
		ctx := echoCtx.Request().Context()
		this.seen.calls++
		this.seen.spanContext = trace.SpanContextFromContext(ctx)
		this.seen.idempotencyId, _ = corectx.IdempotencyIdOf(ctx)
		return echoCtx.JSON(http.StatusOK, map[string]string{"ok": "yes"})
	})
	this.server.GET("/unauthorized", func(echoCtx *echo.Context) error {
		return echoCtx.JSON(http.StatusUnauthorized, map[string]string{})
	})
	this.server.GET("/forbidden", func(echoCtx *echo.Context) error {
		return echoCtx.JSON(http.StatusForbidden, map[string]string{})
	})
	this.server.GET("/fail", func(echoCtx *echo.Context) error {
		return errors.New("database unavailable")
	})
	this.server.GET("/panic", func(echoCtx *echo.Context) error {
		panic("boom\nInjected: line")
	})
}

func (this *ingressFixture) serve(target string, header http.Header) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	for name, values := range header {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
	recorder := httptest.NewRecorder()
	this.server.ServeHTTP(recorder, request)
	return recorder
}

func (this *ingressFixture) onlySpan(t *testing.T) tracetest.SpanStub {
	t.Helper()
	spans := this.exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("want exactly one exported span, got %d", len(spans))
	}
	return spans[0]
}
