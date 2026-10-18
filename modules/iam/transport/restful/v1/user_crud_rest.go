package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable/filefield"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// UserEngineName is the container name of the iam_user onion, declared once for the struct tag.
const UserEngineName = "dynengine_iam_user"

type userCrudRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_user"`
}

func NewUserCrudRest(params userCrudRestParams) *UserCrudRest {
	svc := params.Engine.ApplicationService().(itUser.UserCrudApplicationService)
	rest := &UserCrudRest{Rest: filefield.NewRest("user", svc), userSvc: svc}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// UserCrudRest serves the built-in CRUD of the iam_user resource, its file routes and its
// custom actions.
type UserCrudRest struct {
	composable.CrudRestBase
	filefield.Rest
	userSvc itUser.UserCrudApplicationService
}

func (this *UserCrudRest) SendInvitation(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "send user invitation", payload, this.userSvc.SendInvitation,
		composable.Identity[any])
}
