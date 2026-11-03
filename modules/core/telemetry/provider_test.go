package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func baseSettings() Settings {
	return Settings{
		ServiceName:       "test-service",
		Propagators:       []string{"tracecontext", "baggage"},
		Sampler:           SamplerParentBasedRatio,
		SamplerArg:        "1.0",
		Exporter:          ExporterNone,
		OtlpProtocol:      ProtocolHttpProtobuf,
		BspMaxQueueSize:   16,
		BspMaxExportBatch: 4,
		BspExportTimeout:  time.Second,
		BspScheduleDelay:  time.Second,
	}
}

func resourceValue(provider *sdktrace.TracerProvider, key string) string {
	_, span := provider.Tracer("t").Start(context.Background(), "probe")
	defer span.End()
	readOnly := span.(sdktrace.ReadOnlySpan)
	value, _ := readOnly.Resource().Set().Value(attribute.Key(key))
	return value.AsString()
}

// AC-11 / SMP-03: exporter none and sampled=0 still produce valid, propagatable IDs.
func TestNewTracerProvider_NoneExporterStillCreatesValidIds(t *testing.T) {
	settings := baseSettings()
	settings.SamplerArg = "0"
	provider, warnings, err := NewTracerProvider(context.Background(), settings)
	require.NoError(t, err)
	assert.Empty(t, warnings)

	_, span := provider.Tracer("t").Start(context.Background(), "root")
	spanCtx := span.SpanContext()
	span.End()
	assert.True(t, spanCtx.IsValid())
	assert.False(t, spanCtx.IsSampled())
	require.NoError(t, provider.Shutdown(context.Background()))
}

func TestNewTracerProvider_OtlpWithoutEndpointExportsNothing(t *testing.T) {
	settings := baseSettings()
	settings.Exporter = ExporterOtlp
	provider, warnings, err := NewTracerProvider(context.Background(), settings)
	require.NoError(t, err)
	assert.Len(t, warnings, 1)
	require.NoError(t, provider.Shutdown(context.Background()))
}

func TestNewTracerProvider_ResourceAttributes(t *testing.T) {
	settings := baseSettings()
	settings.ResourceAttributes = map[string]string{
		"deployment.environment.name": "staging",
		"service.version":             "9.9.9",
		"service.instance.id":         "fixed",
	}
	provider, _, err := NewTracerProvider(context.Background(), settings)
	require.NoError(t, err)

	assert.Equal(t, "test-service", resourceValue(provider, "service.name"))
	assert.Equal(t, "9.9.9", resourceValue(provider, "service.version"))
	assert.Equal(t, "staging", resourceValue(provider, "deployment.environment.name"))
	instanceId := resourceValue(provider, "service.instance.id")
	assert.NotEqual(t, "fixed", instanceId)
	assert.NotEmpty(t, instanceId)
}

func TestNewSampler_FallsBackOnBadConfig(t *testing.T) {
	_, warnings := newSampler("bogus", "2.5")
	assert.Len(t, warnings, 2)

	sampler, warnings := newSampler(SamplerParentBasedRatio, "0.1")
	assert.Empty(t, warnings)
	assert.Contains(t, sampler.Description(), "ParentBased")
}

// SMP-02: a sampled remote parent stays sampled even with a root ratio of 0.
func TestNewSampler_ParentDecisionWins(t *testing.T) {
	sampler, _ := newSampler(SamplerParentBasedRatio, "0")
	provider := sdktrace.NewTracerProvider(sdktrace.WithSampler(sampler))
	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{1},
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	ctx := trace.ContextWithRemoteSpanContext(context.Background(), parent)
	_, span := provider.Tracer("t").Start(ctx, "child")
	defer span.End()
	assert.True(t, span.SpanContext().IsSampled())
	assert.Equal(t, parent.TraceID(), span.SpanContext().TraceID())
}

func TestOtlpHttpTracesUrl(t *testing.T) {
	cases := map[string]string{
		"http://collector:4318":           "http://collector:4318/v1/traces",
		"http://collector:4318/":          "http://collector:4318/v1/traces",
		"https://otel.example/base":       "https://otel.example/base/v1/traces",
		"http://collector:4318/v1/traces": "http://collector:4318/v1/traces",
	}
	for input, want := range cases {
		got, err := otlpHttpTracesUrl(input)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
	_, err := otlpHttpTracesUrl("collector:4318")
	assert.Error(t, err)
}

type countingLogger struct {
	logging.LoggerService
	warns int
}

func (this *countingLogger) Warn(context.Context, string, logging.Attr) {
	this.warns++
}

func TestRateLimitedErrorHandler_LogsOncePerInterval(t *testing.T) {
	logger := &countingLogger{}
	handler := newRateLimitedErrorHandler(logger, time.Hour)
	for i := 0; i < 10; i++ {
		handler.Handle(assert.AnError)
	}
	assert.Equal(t, 1, logger.warns)
	assert.Equal(t, 9, handler.suppressed)
}

func TestResourceLogFields_CarryServiceIdentity(t *testing.T) {
	res, err := newResource(Settings{ServiceName: "coreplatform-backend", ResourceAttributes: map[string]string{
		"service.version": "9.9.9", "deployment.environment.name": "staging",
	}})
	assert.NoError(t, err)
	fields := resourceLogFields(res)
	assert.Equal(t, "coreplatform-backend", fields["service.name"])
	assert.Equal(t, "staging", fields["deployment.environment.name"])
	assert.NotEmpty(t, fields["service.instance.id"])
	assert.NotEmpty(t, fields["service.version"])
}
