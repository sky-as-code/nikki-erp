package middlewares

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// An Echo middleware that supports re-configuring CORS at runtime.
// Only takes effect if enabled in the configuration.
func CorsEchoMiddleware() echo.MiddlewareFunc {
	return createMiddlewareFunc(func(c *echo.Context, guardSvc reguard.RequestGuardService, next echo.HandlerFunc) error {
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

type handlerFn func(c *echo.Context, guardSvc reguard.RequestGuardService, next echo.HandlerFunc) error

func createMiddlewareFunc(handle handlerFn) echo.MiddlewareFunc {
	var guardSvc reguard.RequestGuardService
	ft.PanicOnErr(deps.Invoke(func(guard reguard.RequestGuardService) {
		guardSvc = guard
	}))
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return handle(c, guardSvc, next)
		}
	}
}
