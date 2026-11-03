package vietqr

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/httpclienttest"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// The gateway is not on the propagation allowlist, so neither the login nor the payment call
// from a traced request with an idempotency id sends traceparent, baggage or X-Idempotency-ID.
func TestGatewayCallsCarryNoTelemetryHeaders(t *testing.T) {
	httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"access_token":"tok","token_type":"Bearer","expires_in":300}`))
	})
	shared := httpclienttest.NewClient(t, nil, "orders.internal")
	adapter := NewAdapter(Config{}, httpclient.NewHttpCaller(server.URL, shared, logging.NewLogger()))
	ctx, _ := httpclienttest.TracedContext(t)

	config := Config{Username: "user", Password: "pass"}
	_, err := adapter.sendOnce(corectx.NewRequestContext(ctx), config, "/api/check", map[string]any{})

	require.NoError(t, err)
	require.Len(t, server.Received(), 2, "login, then the call")
	server.AssertNoTelemetryHeaders(t)
}
