package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/action"
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

func (this ActionRest) DeleteActionHard(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST delete action hard"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.ActionSvc.DeleteActionHard,
		func(request DeleteActionHardByIdRequest) it.DeleteActionHardByIdCommand {
			return it.DeleteActionHardByIdCommand(request)
		},
		func(result it.DeleteActionHardByIdResult) DeleteActionHardByIdResponse {
			response := DeleteActionHardByIdResponse{}
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
		func(request SearchActionsRequest) it.SearchActionsQuery {
			return it.SearchActionsQuery(request)
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
