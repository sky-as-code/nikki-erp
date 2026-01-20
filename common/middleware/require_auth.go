package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RequireAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			jwt := JwtFromContext(c.Request().Context())
			if jwt == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized: Token required",
				})
			}
			return next(c)
		}
	}
}
