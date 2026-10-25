package middlewares

import (
	"github.com/labstack/echo/v5"
)

// Shortcut for lazy-wrapped ServiceAuthorizeMiddleware.
func ServiceAuthz() echo.MiddlewareFunc {
	return ServiceAuthorizeMiddleware()
}

// Placeholder for routes called by other services with a service token in the `Authorization`
// header. It passes every request through until service tokens are issued; the verification it
// will rely on is already RequestGuardService.VerifyServiceAccessToken.
func ServiceAuthorizeMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return next
	}
}
