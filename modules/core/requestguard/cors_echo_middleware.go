package requestguard

import (
	// "net/http"

	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// An Echo middleware that supports re-configuring CORS at runtime
func CorsEchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		var guardSvc RequestGuardService
		ft.PanicOnErr(deps.Invoke(func(guard RequestGuardService) {
			guardSvc = guard
		}))
		return func(c echo.Context) error {
			corsMiddleware, err := guardSvc.GetCorsMiddleware()
			if err != nil {
				return err
			}
			// CORS is disabled (for service-to-service calls)
			if corsMiddleware == nil {
				return next(c)
			}
			return corsMiddleware(next)(c)
		}
	}
}
