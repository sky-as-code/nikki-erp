package middlewares

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	parentTraceId     = "4bf92f3577b34da6a3ce929d0e0e4736"
	parentSpanId      = "00f067aa0ba902b7"
	parentTraceparent = "00-" + parentTraceId + "-" + parentSpanId + "-01"
)

var traceIdPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

func stubAttribute(attrs []attribute.KeyValue, key string) (attribute.Value, bool) {
	for _, kv := range attrs {
		if string(kv.Key) == key {
			return kv.Value, true
		}
	}
	return attribute.Value{}, false
}

// AC-04: a direct call without trace context gets a root SERVER span; header, handler context and
// span agree on the trace ID.
func TestTracing_NoTraceparentStartsRootServerSpan(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	recorder := fixture.serve("/items/42?secret=token", nil)

	require.Equal(t, http.StatusOK, recorder.Code)
	traceId := recorder.Header().Get(HeaderTraceId)
	assert.Regexp(t, traceIdPattern, traceId)
	assert.Equal(t, traceId, fixture.seen.spanContext.TraceID().String())

	span := fixture.onlySpan(t)
	assert.Equal(t, traceId, span.SpanContext.TraceID().String())
	assert.False(t, span.Parent.IsValid())
	assert.Equal(t, trace.SpanKindServer, span.SpanKind)
	assert.Equal(t, "GET /items/:id", span.Name)
	route, _ := stubAttribute(span.Attributes, "http.route")
	assert.Equal(t, "/items/:id", route.AsString())
	status, _ := stubAttribute(span.Attributes, "http.response.status_code")
	assert.Equal(t, int64(http.StatusOK), status.AsInt64())
	for _, kv := range span.Attributes {
		assert.NotContains(t, kv.Value.Emit(), "secret=token")
	}
	assert.Equal(t, codes.Unset, span.Status.Code)
}

// AC-05: a valid parent is continued, never regenerated.
func TestTracing_ValidTraceparentIsContinued(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	recorder := fixture.serve("/items/1", http.Header{"Traceparent": {parentTraceparent}})

	assert.Equal(t, parentTraceId, recorder.Header().Get(HeaderTraceId))
	span := fixture.onlySpan(t)
	assert.Equal(t, parentTraceId, span.SpanContext.TraceID().String())
	assert.Equal(t, parentSpanId, span.Parent.SpanID().String())
	assert.True(t, span.Parent.IsRemote())
	assert.NotEqual(t, parentSpanId, span.SpanContext.SpanID().String())
}

// AC-06: an untrustworthy parent yields a new trace and never a 4xx.
func TestTracing_InvalidTraceparentFallsBackToNewTrace(t *testing.T) {
	cases := map[string][]string{
		"malformed":       {"garbage"},
		"zero trace id":   {"00-00000000000000000000000000000000-" + parentSpanId + "-01"},
		"forbidden ver":   {"ff-" + parentTraceId + "-" + parentSpanId + "-01"},
		"comma list":      {parentTraceparent + "," + parentTraceparent},
		"repeated header": {parentTraceparent, parentTraceparent},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			fixture := newIngressFixture(t, sdktrace.AlwaysSample())
			header := http.Header{"Traceparent": values, "Tracestate": {"vendor=1"}}
			recorder := fixture.serve("/items/1", header)

			assert.Equal(t, http.StatusOK, recorder.Code)
			traceId := recorder.Header().Get(HeaderTraceId)
			assert.Regexp(t, traceIdPattern, traceId)
			assert.NotEqual(t, parentTraceId, traceId)
			span := fixture.onlySpan(t)
			assert.False(t, span.Parent.IsValid())
			assert.Empty(t, span.SpanContext.TraceState().String())
		})
	}
}

// AC-07: a broken tracestate is dropped by the propagator; the valid parent survives.
func TestTracing_InvalidTracestateKeepsTrace(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	header := http.Header{"Traceparent": {parentTraceparent}, "Tracestate": {"=,,bad key=%%"}}
	recorder := fixture.serve("/items/1", header)

	assert.Equal(t, parentTraceId, recorder.Header().Get(HeaderTraceId))
	span := fixture.onlySpan(t)
	assert.Equal(t, parentSpanId, span.Parent.SpanID().String())
	assert.Empty(t, span.SpanContext.TraceState().String())
}

// AC-11 + ID-AC-11: unsampled, nothing exported, yet IDs exist and the idempotency ID works.
func TestTracing_UnsampledStillHasIdsAndIdempotency(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.NeverSample())
	unsampledParent := "00-" + parentTraceId + "-" + parentSpanId + "-00"
	recorder := fixture.serve("/items/1", http.Header{"Traceparent": {unsampledParent}})

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, parentTraceId, recorder.Header().Get(HeaderTraceId))
	assert.True(t, fixture.seen.spanContext.IsValid())
	assert.False(t, fixture.seen.spanContext.IsSampled())
	assert.Empty(t, fixture.exporter.GetSpans())
	assert.NotEmpty(t, recorder.Header().Get(HeaderIdempotencyId))
	assert.Equal(t, recorder.Header().Get(HeaderIdempotencyId), fixture.seen.idempotencyId)
}

// AC-15: application-handled 4xx/5xx and a panic all answer X-Trace-ID, end the span once, are
// access-logged, and only 5xx marks the span as an error.
func TestTracing_ErrorResponsesKeepTraceAndStatusConvention(t *testing.T) {
	cases := []struct {
		target     string
		wantStatus int
		wantError  bool
	}{
		{"/unauthorized", http.StatusUnauthorized, false},
		{"/forbidden", http.StatusForbidden, false},
		{"/does-not-exist", http.StatusNotFound, false},
		{"/fail", http.StatusInternalServerError, true},
		{"/panic", http.StatusInternalServerError, true},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			fixture := newIngressFixture(t, sdktrace.AlwaysSample())
			recorder := fixture.serve(tc.target, nil)

			assert.Equal(t, tc.wantStatus, recorder.Code)
			assert.Regexp(t, traceIdPattern, recorder.Header().Get(HeaderTraceId))
			span := fixture.onlySpan(t)
			assert.Equal(t, tc.wantError, span.Status.Code == codes.Error)
			status, _ := stubAttribute(span.Attributes, "http.response.status_code")
			assert.Equal(t, int64(tc.wantStatus), status.AsInt64())

			lines := fixture.logger.find(accessLogMessage)
			require.Len(t, lines, 1)
			assert.Equal(t, tc.wantStatus, lines[0].data["http.response.status_code"])
		})
	}
}

func TestRecover_PanicIsSanitizedRecordedAndAnsweredAs500(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	recorder := fixture.serve("/panic", nil)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.JSONEq(t, `{"message":"Internal Server Error"}`, recorder.Body.String())

	span := fixture.onlySpan(t)
	require.Len(t, span.Events, 1)
	assert.Equal(t, "exception", span.Events[0].Name)
	message, _ := stubAttribute(span.Events[0].Attributes, "exception.message")
	assert.NotContains(t, message.AsString(), "\n")

	panics := fixture.logger.find("handler panicked")
	require.Len(t, panics, 1)
	assert.False(t, strings.Contains(panics[0].err.Error(), "\n"))
}

func TestSanitizePanicMessage_Truncates(t *testing.T) {
	message := sanitizePanicMessage(strings.Repeat("é", 300))
	assert.Equal(t, panicMessageMaxRunes+3, len([]rune(message)))
}
