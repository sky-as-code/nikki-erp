package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// An Echo middleware that supports re-configuring CORS at runtime.
// Only takes effect if enabled in the configuration.
func CorsEchoMiddleware() echo.MiddlewareFunc {
	return createMiddlewareFunc(func(c *echo.Context, guardSvc requestguard.RequestGuardService, next echo.HandlerFunc) error {
		reqCtx, err := corectx.AsRequestContext(c)
		if err != nil {
			return err
		}
		corsMiddleware, err := guardSvc.GetCorsMiddleware(reqCtx)
		if err != nil {
			return err
		}
		// If CORS is disabled (for service-to-service calls)
		if corsMiddleware == nil {
			return next(c)
		}
		return corsMiddleware(next)(c)
	})
}

// Verify mTLS connection.
// Only takes effect if enabled in the configuration.
func TrustedConnectionMiddleware() echo.MiddlewareFunc {
	return createMiddlewareFunc(func(echoCtx *echo.Context, guardSvc requestguard.RequestGuardService, next echo.HandlerFunc) error {
		reqCtx, err := corectx.AsRequestContext(echoCtx)
		if err != nil {
			return err
		}
		result, err := guardSvc.VerifyTrustedConnection(reqCtx, echoCtx.Request())
		if err != nil {
			return err
		}
		if !result.IsOk {
			return echoCtx.JSON(http.StatusForbidden, result.ClientError)
		}
		return next(echoCtx)
	})
}

type handlerFn func(c *echo.Context, guardSvc requestguard.RequestGuardService, next echo.HandlerFunc) error

func createMiddlewareFunc(handle handlerFn) echo.MiddlewareFunc {
	var guardSvc requestguard.RequestGuardService
	ft.PanicOnErr(deps.Invoke(func(guard requestguard.RequestGuardService) {
		guardSvc = guard
	}))
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return handle(c, guardSvc, next)
		}
	}
}
