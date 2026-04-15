package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
)

type actionRestParams struct {
	dig.In

	ActionSvc it.ActionService
}

func NewActionRest(params actionRestParams) *ActionRest {
	return &ActionRest{ActionSvc: params.ActionSvc}
}

type ActionRest struct {
	httpserver.RestBase
	ActionSvc it.ActionService
}

func (this ActionRest) CreateAction(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create action"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.ActionSvc.CreateAction,
		func(request CreateActionRequest) it.CreateActionCommand {
			cmd := it.CreateActionCommand{}
			cmd.SetFieldData(request.DynamicFields)
			cmd.Action.SetResourceId(util.ToPtr(model.Id(request.ResourceId)))
			return cmd
		},
		func(data domain.Action) CreateActionResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this ActionRest) DeleteAction(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete action"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ActionSvc.DeleteAction,
		func(request DeleteActionRequest) it.DeleteActionCommand {
			return it.DeleteActionCommand(request)
		},
		func(data dyn.MutateResultData) DeleteActionResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this ActionRest) GetAction(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get action"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ActionSvc.GetAction,
		func(request GetActionRequest) it.GetActionQuery {
			return it.GetActionQuery(request)
		},
		func(data domain.Action) GetActionResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this ActionRest) ActionExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST action exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ActionSvc.ActionExists,
		func(request ActionExistsRequest) it.ActionExistsQuery {
			return it.ActionExistsQuery(request)
		},
		func(data dyn.ExistsResultData) ActionExistsResponse {
			return ActionExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this ActionRest) SearchActions(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search actions"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ActionSvc.SearchActions,
		func(request SearchActionsRequest) it.SearchActionsQuery {
			return it.SearchActionsQuery(request)
		},
		func(data it.SearchActionsResultData) SearchActionsResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this ActionRest) UpdateAction(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update action"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.ActionSvc.UpdateAction,
		func(request UpdateActionRequest) it.UpdateActionCommand {
			cmd := it.UpdateActionCommand{}
			cmd.SetFieldData(request.DynamicFields)
			cmd.Action.SetResourceId(util.ToPtr(model.Id(request.ResourceId)))
			cmd.Action.SetId(util.ToPtr(model.Id(request.ActionId)))
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
