package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
)

func RequestContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := crud.NewRequestContext(c.Request().Context())
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func RequestContextMiddleware2(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := dynamicentity.NewRequestContext(c.Request().Context())
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
