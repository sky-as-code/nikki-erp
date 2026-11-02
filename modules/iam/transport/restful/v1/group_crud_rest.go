package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itGroup "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

// GroupEngineName is the container name of the iam_group onion, declared once for the struct tag.
const GroupEngineName = "dynengine_iam_group"

type groupCrudRestParams struct {
	dig.In

	Engine         dynamicresource.DynamicResourceEngineOnion `name:"dynengine_iam_group"`
	PrincipalRoles itRole.PrincipalRolesApplicationService
}

func NewGroupCrudRest(params groupCrudRestParams) *GroupCrudRest {
	rest := &GroupCrudRest{
		groupSvc: params.Engine.ApplicationService().(itGroup.GroupCrudApplicationService),
		rolesSvc: params.PrincipalRoles,
	}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// GroupCrudRest serves the built-in CRUD of the iam_group resource and its custom actions.
type GroupCrudRest struct {
	dynamicresource.CrudRestBase
	groupSvc itGroup.GroupCrudApplicationService
	rolesSvc itRole.PrincipalRolesApplicationService
}

func (this *GroupCrudRest) Roles(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "search group roles", payload, this.rolesSvc.GroupRoles,
		dynamicresource.SearchResponse)
}

func (this *GroupCrudRest) ManageUsers(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage group users", payload, this.groupSvc.ManageUsers,
		dynamicresource.MutateResponse)
}

func (this *GroupCrudRest) ManageRoles(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "manage group roles", payload, this.groupSvc.ManageRoles,
		dynamicresource.MutateResponse)
}
