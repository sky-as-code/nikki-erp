package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ID-01: the backend mints no request id of its own; the trace id is the only correlation id.
func TestRequestContext_SendsNoRequestIdHeader(t *testing.T) {
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	recorder := fixture.serve("/items/1", nil)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, recorder.Header().Get("X-Request-Id"))
	assert.NotEmpty(t, recorder.Header().Get(HeaderTraceId))
}

func TestRequestContext_LabelsTheRequestAsHttp(t *testing.T) {
	server := echo.New()
	server.Use(RequestContextMiddleware)
	var seen corectx.RequestMeta
	server.GET("/", func(echoCtx *echo.Context) error {
		seen, _ = corectx.RequestMetaOf(echoCtx.Request().Context())
		return echoCtx.NoContent(http.StatusNoContent)
	})

	server.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, corectx.RequestMeta{Source: corectx.RequestSourceHttp}, seen)
}
