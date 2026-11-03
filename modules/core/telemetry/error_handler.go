package telemetry

import (
	"context"
	"sync"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const errorLogInterval = 30 * time.Second

// rateLimitedErrorHandler receives the SDK's internal errors (export failures, dropped spans),
// which repeat for as long as a receiver is down: one line per interval reports how many it hid.
type rateLimitedErrorHandler struct {
	logger     logging.LoggerService
	interval   time.Duration
	mutex      sync.Mutex
	lastLogged time.Time
	suppressed int
}

func newRateLimitedErrorHandler(logger logging.LoggerService, interval time.Duration) *rateLimitedErrorHandler {
	return &rateLimitedErrorHandler{logger: logger, interval: interval}
}

func (this *rateLimitedErrorHandler) Handle(err error) {
	if err == nil || this.logger == nil {
		return
	}
	this.mutex.Lock()
	now := time.Now()
	if !this.lastLogged.IsZero() && now.Sub(this.lastLogged) < this.interval {
		this.suppressed++
		this.mutex.Unlock()
		return
	}
	suppressed := this.suppressed
	this.suppressed = 0
	this.lastLogged = now
	this.mutex.Unlock()

	this.logger.Warn(context.Background(), "OpenTelemetry SDK error", logging.Attr{
		"error":      err.Error(),
		"suppressed": suppressed,
	})
}
