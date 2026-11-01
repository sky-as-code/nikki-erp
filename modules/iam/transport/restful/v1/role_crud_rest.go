package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

// RoleEngineName is the container name of the iam_role onion, declared once for the struct tag.
const RoleEngineName = "dynengine_iam_role"

type roleCrudRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_iam_role"`
}

func NewRoleCrudRest(params roleCrudRestParams) *RoleCrudRest {
	rest := &RoleCrudRest{roleSvc: params.Engine.ApplicationService().(itRole.RoleCrudApplicationService)}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// RoleCrudRest serves the built-in CRUD of the iam_role resource and its custom actions.
type RoleCrudRest struct {
	dynamicresource.CrudRestBase
	roleSvc itRole.RoleCrudApplicationService
}

func (this *RoleCrudRest) ManageUsers(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage role users", payload, this.roleSvc.ManageUsers,
		dynamicresource.MutateResponse)
}

func (this *RoleCrudRest) ManageServices(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage role services", payload, this.roleSvc.ManageServices,
		dynamicresource.MutateResponse)
}

func (this *RoleCrudRest) ManageGroups(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage role groups", payload, this.roleSvc.ManageGroups,
		dynamicresource.MutateResponse)
}

func (this *RoleCrudRest) Describe(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "describe roles", payload, this.roleSvc.Describe,
		dynamicresource.Identity[itRole.DescribeRolesResultData])
}
