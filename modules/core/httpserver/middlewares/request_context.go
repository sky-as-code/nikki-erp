package middlewares

import (
	"github.com/labstack/echo/v4"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func RequestContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := crud.NewRequestContext(c.Request().Context())
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func RequestContextMiddleware2(moduleName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := corectx.NewRequestContextM(c.Request().Context(), moduleName)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func RequestContextMiddleware3(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := corectx.NewRequestContext(c.Request().Context())
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
