package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const (
	corsOrigin        = "https://app.example.com"
	corsAllowHeaders  = "Accept,Authorization,Content-Type,traceparent,tracestate,baggage,X-Idempotency-ID"
	corsExposeHeaders = "X-Trace-ID,X-Idempotency-ID"
)

// newCorsIngressServer mirrors the server chain: CORS sits after RequestContext and before the
// global lazywares, and the idempotency 400 is answered through the same CORS middleware.
func newCorsIngressServer(t *testing.T) (*echo.Echo, *recordingLogger) {
	t.Helper()
	fixture := newIngressFixture(t, sdktrace.AlwaysSample())
	cors := middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{corsOrigin},
		AllowHeaders:  strings.Split(corsAllowHeaders, ","),
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		ExposeHeaders: strings.Split(corsExposeHeaders, ","),
	})
	server := echo.New()
	server.Use(TracingMiddleware, AccessLogMiddleware(fixture.logger), IdempotencyMiddleware(fixture.logger, cors),
		RecoverMiddleware(fixture.logger), RequestContextMiddleware, cors, rejectUnknownTenant)
	server.GET("/items/:id", func(echoCtx *echo.Context) error {
		return echoCtx.JSON(http.StatusOK, map[string]string{"ok": "yes"})
	})
	return server, fixture.logger
}

// rejectUnknownTenant stands in for the tenant lazyware, which fails with an error.
func rejectUnknownTenant(next echo.HandlerFunc) echo.HandlerFunc {
	return func(echoCtx *echo.Context) error {
		if echoCtx.Request().Header.Get("X-Test-Tenant") == "unknown" {
			return echo.NewHTTPError(http.StatusBadRequest, "tenant not found")
		}
		return next(echoCtx)
	}
}

func serveCors(server *echo.Echo, method string, header http.Header) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, "/items/1", nil)
	for name, values := range header {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	return recorder
}

func assertReadableCrossOrigin(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	assert.Equal(t, []string{corsOrigin}, recorder.Header().Values(echo.HeaderAccessControlAllowOrigin))
	assert.Equal(t, corsExposeHeaders, recorder.Header().Get(echo.HeaderAccessControlExposeHeaders))
}

// AC-17: a cross-origin request with an invalid idempotency id gets a 400 the browser can read.
func TestCors_InvalidIdempotencyIdIsReadableCrossOrigin(t *testing.T) {
	server, logger := newCorsIngressServer(t)
	recorder := serveCors(server, http.MethodGet, http.Header{
		echo.HeaderOrigin: {corsOrigin}, HeaderIdempotencyId: {"not-a-uuid"},
	})

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assertReadableCrossOrigin(t, recorder)
	assertInvalidIdempotencyBody(t, recorder.Body.Bytes())
	assert.Regexp(t, traceIdPattern, recorder.Header().Get(HeaderTraceId))
	require.Len(t, logger.find(accessLogMessage), 1)
}

func TestCors_TenantRejectionIsReadableCrossOrigin(t *testing.T) {
	server, _ := newCorsIngressServer(t)
	recorder := serveCors(server, http.MethodGet, http.Header{
		echo.HeaderOrigin: {corsOrigin}, "X-Test-Tenant": {"unknown"},
	})

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assertReadableCrossOrigin(t, recorder)
	assert.NotEmpty(t, recorder.Header().Get(HeaderIdempotencyId))
}

func TestCors_AcceptedRequestHasOneAllowOriginHeader(t *testing.T) {
	server, _ := newCorsIngressServer(t)
	recorder := serveCors(server, http.MethodGet, http.Header{echo.HeaderOrigin: {corsOrigin}})

	assert.Equal(t, http.StatusOK, recorder.Code)
	assertReadableCrossOrigin(t, recorder)
	assert.NotEmpty(t, recorder.Header().Get(HeaderIdempotencyId))
}

func TestCors_PreflightSucceedsWithTraceHeadersAllowed(t *testing.T) {
	server, logger := newCorsIngressServer(t)
	recorder := serveCors(server, http.MethodOptions, http.Header{
		echo.HeaderOrigin:                      {corsOrigin},
		echo.HeaderAccessControlRequestMethod:  {http.MethodPost},
		echo.HeaderAccessControlRequestHeaders: {"traceparent,X-Idempotency-ID"},
	})

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, []string{corsOrigin}, recorder.Header().Values(echo.HeaderAccessControlAllowOrigin))
	assert.Equal(t, corsAllowHeaders, recorder.Header().Get(echo.HeaderAccessControlAllowHeaders))
	assert.NotEmpty(t, recorder.Header().Get(HeaderIdempotencyId))
	require.Len(t, logger.find(accessLogMessage), 1)
}

func TestCors_SameOriginRequestGetsNoCorsHeaders(t *testing.T) {
	server, _ := newCorsIngressServer(t)
	invalid := serveCors(server, http.MethodGet, http.Header{HeaderIdempotencyId: {"not-a-uuid"}})
	accepted := serveCors(server, http.MethodGet, nil)

	assert.Equal(t, http.StatusBadRequest, invalid.Code)
	assert.Equal(t, http.StatusOK, accepted.Code)
	for _, recorder := range []*httptest.ResponseRecorder{invalid, accepted} {
		assert.Empty(t, recorder.Header().Values(echo.HeaderAccessControlAllowOrigin))
		assert.Empty(t, recorder.Header().Get(echo.HeaderAccessControlExposeHeaders))
	}
}

// The idempotency 400 reaches CORS before RequestContextMiddleware ran.
func TestCorsRequestContext_ToleratesMissingRequestContext(t *testing.T) {
	echoCtx := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	assert.NotNil(t, corsRequestContext(echoCtx))

	reqCtx := corectx.NewRequestContext(echoCtx.Request().Context())
	echoCtx.SetRequest(echoCtx.Request().WithContext(reqCtx))
	assert.Same(t, reqCtx, corsRequestContext(echoCtx))
}
