package momo

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/httpclienttest"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// The gateway is not on the propagation allowlist, so a payment call from a traced request with
// an idempotency id sends it neither traceparent, baggage nor X-Idempotency-ID.
func TestGatewayCallsCarryNoTelemetryHeaders(t *testing.T) {
	httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{}`))
	})
	shared := httpclienttest.NewClient(t, nil, "orders.internal")
	adapter := NewAdapter(Config{ApiEndpoint: server.URL},
		httpclient.NewHttpCaller(server.URL, shared, logging.NewLogger()))
	ctx, _ := httpclienttest.TracedContext(t)

	out := map[string]any{}
	_, err := adapter.post(corectx.NewRequestContext(ctx), "/v2/gateway/api/query", map[string]any{}, &out)

	require.NoError(t, err)
	server.AssertNoTelemetryHeaders(t)
}
