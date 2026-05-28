package v1

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
)

type actionRestParams struct {
	dig.In

	ActionSvc it.ActionAppService
}

func NewActionRest(params actionRestParams) *ActionRest {
	return &ActionRest{ActionSvc: params.ActionSvc}
}

type ActionRest struct {
	httpserver.RestBase
	ActionSvc it.ActionAppService
}

func (this ActionRest) CreateAction(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create action"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2[CreateActionRequest, CreateActionResponse, it.CreateActionCommand, domain.Action](
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
	return httpserver.ServeGeneralMutate[DeleteActionRequest, DeleteActionResponse](
		"delete action",
		echoCtx,
		this.ActionSvc.DeleteAction,
	)
}

func (this ActionRest) GetAction(echoCtx *echo.Context) (err error) {
	return httpserver.ServeGetOne2[GetActionRequest, GetActionResponse, domain.Action](
		"get action",
		echoCtx,
		this.ActionSvc.GetAction,
	)
}

func (this ActionRest) ActionExists(echoCtx *echo.Context) (err error) {
	return httpserver.ServeRequest2[ActionExistsRequest, ActionExistsResponse, it.ActionExistsQuery, dyn.ExistsResultData](
		echoCtx,
		this.ActionSvc.ActionExists,
		func(request ActionExistsRequest) it.ActionExistsQuery {
			return request
		},
		func(data dyn.ExistsResultData) ActionExistsResponse {
			return ActionExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this ActionRest) SearchActions(echoCtx *echo.Context) (err error) {
	return httpserver.ServeSearch[SearchActionsRequest, SearchActionsResponse, domain.Action](
		"search actions",
		echoCtx,
		this.ActionSvc.SearchActions,
	)
}

func (this ActionRest) UpdateAction(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update action"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2[UpdateActionRequest, UpdateActionResponse, it.UpdateActionCommand, dyn.MutateResultData](
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

/*
 * Non-CRUD APIs
 */

func (this ActionRest) GetModelSchema(echoCtx *echo.Context) (err error) {
	schema := dmodel.MustGetSchema(domain.ActionSchemaName)
	echoCtx.JSON(http.StatusOK, schema.ToSimplized())
	return nil
}
