package httpserver

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type RestBase struct {
	ConfigSvc config.ConfigService
	Logger    logging.LoggerService
	CqrsBus   cqrs.CqrsBus
}

func JsonCreated(echoCtx echo.Context, data any) error {
	return echoCtx.JSON(http.StatusCreated, data)
}

func JsonOk(echoCtx echo.Context, data any) error {
	return echoCtx.JSON(http.StatusOK, data)
}

func JsonBadRequest(echoCtx echo.Context, err any) error {
	return echoCtx.JSON(http.StatusBadRequest, err)
}
