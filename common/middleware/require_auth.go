package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/sky-as-code/nikki-erp/common/fault"
)

func RequireAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			jwtToken := JwtFromContext(c.Request().Context())
			userId := GetUserIdFromContext(c.Request().Context())
			if jwtToken == "" || userId == "" {
				return &fault.ClientError{
					Code:    "403",
					Details: "Token required or invalid",
				}
			}
			return next(c)
		}
	}
}
