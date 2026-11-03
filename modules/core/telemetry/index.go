package telemetry

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

var (
	stateMutex      sync.RWMutex
	tracerProvider  *sdktrace.TracerProvider
	activeSettings  Settings
	defaultShutdown = 5 * time.Second
)

// InitSubModule installs the global tracer provider, propagator and error handler. It runs
// right after config, before anything that builds a client or serves traffic.
func InitSubModule(cfg config.ConfigService, logger logging.LoggerService) error {
	settings := LoadSettings(cfg)
	res, err := newResource(settings)
	if err != nil {
		return err
	}
	logging.SetResourceFields(resourceLogFields(res))
	provider, warnings, err := newTracerProviderWith(context.Background(), settings, res)
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		logger.Warn(context.Background(), "telemetry setting fell back to a default", logging.Attr{"reason": warning})
	}
	otel.SetErrorHandler(newRateLimitedErrorHandler(logger, errorLogInterval))
	otel.SetTextMapPropagator(NewPropagator(settings.Propagators, settings.BaggageAllowlist))
	otel.SetTracerProvider(provider)

	stateMutex.Lock()
	tracerProvider = provider
	activeSettings = settings
	stateMutex.Unlock()

	logger.Info(context.Background(), "telemetry initialized", logging.Attr{
		"service_name": settings.ServiceName,
		"exporter":     settings.Exporter,
		"sampler":      settings.Sampler,
	})
	return deps.Register(func() trace.TracerProvider { return provider })
}

// PropagationHostAllowlist is the configured CORE.TELEMETRY.PROPAGATION_HOST_ALLOWLIST, for
// building NewTransport.
func PropagationHostAllowlist() []string {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	return append([]string{}, activeSettings.HostAllowlist...)
}

// ShutdownTimeout is CORE.TELEMETRY.SHUTDOWN_FLUSH_SECS, or 5 seconds before InitSubModule ran.
func ShutdownTimeout() time.Duration {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	if activeSettings.ShutdownFlush > 0 {
		return activeSettings.ShutdownFlush
	}
	return defaultShutdown
}

// Shutdown flushes pending spans and stops the provider. Call it after the HTTP server has
// drained, with a bounded ctx: an unreachable receiver must not hold the process.
func Shutdown(ctx context.Context) error {
	stateMutex.RLock()
	provider := tracerProvider
	stateMutex.RUnlock()
	if provider == nil {
		return nil
	}
	return provider.Shutdown(ctx)
}
