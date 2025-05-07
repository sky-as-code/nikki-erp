package middleware

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
)

type contextKey struct {
	name string
}

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var jwtCtxKey = &contextKey{"CaptureBearerToken"}

func CaptureBearerToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("authorization")
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) < 2 {
			return next(c)
		}

		jwt := splitToken[1]
		ctx := context.WithValue(c.Request().Context(), jwtCtxKey, jwt)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func JwtFromContext(ctx context.Context) string {
	raw, _ := ctx.Value(jwtCtxKey).(string)
	return raw
}
