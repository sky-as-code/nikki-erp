package middleware

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sky-as-code/nikki-erp/common/util"
)

type contextKey struct {
	name string
}

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var jwtCtxKey = &contextKey{"CaptureBearerToken"}
var userIdCtxKey = &contextKey{"UserId"}
var deviceIdCtxKey = &contextKey{"DeviceId"}
var rolesCtxKey = &contextKey{"Roles"}

// CaptureBearerToken captures and parses JWT token, sets user info to context
func CaptureBearerToken(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			splitToken := strings.Split(authHeader, "Bearer ")
			if len(splitToken) < 2 {
				return next(c)
			}

			tokenString := splitToken[1]
			if tokenString == "" {
				return next(c)
			}

			// Parse JWT token
			payload, err := util.ParseGJWToken(tokenString, secretKey)
			if err != nil {
				return next(c)
			}

			// Set token and parsed data to context
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, jwtCtxKey, tokenString)
			ctx = context.WithValue(ctx, userIdCtxKey, payload.UserId)
			ctx = context.WithValue(ctx, deviceIdCtxKey, payload.DId)
			ctx = context.WithValue(ctx, rolesCtxKey, payload.Roles)

			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func JwtFromContext(ctx context.Context) string {
	raw, _ := ctx.Value(jwtCtxKey).(string)
	return raw
}

func GetUserIdFromContext(ctx context.Context) string {
	raw, _ := ctx.Value(userIdCtxKey).(string)
	return raw
}

func GetDeviceIdFromContext(ctx context.Context) string {
	raw, _ := ctx.Value(deviceIdCtxKey).(string)
	return raw
}

func GetRolesFromContext(ctx context.Context) []string {
	raw, _ := ctx.Value(rolesCtxKey).([]string)
	return raw
}
