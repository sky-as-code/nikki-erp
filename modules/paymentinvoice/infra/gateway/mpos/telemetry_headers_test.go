package mpos

import (
	"net/http"
	"testing"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/httpclienttest"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// The gateway is not on the propagation allowlist, so a payment call from a traced request with
// an idempotency id sends it neither traceparent, baggage nor X-Idempotency-ID. The reply is
// refused (it is not a real mPOS reply); only what reached the gateway matters here.
func TestGatewayCallsCarryNoTelemetryHeaders(t *testing.T) {
	httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{}`))
	})
	shared := httpclienttest.NewClient(t, nil, "orders.internal")
	adapter := NewAdapter(Config{}, httpclient.NewHttpCaller(server.URL, shared, logging.NewLogger()))
	ctx, _ := httpclienttest.TracedContext(t)

	config := Config{MerchantId: "M1", SecretKey: testSecretKey}
	_, _ = adapter.call(corectx.NewRequestContext(ctx), config, "/api/check", map[string]any{}, nil)

	server.AssertNoTelemetryHeaders(t)
}
