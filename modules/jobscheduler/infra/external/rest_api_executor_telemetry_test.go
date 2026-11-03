package external

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/httpclienttest"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

// An action's configured headers cannot forge propagation: copies of the telemetry headers are
// stripped, and an allowlisted receiver gets the run's own id and trace instead. The scheduler's
// own business headers are untouched.
func TestConfiguredTelemetryHeadersAreReplacedByTheRunsOwn(t *testing.T) {
	httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, nil)
	executor := NewRestApiExecutor(httpclienttest.NewClient(t, nil, server.HostPort(t)))
	ctx, id := httpclienttest.TracedContext(t)

	outcome := executor.Execute(ctx, testInput(map[string]any{
		"method": "GET",
		"url":    server.URL + "/run",
		"headers": map[string]any{
			telemetry.HeaderIdempotencyId: "0190a1b2-c3d4-7e5f-8a6b-7c8d9e0f1a2b",
			"traceparent":                 "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
		},
	}))

	require.Empty(t, outcome.ErrorCode)
	received := server.Received()
	require.Len(t, received, 1)
	assert.Equal(t, id, received[0].Get(telemetry.HeaderIdempotencyId))
	assert.NotContains(t, received[0].Get("traceparent"), "0af7651916cd43dd8448eb211c80319c")
	assert.Equal(t, "inventory:rebuild:2026-08-20T10:00:00Z", received[0].Get(HeaderIdempotencyKey))
}

func TestNonAllowlistedActionUrlReceivesNoTelemetryHeaders(t *testing.T) {
	httpclienttest.InstallTelemetry(t)
	server := httpclienttest.NewHeaderServer(t, nil)
	executor := NewRestApiExecutor(httpclienttest.NewClient(t, nil))
	ctx, _ := httpclienttest.TracedContext(t)

	outcome := executor.Execute(ctx, testInput(map[string]any{"method": "POST", "url": server.URL + "/run"}))

	require.Empty(t, outcome.ErrorCode)
	server.AssertNoTelemetryHeaders(t)
}
