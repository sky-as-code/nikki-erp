package v1

import (
	"github.com/labstack/echo/v4"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
)

type actionRestParams struct {
	dig.In

	Config  config.ConfigService
	Logger  logging.LoggerService
	CqrsBus cqrs.CqrsBus
}

func NewActionRest(params actionRestParams) *ActionRest {
	return &ActionRest{
		RestBase: httpserver.RestBase{
			ConfigSvc: params.Config,
			Logger:    params.Logger,
			CqrsBus:   params.CqrsBus,
		},
	}
}

type ActionRest struct {
	httpserver.RestBase
}

func (this ActionRest) CreateAction(echoCtx echo.Context) (err error) {
	request := &CreateActionRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.CreateActionResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := CreateActionResponse{}
	response.FromAction(*result.Data)

	return httpserver.JsonCreated(echoCtx, response)
}

func (this ActionRest) UpdateAction(echoCtx echo.Context) (err error) {
	request := &UpdateActionRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.UpdateActionResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := UpdateActionResponse{}
	response.FromAction(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this ActionRest) GetActionById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get action by id"); e != nil {
			err = e
		}
	}()

	request := &GetActionByIdRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.GetActionByIdResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := GetActionByIdResponse{}
	response.FromAction(*result.Data)

	return httpserver.JsonOk(echoCtx, response)
}

func (this ActionRest) SearchActions(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list actions"); e != nil {
			err = e
		}
	}()

	request := &SearchActionsRequest{}
	if err = echoCtx.Bind(request); err != nil {
		return err
	}

	result := it.SearchActionsResult{}
	err = this.CqrsBus.Request(echoCtx.Request().Context(), *request, &result)

	if err != nil {
		return err
	}

	if result.ClientError != nil {
		return httpserver.JsonBadRequest(echoCtx, result.ClientError)
	}

	response := SearchActionsResponse{}
	response.FromResultWithResources(result.Data)

	return httpserver.JsonOk(echoCtx, response)
}
