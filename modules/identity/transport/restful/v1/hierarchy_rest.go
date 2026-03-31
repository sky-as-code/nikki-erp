package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
)

type hierarchyRestParams struct {
	dig.In

	HierarchySvc it.HierarchyService
}

func NewHierarchyRest(params hierarchyRestParams) *HierarchyRest {
	return &HierarchyRest{
		HierarchySvc: params.HierarchySvc,
	}
}

type HierarchyRest struct {
	httpserver.RestBase
	HierarchySvc it.HierarchyService
}

func (this HierarchyRest) CreateHierarchyLevel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create hierarchy level"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.HierarchySvc.CreateHierarchyLevel,
		func(requestFields dmodel.DynamicFields) it.CreateHierarchyLevelCommand {
			cmd := it.CreateHierarchyLevelCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.HierarchyLevel) CreateHierarchyLevelResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this HierarchyRest) DeleteHierarchyLevel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete hierarchy level"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.HierarchySvc.DeleteHierarchyLevel,
		func(request DeleteHierarchyLevelRequest) it.DeleteHierarchyLevelCommand {
			return it.DeleteHierarchyLevelCommand(request)
		},
		func(data dyn.MutateResultData) DeleteHierarchyLevelResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this HierarchyRest) GetHierarchyLevel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get hierarchy level"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.HierarchySvc.GetHierarchyLevel,
		func(request GetHierarchyLevelRequest) it.GetHierarchyLevelQuery {
			return it.GetHierarchyLevelQuery(request)
		},
		func(data domain.HierarchyLevel) dmodel.DynamicFields {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this HierarchyRest) HierarchyLevelExists(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST hierarchy level exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.HierarchySvc.HierarchyLevelExists,
		func(request HierarchyLevelExistsRequest) it.HierarchyLevelExistsQuery {
			return it.HierarchyLevelExistsQuery(request)
		},
		func(data dyn.ExistsResultData) HierarchyLevelExistsResponse {
			return HierarchyLevelExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this HierarchyRest) ManageHierarchyUsers(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST manage hierarchy users"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.HierarchySvc.ManageHierarchyLevelUsers,
		func(request ManageHierarchyLevelUsersRequest) it.ManageHierarchyLevelUsersCommand {
			return it.ManageHierarchyLevelUsersCommand(request)
		},
		func(data dyn.MutateResultData) ManageHierarchyLevelUsersResponse {
			return httpserver.NewRestMutateResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this HierarchyRest) SearchHierarchyLevels(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search hierarchy levels"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.HierarchySvc.SearchHierarchyLevels,
		func(request SearchHierarchyLevelsRequest) it.SearchHierarchyLevelsQuery {
			return it.SearchHierarchyLevelsQuery(request)
		},
		func(data it.SearchHierarchyLevelsResultData) SearchHierarchyLevelsResponse {
			return httpserver.NewSearchUsersResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this HierarchyRest) UpdateHierarchyLevel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update hierarchy level"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.HierarchySvc.UpdateHierarchyLevel,
		func(requestFields dmodel.DynamicFields) it.UpdateHierarchyLevelCommand {
			cmd := it.UpdateHierarchyLevelCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data dyn.MutateResultData) UpdateHierarchyLevelResponse {
			return httpserver.NewRestUpdateResponse2(data)
		},
		httpserver.JsonOk,
	)
}
