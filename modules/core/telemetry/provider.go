package telemetry

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"go.bryk.io/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/sky-as-code/nikki-erp/common/env"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
)

// ServiceVersion is set at build time: -ldflags "-X .../modules/core/telemetry.ServiceVersion=1.2.3".
var ServiceVersion = ""

const (
	attrDeploymentEnvironment = "deployment.environment.name"
	otlpTracesPath            = "/v1/traces"
)

// NewTracerProvider always returns a real SDK provider, so trace and span IDs exist and propagate
// even when nothing is exported. The returned warnings are settings that fell back to a default.
func NewTracerProvider(ctx context.Context, settings Settings) (*sdktrace.TracerProvider, []string, error) {
	res, err := newResource(settings)
	if err != nil {
		return nil, nil, err
	}
	return newTracerProviderWith(ctx, settings, res)
}

func newTracerProviderWith(
	ctx context.Context, settings Settings, res *resource.Resource,
) (*sdktrace.TracerProvider, []string, error) {
	sampler, warnings := newSampler(settings.Sampler, settings.SamplerArg)
	options := []sdktrace.TracerProviderOption{sdktrace.WithSampler(sampler), sdktrace.WithResource(res)}

	exporter, exporterWarning, err := newExporter(ctx, settings)
	if err != nil {
		return nil, warnings, err
	}
	if exporterWarning != "" {
		warnings = append(warnings, exporterWarning)
	}
	if exporter != nil {
		options = append(options, sdktrace.WithBatcher(exporter, batchOptions(settings)...))
	}
	return sdktrace.NewTracerProvider(options...), warnings, nil
}

func batchOptions(settings Settings) []sdktrace.BatchSpanProcessorOption {
	return []sdktrace.BatchSpanProcessorOption{
		sdktrace.WithMaxQueueSize(settings.BspMaxQueueSize),
		sdktrace.WithMaxExportBatchSize(settings.BspMaxExportBatch),
		sdktrace.WithExportTimeout(settings.BspExportTimeout),
		sdktrace.WithBatchTimeout(settings.BspScheduleDelay),
	}
}

func newSampler(name string, rawArg string) (sdktrace.Sampler, []string) {
	warnings := []string{}
	ratio, err := strconv.ParseFloat(rawArg, 64)
	if err != nil || ratio < 0 || ratio > 1 {
		warnings = append(warnings, fmt.Sprintf("OTEL.TRACES_SAMPLER_ARG %q is not a ratio in [0,1]; using 1.0", rawArg))
		ratio = 1
	}
	switch name {
	case SamplerParentBasedRatio:
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio)), warnings
	case "parentbased_always_on":
		return sdktrace.ParentBased(sdktrace.AlwaysSample()), warnings
	case "parentbased_always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample()), warnings
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(ratio), warnings
	case "always_on":
		return sdktrace.AlwaysSample(), warnings
	case "always_off":
		return sdktrace.NeverSample(), warnings
	}
	warnings = append(warnings, fmt.Sprintf("OTEL.TRACES_SAMPLER %q is unknown; using %s", name, SamplerParentBasedRatio))
	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio)), warnings
}

// resourceLogFields are the resource attributes the log handler may repeat; it keeps service.name only (§7.2).
func resourceLogFields(res *resource.Resource) map[string]string {
	fields := map[string]string{}
	for _, attr := range res.Attributes() {
		fields[string(attr.Key)] = attr.Value.Emit()
	}
	return fields
}

// newResource builds the resource from settings alone: resource.Default would also read
// OTEL_RESOURCE_ATTRIBUTES itself and bypass the config chain.
func newResource(settings Settings) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{}
	for key, value := range settings.ResourceAttributes {
		attrs = append(attrs, attribute.String(key, value))
	}
	instanceId, err := newInstanceId()
	if err != nil {
		return nil, err
	}
	attrs = append(attrs,
		semconv.ServiceName(settings.ServiceName),
		semconv.ServiceVersion(resolveServiceVersion(settings.ResourceAttributes)),
		semconv.ServiceInstanceID(instanceId),
		attribute.String(attrDeploymentEnvironment, resolveEnvironment(settings.ResourceAttributes)),
	)
	return resource.NewWithAttributes(semconv.SchemaURL, attrs...), nil
}

