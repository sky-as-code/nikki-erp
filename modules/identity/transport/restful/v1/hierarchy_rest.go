package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
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
	err = httpserver.ServeRequest(
		echoCtx, this.HierarchySvc.CreateHierarchyLevel,
		func(request CreateHierarchyLevelRequest) it.CreateHierarchyLevelCommand {
			return it.CreateHierarchyLevelCommand(request)
		},
		func(result it.CreateHierarchyLevelResult) CreateHierarchyLevelResponse {
			response := CreateHierarchyLevelResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)
	return err
}

func (this HierarchyRest) UpdateHierarchyLevel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update hierarchy level"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.HierarchySvc.UpdateHierarchyLevel,
		func(request UpdateHierarchyLevelRequest) it.UpdateHierarchyLevelCommand {
			return it.UpdateHierarchyLevelCommand(request)
		},
		func(result it.UpdateHierarchyLevelResult) UpdateHierarchyLevelResponse {
			response := UpdateHierarchyLevelResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this HierarchyRest) DeleteHierarchyLevel(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete hierarchy level"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.HierarchySvc.DeleteHierarchyLevel,
		func(request DeleteHierarchyLevelRequest) it.DeleteHierarchyLevelCommand {
			return it.DeleteHierarchyLevelCommand(request)
		},
		func(result it.DeleteHierarchyLevelResult) DeleteHierarchyLevelResponse {
			response := DeleteHierarchyLevelResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this HierarchyRest) GetHierarchyLevelById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get hierarchy level by id"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.HierarchySvc.GetHierarchyLevelById,
		func(request GetHierarchyLevelByIdRequest) it.GetHierarchyLevelByIdQuery {
			return it.GetHierarchyLevelByIdQuery(request)
		},
		func(result it.GetHierarchyLevelByIdResult) GetHierarchyLevelByIdResponse {
			response := GetHierarchyLevelByIdResponse{}
			response.FromHierarchyLevel(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this HierarchyRest) SearchHierarchyLevels(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search hierarchy levels"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.HierarchySvc.SearchHierarchyLevels,
		func(request SearchHierarchyLevelsRequest) it.SearchHierarchyLevelsQuery {
			return it.SearchHierarchyLevelsQuery(request)
		},
		func(result it.SearchHierarchyLevelsResult) SearchHierarchyLevelsResponse {
			response := SearchHierarchyLevelsResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}

func (this HierarchyRest) ManageHierarchyUsers(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST manage hierarchy users"); e != nil {
			err = e
		}
	}()
	err = httpserver.ServeRequest(
		echoCtx, this.HierarchySvc.AddRemoveUsers,
		func(request ManageUsersHierarchyRequest) it.AddRemoveUsersCommand {
			return it.AddRemoveUsersCommand(request)
		},
		func(result it.AddRemoveUsersResult) ManageUsersHierarchyResponse {
			response := ManageUsersHierarchyResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)
	return err
}
