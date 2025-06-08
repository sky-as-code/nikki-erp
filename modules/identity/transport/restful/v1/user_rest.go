package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
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
		RestBase: httpserver.RestBase{
			ConfigSvc: params.Config,
			Logger:    params.Logger,
			CqrsBus:   params.CqrsBus,
		},
	}
}

type UserRest struct {
	httpserver.RestBase
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

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := CreateUserResponse{}
	response.FromUser(*result.Data)

	return httpserver.JsonCreated(echoCtx, response)
}

func (this UserRest) UpdateUser(echoCtx echo.Context) (err error) {
	request := &UpdateUserRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.UpdateUserResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := UpdateUserResponse{}
	response.FromUser(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this UserRest) DeleteUser(echoCtx echo.Context) (err error) {
	request := &DeleteUserRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.DeleteUserResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := DeleteUserResponse{
		DeletedAt: result.Data.DeletedAt.Unix(),
	}

	return httpserver.JsonOk(echoCtx, response)
}

func (this UserRest) GetUserById(echoCtx echo.Context) (err error) {
	request := &GetUserByIdRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.GetUserByIdResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := GetUserByIdResponse{}
	response.FromUser(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this UserRest) SearchUsers(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to search users"); e != nil {
			err = e
		}
	}()

	request := &SearchUsersRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.SearchUsersResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := SearchUsersResponse{}
	response.FromResult(result.Data)

	return httpserver.JsonOk(echoCtx, response)
}
