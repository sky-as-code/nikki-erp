package middlewares

import (
	"github.com/labstack/echo/v5"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

func RequestContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := corectx.NewRequestContext(c.Request().Context())
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
