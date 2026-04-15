package httpserver

import (
	"fmt"

	"github.com/labstack/echo/v5"

	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func CustomHttpErrorHandler(defaultHandler echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(ctx *echo.Context, err error) {
		msg := fmt.Sprintf("Error from endpoint: %s %s", ctx.Request().Method, ctx.Request().URL.Path)
		logging.Logger().Error(msg, err)

		defaultHandler(ctx, err)
	}
}
