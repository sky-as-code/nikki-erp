package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/role_suite"
)

type roleSuiteRestParams struct {
	dig.In

	RoleSuiteService it.RoleSuiteService
}

func NewRoleSuiteRest(params roleSuiteRestParams) *RoleSuiteRest {
	return &RoleSuiteRest{
		roleSuiteService: params.RoleSuiteService,
	}
}

type RoleSuiteRest struct {
	roleSuiteService it.RoleSuiteService
}

func (this RoleSuiteRest) CreateRoleSuite(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create role suite"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleSuiteService.CreateRoleSuite,
		func(request CreateRoleSuiteRequest) it.CreateRoleSuiteCommand {
			return it.CreateRoleSuiteCommand(request)
		},
		func(result it.CreateRoleSuiteResult) CreateRoleSuiteResponse {
			response := CreateRoleSuiteResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this RoleSuiteRest) UpdateRoleSuite(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST update role suite"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleSuiteService.UpdateRoleSuite,
		func(request UpdateRoleSuiteRequest) it.UpdateRoleSuiteCommand {
			return it.UpdateRoleSuiteCommand(request)
		},
		func(result it.UpdateRoleSuiteResult) UpdateRoleSuiteResponse {
			response := UpdateRoleSuiteResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this RoleSuiteRest) DeleteRoleSuite(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST delete role suite"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleSuiteService.DeleteHardRoleSuite,
		func(request DeleteRoleSuiteRequest) it.DeleteRoleSuiteCommand {
			return it.DeleteRoleSuiteCommand(request)
		},
		func(result it.DeleteRoleSuiteResult) DeleteRoleSuiteResponse {
			response := DeleteRoleSuiteResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this RoleSuiteRest) GetRoleSuiteById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get role suite by id"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleSuiteService.GetRoleSuiteById,
		func(request GetRoleSuiteByIdRequest) it.GetRoleSuiteByIdQuery {
			return it.GetRoleSuiteByIdQuery(request)
		},
		func(result it.GetRoleSuiteByIdResult) GetRoleSuiteByIdResponse {
			response := GetRoleSuiteByIdResponse{}
			response.FromRoleSuite(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this RoleSuiteRest) SearchRoleSuites(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST search role suites"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleSuiteService.SearchRoleSuites,
		func(request SearchRoleSuitesRequest) it.SearchRoleSuitesCommand {
			return it.SearchRoleSuitesCommand(request)
		},
		func(result it.SearchRoleSuitesResult) SearchRoleSuitesResponse {
			response := SearchRoleSuitesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}
