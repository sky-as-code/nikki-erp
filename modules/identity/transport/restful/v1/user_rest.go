package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"

	// c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type userRestParams struct {
	dig.In

	Config  config.ConfigService
	Logger  logging.LoggerService
	CqrsBus cqrs.CqrsBus
}

func NewUserRest(params userRestParams) *UserRest {
	return &UserRest{
		ConfigSvc: params.Config,
		Logger:    params.Logger,
		CqrsBus:   params.CqrsBus,
	}
}

type UserRest struct {
	ConfigSvc config.ConfigService
	Logger    logging.LoggerService
	CqrsBus   cqrs.CqrsBus
}

func (this UserRest) CreateUser(echoCtx echo.Context) (err error) {
	request := &CreateUserRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.CreateUserResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	// response := PrepareFileAccessResponse{
	// 	Token: result.Token,
	// 	Ttl:   result.ExpiresAt.Unix(),
	// 	Url:   result.CoolUrl,
	// }
	// h.logger.Infof("[PrepareFileAccess] response: %v", response)
	// echoCtx.Response().Header().Set("Content-Type", "application/json")
	return echoCtx.JSON(http.StatusOK, CreateUserResponse{
		Data:   result,
		Errors: result.Errors,
	})
}
