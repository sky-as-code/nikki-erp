package httpserver

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func CustomHttpErrorHandler(defaultHandler echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, ctx echo.Context) {
		if ctx.Response().Committed {
			// Response already sent
			return
		}
		msg := fmt.Sprintf("Error from endpoint: %s %s", ctx.Request().Method, ctx.Request().URL.Path)
		logging.Logger().Error(msg, err)

		defaultHandler(err, ctx)
	}
}
