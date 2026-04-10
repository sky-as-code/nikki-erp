package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/role"
)

type roleRestParams struct {
	dig.In

	RoleSvc it.RoleService
}

func NewRoleRest(params roleRestParams) *RoleRest {
	return &RoleRest{RoleSvc: params.RoleSvc}
}

type RoleRest struct {
	httpserver.RestBase
	RoleSvc it.RoleService
}

func (this RoleRest) CreateRole(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create role"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.RoleSvc.CreateRole,
		func(requestFields dmodel.DynamicFields) it.CreateRoleCommand {
			cmd := it.CreateRoleCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		func(data domain.Role) CreateRoleResponse {
			response := httpserver.NewRestCreateResponseDyn(data.GetFieldData())
			return *response
		},
		httpserver.JsonCreated,
	)
}

func (this RoleRest) DeleteRole(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete role"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleSvc.DeleteRole,
		func(request DeleteRoleRequest) it.DeleteRoleCommand {
			return it.DeleteRoleCommand(request)
		},
		func(data dyn.MutateResultData) DeleteRoleResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this RoleRest) GetRole(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get role"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleSvc.GetRole,
		func(request GetRoleRequest) it.GetRoleQuery {
			return it.GetRoleQuery(request)
		},
		func(data domain.Role) GetRoleResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this RoleRest) ManageRoleEntitlements(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST manage role entitlements"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleSvc.ManageRoleEntitlements,
		func(request ManageRoleEntitlementsRequest) it.ManageRoleEntitlementsCommand {
			return it.ManageRoleEntitlementsCommand(request)
		},
		func(data dyn.MutateResultData) ManageRoleEntitlementsResponse {
			return httpserver.NewRestMutateResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this RoleRest) RoleExists(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST role exists"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleSvc.RoleExists,
		func(request RoleExistsRequest) it.RoleExistsQuery {
			return it.RoleExistsQuery(request)
		},
		func(data dyn.ExistsResultData) RoleExistsResponse {
			return RoleExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this RoleRest) SearchRoles(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search roles"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleSvc.SearchRoles,
		func(request SearchRolesRequest) it.SearchRolesQuery {
			return it.SearchRolesQuery(request)
		},
		func(data it.SearchRolesResultData) SearchRolesResponse {
			return httpserver.NewSearchUsersResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this RoleRest) SetRoleIsArchived(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST set role is_archived"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequest2(
		echoCtx,
		this.RoleSvc.SetRoleIsArchived,
		func(request SetRoleIsArchivedRequest) it.SetRoleIsArchivedCommand {
			return request
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}

func (this RoleRest) UpdateRole(echoCtx echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update role"); e != nil {
			err = e
		}
	}()
	return httpserver.ServeRequestDynamic(
		echoCtx,
		this.RoleSvc.UpdateRole,
		func(requestFields dmodel.DynamicFields) it.UpdateRoleCommand {
			cmd := it.UpdateRoleCommand{}
			cmd.SetFieldData(requestFields)
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
