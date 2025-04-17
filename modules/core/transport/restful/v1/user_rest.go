package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/config"
	// c "github.com/sky-as-code/nikki-erp/common/constants"
	"github.com/sky-as-code/nikki-erp/common/logging"
)

type userRestParams struct {
	dig.In

	Config config.ConfigService
	Logger logging.LoggerService
}

func NewUserRest(params userRestParams) *UserRest {
	return &UserRest{
		ConfigSvc: params.Config,
		Logger:    params.Logger,
	}
}

type UserRest struct {
	ConfigSvc config.ConfigService
	Logger    logging.LoggerService
}

func (h UserRest) CreateUser(echoCtx echo.Context) (err error) {
	// request := &PrepareFileAccessRequest{}
	// if err = echoCtx.Bind(request); err != nil {
	// 	return err
	// }

	// result, fileErr := h.wopiSvc.PrepareFileAccess(request.FileId, "")

	// if fileErr != nil {
	// 	return fileErr
	// }

	// response := PrepareFileAccessResponse{
	// 	Token: result.Token,
	// 	Ttl:   result.ExpiresAt.Unix(),
	// 	Url:   result.CoolUrl,
	// }
	// h.logger.Infof("[PrepareFileAccess] response: %v", response)
	// echoCtx.Response().Header().Set("Content-Type", "application/json")
	return echoCtx.String(http.StatusOK, "OK")
}
