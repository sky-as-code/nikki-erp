package middleware

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sky-as-code/nikki-erp/common/fault"
)

type PermissionChecker interface {
	CheckPermission(ctx context.Context, subjectRef, resourceName, actionName, scopeRef string) (allow bool, err error)
}

type ScopeRefExtractor func(c echo.Context) string

func DefaultScopeRefExtractor(c echo.Context) string {
	if s := c.QueryParam("scopeRef"); s != "" {
		return s
	}
	return ""
}

// Must run after RequireAuthMiddleware and RequestContextMiddleware (userId is assumed present).
func RequirePermission(
	checker PermissionChecker,
	resource string,
	action string,
	scopeRefFrom ScopeRefExtractor,
) echo.MiddlewareFunc {
	if scopeRefFrom == nil {
		scopeRefFrom = DefaultScopeRefExtractor
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userId := GetUserIdFromContext(c.Request().Context())
			scopeRef := scopeRefFrom(c)
			ctx := c.Request().Context()

			allow, err := checker.CheckPermission(ctx, userId, resource, action, scopeRef)
			fault.PanicOnErr(err)

			if !allow {
				return echo.NewHTTPError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
			}
			return next(c)
		}
	}
}
