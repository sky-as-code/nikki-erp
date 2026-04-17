package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

type moduleRestParams struct {
	dig.In

	ModuleSvc it.ModuleService
}

func NewModuleRest(params moduleRestParams) *ModuleRest {
	return &ModuleRest{moduleSvc: params.ModuleSvc}
}

type ModuleRest struct {
	moduleSvc it.ModuleService
}

func (this ModuleRest) CreateModule(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create module metadata"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.moduleSvc.CreateModule,
		func(request CreateModuleRequest) it.CreateModuleCommand {
			cmd := it.CreateModuleCommand{}
			cmd.SetFieldData(request.DynamicFields)
			return cmd
		},
		func(data domain.ModuleMetadata) CreateModuleResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated,
	)
}

func (this ModuleRest) DeleteModule(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete module metadata"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.moduleSvc.DeleteModule,
		func(request DeleteModuleRequest) it.DeleteModuleCommand {
			return it.DeleteModuleCommand(request)
		},
		func(data dyn.MutateResultData) DeleteModuleResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this ModuleRest) GetModule(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get module metadata"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.moduleSvc.GetModule,
		func(request GetModuleRequest) it.GetModuleQuery {
			return it.GetModuleQuery(request)
		},
		func(data domain.ModuleMetadata) GetModuleResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this ModuleRest) ModuleExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST module metadata exists"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.moduleSvc.ModuleExists,
		func(request ModuleExistsRequest) it.ModuleExistsQuery {
			return it.ModuleExistsQuery(request)
		},
		func(data dyn.ExistsResultData) ModuleExistsResponse {
			return ModuleExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this ModuleRest) SearchModules(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search module metadata"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.moduleSvc.SearchModules,
		func(request SearchModulesRequest) it.SearchModulesQuery {
			return it.SearchModulesQuery(request)
		},
		func(data it.SearchModulesResultData) SearchModulesResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
	)
}

func (this ModuleRest) UpdateModule(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update module metadata"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.moduleSvc.UpdateModule,
		func(request UpdateModuleRequest) it.UpdateModuleCommand {
			cmd := it.UpdateModuleCommand{}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.Id)))
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
