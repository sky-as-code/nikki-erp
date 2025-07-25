package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
)

type actionRestParams struct {
	dig.In

	ActionSvc it.ActionService
}

func NewActionRest(params actionRestParams) *ActionRest {
	return &ActionRest{
		ActionSvc: params.ActionSvc,
	}
}

type ActionRest struct {
	httpserver.RestBase
	ActionSvc it.ActionService
}

func (this ActionRest) CreateAction(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create action"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ActionSvc.CreateAction,
		func(request CreateActionRequest) it.CreateActionCommand {
			return it.CreateActionCommand(request)
		},
		func(result it.CreateActionResult) CreateActionResponse {
			response := CreateActionResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this ActionRest) UpdateAction(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST update action"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ActionSvc.UpdateAction,
		func(request UpdateActionRequest) it.UpdateActionCommand {
			return it.UpdateActionCommand(request)
		},
		func(result it.UpdateActionResult) UpdateActionResponse {
			response := UpdateActionResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this ActionRest) DeleteHardAction(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST delete hard action"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ActionSvc.DeleteHardAction,
		func(request DeleteHardActionRequest) it.DeleteHardActionCommand {
			return it.DeleteHardActionCommand(request)
		},
		func(result it.DeleteHardActionResult) DeleteHardActionResponse {
			response := DeleteHardActionResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this ActionRest) GetActionById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get action by id"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ActionSvc.GetActionById,
		func(request GetActionByIdRequest) it.GetActionByIdQuery {
			return it.GetActionByIdQuery(request)
		},
		func(result it.GetActionByIdResult) GetActionByIdResponse {
			response := GetActionByIdResponse{}
			response.FromAction(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this ActionRest) SearchActions(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST search actions"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ActionSvc.SearchActions,
		func(request SearchActionsRequest) it.SearchActionsCommand {
			return it.SearchActionsCommand(request)
		},
		func(result it.SearchActionsResult) SearchActionsResponse {
			response := SearchActionsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}
