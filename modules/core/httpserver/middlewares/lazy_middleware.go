package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func Lazyware(middlewareCreator func() echo.MiddlewareFunc) *LazyMiddleware {
	return &LazyMiddleware{
		middlewareCreator: middlewareCreator,
		middlewareFn:      nil,
	}
}

type LazyMiddleware struct {
	middlewareCreator func() echo.MiddlewareFunc
	middlewareFn      echo.MiddlewareFunc
}

func (this *LazyMiddleware) Enable() {
	this.middlewareFn = this.middlewareCreator()
}

func (this *LazyMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			if this.middlewareFn == nil {
				return errors.Errorf("middleware is not enabled for creator: %T", this.middlewareCreator)
			}
			return this.middlewareFn(next)(echoCtx)
		}
	}
}
