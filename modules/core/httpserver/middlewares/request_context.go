package middlewares

import (
	"github.com/labstack/echo/v5"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// RequestContextMiddleware opens the request context and labels it as an HTTP execution. It mints
// no id: the trace id (X-Trace-ID) is what correlates a request's audit entries and log lines.
func RequestContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := corectx.NewRequestContext(c.Request().Context())
		corectx.SetRequestMeta(ctx, corectx.NewRequestMeta(corectx.RequestSourceHttp))
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
