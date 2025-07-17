package v1

import (
	"github.com/labstack/echo/v4"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"go.uber.org/dig"
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
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create role suite"); e != nil {
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

func (this RoleSuiteRest) GetRoleSuiteById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get role suite by id"); e != nil {
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
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search role suites"); e != nil {
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
