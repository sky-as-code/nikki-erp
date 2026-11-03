package telemetry

import (
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

const (
	ExporterNone    = "none"
	ExporterOtlp    = "otlp"
	ExporterConsole = "console"

	ProtocolHttpProtobuf = "http/protobuf"
	ProtocolGrpc         = "grpc"

	SamplerParentBasedRatio = "parentbased_traceidratio"
)

type Settings struct {
	ServiceName        string
	ResourceAttributes map[string]string
	Propagators        []string
	Sampler            string
	SamplerArg         string
	Exporter           string
	OtlpProtocol       string
	OtlpEndpoint       string
	OtlpHeaders        map[string]string
	BspMaxQueueSize    int
	BspMaxExportBatch  int
	BspExportTimeout   time.Duration
	BspScheduleDelay   time.Duration
	BaggageAllowlist   []string
	HostAllowlist      []string
	ShutdownFlush      time.Duration
}

// LoadSettings passes a default for every key, so a binary whose config predates telemetry
// still boots with exporting off.
func LoadSettings(cfg config.ConfigService) Settings {
	return Settings{
		ServiceName:        cfg.GetStr(c.OtelServiceName, "nikki-erp"),
		ResourceAttributes: parseKeyValueList(cfg.GetStr(c.OtelResourceAttributes, "")),
		Propagators:        parseList(cfg.GetStr(c.OtelPropagators, "tracecontext,baggage")),
		Sampler:            strings.ToLower(strings.TrimSpace(cfg.GetStr(c.OtelTracesSampler, SamplerParentBasedRatio))),
		SamplerArg:         strings.TrimSpace(cfg.GetStr(c.OtelTracesSamplerArg, "1.0")),
		Exporter:           strings.ToLower(strings.TrimSpace(cfg.GetStr(c.OtelTracesExporter, ExporterNone))),
		OtlpProtocol:       strings.ToLower(strings.TrimSpace(cfg.GetStr(c.OtelExporterOtlpProtocol, ProtocolHttpProtobuf))),
		OtlpEndpoint:       strings.TrimSpace(cfg.GetStr(c.OtelExporterOtlpEndpoint, "")),
		OtlpHeaders:        parseKeyValueList(cfg.GetStr(c.OtelExporterOtlpHeaders, "")),
		BspMaxQueueSize:    cfg.GetInt(c.OtelBspMaxQueueSize, "2048"),
		BspMaxExportBatch:  cfg.GetInt(c.OtelBspMaxExportBatch, "512"),
		BspExportTimeout:   millis(cfg.GetInt(c.OtelBspExportTimeoutMs, "30000")),
		BspScheduleDelay:   millis(cfg.GetInt(c.OtelBspScheduleDelayMs, "5000")),
		BaggageAllowlist:   parseList(cfg.GetStr(c.TelemetryBaggageAllowlist, "")),
		HostAllowlist:      parseList(strings.ToLower(cfg.GetStr(c.TelemetryPropagationHostAllowlist, ""))),
		ShutdownFlush:      time.Duration(cfg.GetInt(c.TelemetryShutdownFlushSecs, "5")) * time.Second,
	}
}

func millis(value int) time.Duration {
	return time.Duration(value) * time.Millisecond
}

func parseList(raw string) []string {
	out := []string{}
	for _, item := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func parseKeyValueList(raw string) map[string]string {
	out := map[string]string{}
	for _, item := range parseList(raw) {
		key, value, found := strings.Cut(item, "=")
		key = strings.TrimSpace(key)
		if found && key != "" {
			out[key] = strings.TrimSpace(value)
		}
	}
	return out
}
