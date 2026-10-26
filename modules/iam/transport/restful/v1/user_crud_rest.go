package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// UserEngineName is the container name of the iam_user onion, declared once for the struct tag.
const UserEngineName = "dynengine_iam_user"

type userCrudRestParams struct {
	dig.In

	Engine         composable.DynamicResourceEngineOnion `name:"dynengine_iam_user"`
	PrincipalRoles itRole.PrincipalRolesApplicationService
}

func NewUserCrudRest(params userCrudRestParams) *UserCrudRest {
	svc := params.Engine.ApplicationService().(itUser.UserCrudApplicationService)
	rest := &UserCrudRest{userSvc: svc, rolesSvc: params.PrincipalRoles}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// UserCrudRest serves the built-in CRUD of the iam_user resource and its custom actions.
type UserCrudRest struct {
	composable.CrudRestBase
	userSvc  itUser.UserCrudApplicationService
	rolesSvc itRole.PrincipalRolesApplicationService
}

func (this *UserCrudRest) ManageRoles(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "manage user roles", payload, this.userSvc.ManageRoles,
		composable.MutateResponse)
}

func (this *UserCrudRest) Roles(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "search user roles", payload, this.rolesSvc.UserRoles,
		composable.SearchResponse)
}

func (this *UserCrudRest) SendInvitation(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "send user invitation", payload, this.userSvc.SendInvitation,
		composable.Identity[any])
}
