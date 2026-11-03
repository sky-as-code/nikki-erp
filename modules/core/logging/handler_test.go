package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const sampleIdempotencyId = "01890a5d-ac96-774b-bcce-b302099a8057"

var timestampPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z$`)

func newBufferLogger() (logging.LoggerService, *bytes.Buffer) {
	buffer := &bytes.Buffer{}
	return logging.NewLoggerTo(buffer, logging.LevelDebug), buffer
}

func readRecords(t *testing.T, buffer *bytes.Buffer) []map[string]any {
	t.Helper()
	records := []map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(buffer.String()), "\n") {
		record := map[string]any{}
		require.NoError(t, json.Unmarshal([]byte(line), &record), line)
		records = append(records, record)
	}
	return records
}

func newTracer(t *testing.T, sampler sdktrace.Sampler) trace.Tracer {
	t.Helper()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSampler(sampler))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	return provider.Tracer("logging-test")
}

// AC-14, LOG-03, LOG-ID-03: a log with no operation has every operation field present and null.
func TestHandler_StartupLogHasNullTelemetryFields(t *testing.T) {
	logger, buffer := newBufferLogger()
	logger.Info(context.Background(), "server starting", nil)

	record := readRecords(t, buffer)[0]
	for _, name := range []string{"trace_id", "span_id", "idempotency_id", "tenant_id", "org_id"} {
		value, present := record[name]
		assert.True(t, present, name)
		assert.Nil(t, value, name)
	}
	assert.Equal(t, "INFO", record["severity"])
	assert.Equal(t, "server starting", record["message"])
	assert.Regexp(t, timestampPattern, record["timestamp"])
	assert.Contains(t, record["source"], "handler_test.go:")
	assert.NotContains(t, record, "data")
	assert.NotContains(t, record, "http.route")
}

// AC-13, LOG-01, LOG-02: the logger reads the span active at each write, not one captured earlier.
func TestHandler_ChildSpanLogCarriesChildSpanId(t *testing.T) {
	tracer := newTracer(t, sdktrace.AlwaysSample())
	logger, buffer := newBufferLogger()

	appCtx, appSpan := tracer.Start(context.Background(), "inventory_product.update")
	logger.Info(appCtx, "application service", nil)
	repoCtx, repoSpan := tracer.Start(appCtx, "repository.update")
	logger.Info(repoCtx, "repository", nil)
	repoSpan.End()
	appSpan.End()

	records := readRecords(t, buffer)
	require.Len(t, records, 2)
	assert.Equal(t, appSpan.SpanContext().TraceID().String(), records[0]["trace_id"])
	assert.Equal(t, records[0]["trace_id"], records[1]["trace_id"])
	assert.Equal(t, appSpan.SpanContext().SpanID().String(), records[0]["span_id"])
	assert.Equal(t, repoSpan.SpanContext().SpanID().String(), records[1]["span_id"])
	assert.NotEqual(t, records[0]["span_id"], records[1]["span_id"])
	assert.NotContains(t, records[1], "trace_flags")
}

// LOG-02: an unsampled span still has valid ids, and they are logged with its flags.
func TestHandler_UnsampledSpanStillLogsIds(t *testing.T) {
	tracer := newTracer(t, sdktrace.NeverSample())
	logger, buffer := newBufferLogger()

	ctx, span := tracer.Start(context.Background(), "unsampled")
	logger.Warn(ctx, "not sampled", nil)
	span.End()

	record := readRecords(t, buffer)[0]
	assert.Regexp(t, `^[0-9a-f]{32}$`, record["trace_id"])
	assert.Regexp(t, `^[0-9a-f]{16}$`, record["span_id"])
	assert.NotContains(t, record, "trace_flags")
}

// ID-AC-13, LOG-ID-01..03: a request log carries the id from ctx; a global log has null; the
// logger never generates one.
func TestHandler_IdempotencyIdFromContextOnly(t *testing.T) {
	logger, buffer := newBufferLogger()
	requestCtx := corectx.WithIdempotencyId(context.Background(), sampleIdempotencyId)

	logger.Infof(requestCtx, "inside request %d", 1)
	logger.Info(context.Background(), "global", nil)
	logger.Info(context.Background(), "global again", nil)

	records := readRecords(t, buffer)
	assert.Equal(t, sampleIdempotencyId, records[0]["idempotency_id"])
	assert.Nil(t, records[1]["idempotency_id"])
	assert.Nil(t, records[2]["idempotency_id"])
}

func TestHandler_IdempotencyIdSurvivesRequestContextWrapping(t *testing.T) {
	logger, buffer := newBufferLogger()
	inner := corectx.WithIdempotencyId(context.Background(), sampleIdempotencyId)
	requestCtx := corectx.NewRequestContext(inner)
	dyn.WithActingOrg(requestCtx, model.Id("01JORG000000000000000000AA"))

	logger.Debug(requestCtx, "wrapped", nil)

	record := readRecords(t, buffer)[0]
	assert.Equal(t, sampleIdempotencyId, record["idempotency_id"])
	assert.Equal(t, "01JORG000000000000000000AA", record["org_id"])
	assert.Equal(t, "DEBUG", record["severity"])
}

// LOG-05, LOG-ID-02: caller data stays nested and cannot replace a telemetry or resource field.
func TestHandler_CallerDataCannotOverrideTelemetryFields(t *testing.T) {
	tracer := newTracer(t, sdktrace.AlwaysSample())
	logging.SetResourceFields(map[string]string{logging.FieldServiceName: "coreplatform-backend"})
	t.Cleanup(func() { logging.SetResourceFields(nil) })
	logger, buffer := newBufferLogger()

	ctx, span := tracer.Start(corectx.WithIdempotencyId(context.Background(), sampleIdempotencyId), "op")
	logger.Info(ctx, "attempted override", logging.Attr{
		"trace_id": "forged", "span_id": "forged", "idempotency_id": "forged", "tenant_id": "forged",
		"service.name": "forged", "severity": "forged", "message": "forged", "http.route": "/forged",
	})
	span.End()

	record := readRecords(t, buffer)[0]
	assert.Equal(t, span.SpanContext().TraceID().String(), record["trace_id"])
	assert.Equal(t, span.SpanContext().SpanID().String(), record["span_id"])
	assert.Equal(t, sampleIdempotencyId, record["idempotency_id"])
	assert.Nil(t, record["tenant_id"])
	assert.Equal(t, "coreplatform-backend", record["service.name"])
	assert.Equal(t, "INFO", record["severity"])
	assert.Equal(t, "attempted override", record["message"])
	assert.NotContains(t, record, "http.route")
	data, ok := record["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "forged", data["trace_id"])
	assert.Equal(t, "forged", data["service.name"])
}

// A hook may only fill the operation fields; anything else it returns is dropped.
func TestHandler_ContextHookCannotSetOtherFields(t *testing.T) {
	logging.RegisterContextFields(func(context.Context) map[string]any {
		return map[string]any{"trace_id": "hooked", "service.name": "hooked", "custom": "hooked"}
	})
	logger, buffer := newBufferLogger()
	logger.Info(context.Background(), "hooked", nil)

	record := readRecords(t, buffer)[0]
	assert.Nil(t, record["trace_id"])
	assert.NotEqual(t, "hooked", record["service.name"])
	assert.NotContains(t, record, "custom")
}

func TestHandler_ResourceFieldsAreTopLevel(t *testing.T) {
	logging.SetResourceFields(map[string]string{
		logging.FieldServiceName: "coreplatform-backend",
		"service.version":        "1.2.3",
		"unrelated":              "dropped",
	})
	t.Cleanup(func() { logging.SetResourceFields(nil) })
	logger, buffer := newBufferLogger()
	logger.Info(context.Background(), "resource", nil)

	record := readRecords(t, buffer)[0]
	assert.Equal(t, "coreplatform-backend", record["service.name"])
	assert.NotContains(t, record, "service.version")
	assert.NotContains(t, record, "service.instance.id")
	assert.NotContains(t, record, "deployment.environment.name")
	assert.NotContains(t, record, "unrelated")
}

// The access log is the only writer of the HTTP fields, and only through ctx.
func TestHandler_HttpFieldsOnlyFromAccessLogContext(t *testing.T) {
	logger, buffer := newBufferLogger()
	ctx := logging.WithHttpFields(context.Background(), logging.HttpFields{
		Method: "GET", Route: "/v1/products/:id", StatusCode: 200, DurationMs: 1.5,
	})
	logger.Info(ctx, "http request", nil)

	record := readRecords(t, buffer)[0]
	assert.Equal(t, "GET", record["http.request.method"])
	assert.Equal(t, "/v1/products/:id", record["http.route"])
	assert.Equal(t, float64(200), record["http.response.status_code"])
	assert.Equal(t, 1.5, record["duration_ms"])
}

// Nested shape: the error report sits under data, grouped by a bounded error.type.
func TestHandler_ErrorRecordShape(t *testing.T) {
	logger, buffer := newBufferLogger()
	logger.Error(context.Background(), "save failed", errors.Wrap(errors.New("disk full"), "write file"))
	logger.Errorf(context.Background(), "plain %s", "error")

	records := readRecords(t, buffer)
	require.Len(t, records, 2)
	assert.Equal(t, "ERROR", records[0]["severity"])
	assert.Equal(t, "*errors.Error", records[0]["error.type"])
	data, ok := records[0]["data"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, data["message"], "disk full")
	assert.NotContains(t, records[1], "error.type")
	assert.Equal(t, "plain error", records[1]["message"])
}

func TestHandler_LevelFiltersRecords(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := logging.NewLoggerTo(buffer, logging.LevelWarn)
	logger.Info(context.Background(), "dropped", nil)
	logger.Debugf(context.Background(), "dropped %d", 1)
	logger.Warn(context.Background(), "kept", logging.Attr{"err": errors.New("as text")})

	records := readRecords(t, buffer)
	require.Len(t, records, 1)
	assert.Equal(t, "as text", records[0]["data"].(map[string]any)["err"])
}

// Third-party slog users (watermill through InnerLogger) keep their attributes under `attrs`.
func TestHandler_ForeignSlogAttrsAreNested(t *testing.T) {
	logger, buffer := newBufferLogger()
	slogger := logger.InnerLogger().(*slog.Logger).With("topic", "orders").WithGroup("watermill")
	slogger.Info("subscribed", "trace_id", "forged", "err", errors.New("boom"))

	record := readRecords(t, buffer)[0]
	assert.Nil(t, record["trace_id"])
	attrs, ok := record["attrs"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "orders", attrs["topic"])
	assert.Equal(t, "forged", attrs["watermill.trace_id"])
	assert.Equal(t, "boom", attrs["watermill.err"])
}
