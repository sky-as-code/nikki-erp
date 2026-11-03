package client

import (
	"net/http"
	"strings"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

// A downstream value is logged only as a bounded prefix: it is foreign input.
const maxLoggedIdLength = 64

// newInstrumentedTransport layers, outermost first: telemetry.NewTransport (strips inherited
// propagation headers, opens the CLIENT span, injects for allowlisted hosts), then the
// idempotency diagnostics, which therefore see each hop exactly as it goes on the wire.
func newInstrumentedTransport(
	base http.RoundTripper, allowedHosts []string, logger logging.LoggerService,
) http.RoundTripper {
	if logger == nil {
		logger = logging.Logger()
	}
	diagnostics := &idempotencyDiagnostics{next: base, logger: logger}
	return telemetry.NewTransport(diagnostics, allowedHosts)
}

// idempotencyDiagnostics never sets or changes an ID. It reports an outbound call made without
// one in ctx (the ID was lost upstream, and minting one here would hide that), and a downstream
// that answered a sent ID with a different one (PROP-06, ID-AC-17).
type idempotencyDiagnostics struct {
	next   http.RoundTripper
	logger logging.LoggerService
}

func (this *idempotencyDiagnostics) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx := request.Context()
	if _, ok := corectx.IdempotencyIdOf(ctx); !ok {
		this.logger.Warn(ctx, "outbound http call has no idempotency id in context", logging.Attr{
			"host": request.URL.Host,
		})
	}
	response, err := this.next.RoundTrip(request)
	if err != nil {
		return response, err
	}
	sent := request.Header.Get(telemetry.HeaderIdempotencyId)
	received := response.Header.Get(telemetry.HeaderIdempotencyId)
	if sent != "" && received != "" && !strings.EqualFold(sent, received) {
		this.logger.Warn(ctx, "downstream answered with a different idempotency id; keeping ours", logging.Attr{
			"host":                    request.URL.Host,
			"received_idempotency_id": truncate(received, maxLoggedIdLength),
		})
	}
	return response, nil
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
