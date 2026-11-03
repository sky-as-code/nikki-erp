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
		corsMiddleware, err := guardSvc.GetCorsMiddleware(corsRequestContext(c))
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

// corsRequestContext tolerates a request the chain rejected before RequestContextMiddleware ran
// (the idempotency 400 answers through CORS), lending the guard a context it does not keep.
func corsRequestContext(echoCtx *echo.Context) corectx.Context {
	if reqCtx, err := corectx.AsRequestContext(echoCtx); err == nil {
		return reqCtx
	}
	return corectx.NewRequestContext(echoCtx.Request().Context())
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
