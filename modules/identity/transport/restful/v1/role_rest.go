package v1

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
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
	return httpserver.ServeCreate(
		"create role",
		echoCtx,
		&it.CreateRoleCommand{},
		this.RoleSvc.CreateRole,
	)
}

func (this RoleRest) DeleteRole(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"delete role",
		echoCtx,
		this.RoleSvc.DeleteRole,
	)
}

func (this RoleRest) GetRole(echoCtx echo.Context) (err error) {
	return httpserver.ServeGetOne(
		"get role",
		echoCtx,
		this.RoleSvc.GetRole,
	)
}

func (this RoleRest) ManageRoleEntitlements(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"manage role entitlements",
		echoCtx,
		this.RoleSvc.ManageRoleEntitlements,
	)
}

func (this RoleRest) RoleExists(echoCtx echo.Context) (err error) {
	return httpserver.ServeExists(
		"role exists",
		echoCtx,
		this.RoleSvc.RoleExists,
	)
}

func (this RoleRest) SearchRoles(echoCtx echo.Context) (err error) {
	return httpserver.ServeSearch(
		"search roles",
		echoCtx,
		this.RoleSvc.SearchRoles,
		true,
	)
}

func (this RoleRest) SetRoleIsArchived(echoCtx echo.Context) (err error) {
	return httpserver.ServeGeneralMutate(
		"set role is_archived",
		echoCtx,
		this.RoleSvc.SetRoleIsArchived,
	)
}

func (this RoleRest) UpdateRole(echoCtx echo.Context) (err error) {
	return httpserver.ServeUpdate(
		"update role",
		echoCtx,
		&it.UpdateRoleCommand{},
		this.RoleSvc.UpdateRole,
	)
}
