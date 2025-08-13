package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
)

type roleRestParams struct {
	dig.In

	RoleService it.RoleService
}

func NewRoleRest(params roleRestParams) *RoleRest {
	return &RoleRest{
		roleService: params.RoleService,
	}
}

type RoleRest struct {
	roleService it.RoleService
}

func (this RoleRest) CreateRole(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST create action"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleService.CreateRole,
		func(request CreateRoleRequest) it.CreateRoleCommand {
			return it.CreateRoleCommand(request)
		},
		func(result it.CreateRoleResult) CreateRoleResponse {
			response := CreateRoleResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonCreated,
	)

	return err
}

func (this RoleRest) UpdateRole(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST update role"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleService.UpdateRole,
		func(request UpdateRoleRequest) it.UpdateRoleCommand {
			return it.UpdateRoleCommand(request)
		},
		func(result it.UpdateRoleResult) UpdateRoleResponse {
			response := UpdateRoleResponse{}
			response.FromEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this RoleRest) DeleteRoleHard(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST delete role hard"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleService.DeleteRoleHard,
		func(request DeleteRoleHardRequest) it.DeleteRoleHardCommand {
			return it.DeleteRoleHardCommand(request)
		},
		func(result it.DeleteRoleHardResult) DeleteRoleHardResponse {
			response := DeleteRoleHardResponse{}
			response.FromNonEntity(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this RoleRest) GetRoleById(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST get role by id"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleService.GetRoleById,
		func(request GetRoleByIdRequest) it.GetRoleByIdQuery {
			return it.GetRoleByIdQuery(request)
		},
		func(result it.GetRoleByIdResult) GetRoleByIdResponse {
			response := GetRoleByIdResponse{}
			response.FromRole(*result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}

func (this RoleRest) SearchRoles(echoCtx echo.Context) (err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "handle REST search roles"); e != nil {
			err = e
		}
	}()

	err = httpserver.ServeRequest(
		echoCtx, this.roleService.SearchRoles,
		func(request SearchRolesRequest) it.SearchRolesQuery {
			return it.SearchRolesQuery(request)
		},
		func(result it.SearchRolesResult) SearchRolesResponse {
			response := SearchRolesResponse{}
			response.FromResult(result.Data)
			return response
		},
		httpserver.JsonOk,
	)

	return err
}
