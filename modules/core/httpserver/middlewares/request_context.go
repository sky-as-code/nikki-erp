package middlewares

import (
	"github.com/labstack/echo/v4"
	"go.bryk.io/pkg/errors"

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
			ctx := corectx.NewRequestContextF(c.Request().Context(), moduleName, nil)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// Returns pointer to an instance of RequestContext if it exists, otherwise returns an error.
func AsRequestContext(echoCtx echo.Context) (corectx.Context, error) {
	reqCtx, isReqCtx := echoCtx.Request().Context().(corectx.Context)
	if !isReqCtx {
		return nil, errors.New("Must have RequestContextMiddleware2 before this")
	}
	return reqCtx, nil
}
