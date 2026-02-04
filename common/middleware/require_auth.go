package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RequireAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			jwtToken := JwtFromContext(c.Request().Context())
			userId := GetUserIdFromContext(c.Request().Context())
			if jwtToken == "" || userId == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			}
			return next(c)
		}
	}
}
