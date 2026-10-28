package middlewares

import (
	"github.com/labstack/echo/v5"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// HeaderRequestId echoes the id minted for the request, so a client reporting a problem can name
// the audit entries and log lines it produced.
const HeaderRequestId = "X-Request-Id"

// RequestContextMiddleware mints one request id per HTTP request. A client-sent id is not taken
// over: it would let a caller graft its actions onto someone else's audit trail.
func RequestContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := corectx.NewRequestContext(c.Request().Context())
		meta := corectx.NewRequestMeta(corectx.RequestSourceHttp)
		corectx.SetRequestMeta(ctx, meta)
		c.Response().Header().Set(HeaderRequestId, meta.Id)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
