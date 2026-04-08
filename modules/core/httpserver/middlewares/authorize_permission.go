package middlewares

import (
	"github.com/labstack/echo/v4"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	// corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	ext "github.com/sky-as-code/nikki-erp/modules/core/httpserver/external"
)

func AuthorizePermissionMiddleware() echo.MiddlewareFunc {
	deps.Invoke(func(permissionSvc ext.PermissionExtService) {
		// Self-check dependency on startup
	})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			// deps.Invoke(func(permissionSvc ext.PermissionExtService) {
			// 	ctx := echoCtx.Request().Context().(corectx.Context)
			// 	userId := GetUserIdFromContext(echoCtx.Request().Context())
			// 	permissionSvc.IsAuthorized(echoCtx.Request().Context(), itPerm.IsAuthorizedQuery{
			// 		UserId: userId,
			// 	})
			// })
			// jwtToken := JwtFromContext(c.Request().Context())
			// userId := GetUserIdFromContext(c.Request().Context())
			// if jwtToken == "" || userId == "" {
			// 	return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			// }
			return next(echoCtx)
		}
	}
}