func resolveServiceVersion(resourceAttrs map[string]string) string {
	if ServiceVersion != "" {
		return ServiceVersion
	}
	if version := resourceAttrs[string(semconv.ServiceVersionKey)]; version != "" {
		return version
	}
	return config.CommitId
}

func resolveEnvironment(resourceAttrs map[string]string) string {
	if name := resourceAttrs[attrDeploymentEnvironment]; name != "" {
		return name
	}
	if appEnv := env.AppEnv(); appEnv != "" {
		return string(appEnv)
	}
	return string(env.APP_ENV_LOCAL)
}

// newInstanceId must differ per replica and per restart, so it is never taken from config.
func newInstanceId() (string, error) {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown-host"
	}
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return "", errors.Wrap(err, "generate service.instance.id suffix")
	}
	return fmt.Sprintf("%s-%d-%s", hostname, os.Getpid(), hex.EncodeToString(suffix)), nil
}

func newExporter(ctx context.Context, settings Settings) (sdktrace.SpanExporter, string, error) {
	switch settings.Exporter {
	case ExporterNone, "":
		return nil, "", nil
	case ExporterConsole, "stdout":
		exporter, err := stdouttrace.New(stdouttrace.WithWriter(os.Stderr))
		return exporter, "", errors.Wrap(err, "create console span exporter")
	case ExporterOtlp:
		return newOtlpExporter(ctx, settings)
	}
	return nil, fmt.Sprintf("OTEL.TRACES_EXPORTER %q is unknown; exporting nothing", settings.Exporter), nil
}

// newOtlpExporter refuses to guess a receiver: with no endpoint it exports nothing rather than
// sending to the SDK's localhost default.
func newOtlpExporter(ctx context.Context, settings Settings) (sdktrace.SpanExporter, string, error) {
	if settings.OtlpEndpoint == "" {
		return nil, "OTEL.TRACES_EXPORTER is otlp but OTEL.EXPORTER_OTLP_ENDPOINT is empty; exporting nothing", nil
	}
	if settings.OtlpProtocol == ProtocolGrpc {
		exporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpointURL(settings.OtlpEndpoint),
			otlptracegrpc.WithHeaders(settings.OtlpHeaders),
			otlptracegrpc.WithTimeout(settings.BspExportTimeout),
		)
		return exporter, "", errors.Wrap(err, "create OTLP gRPC span exporter")
	}
	warning := ""
	if settings.OtlpProtocol != ProtocolHttpProtobuf {
		warning = fmt.Sprintf("OTEL.EXPORTER_OTLP_PROTOCOL %q is unsupported; using %s",
			settings.OtlpProtocol, ProtocolHttpProtobuf)
	}
	endpoint, err := otlpHttpTracesUrl(settings.OtlpEndpoint)
	if err != nil {
		return nil, warning, err
	}
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithHeaders(settings.OtlpHeaders),
		otlptracehttp.WithTimeout(settings.BspExportTimeout),
	)
	return exporter, warning, errors.Wrap(err, "create OTLP HTTP span exporter")
}

// otlpHttpTracesUrl appends the signal path, as the OTLP spec requires for the non-signal
// OTEL_EXPORTER_OTLP_ENDPOINT variable.
func otlpHttpTracesUrl(base string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil || parsed.Host == "" {
		return "", errors.Errorf("OTEL.EXPORTER_OTLP_ENDPOINT %q is not an absolute URL", base)
	}
	if !strings.HasSuffix(parsed.Path, otlpTracesPath) {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + otlpTracesPath
	}
	return parsed.String(), nil
}
