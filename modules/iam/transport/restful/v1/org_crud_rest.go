package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
)

// OrgEngineName is the container name of the iam_org onion, declared once for the struct tag.
const OrgEngineName = "dynengine_iam_org"

type orgCrudRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_org"`
}

func NewOrgCrudRest(params orgCrudRestParams) *OrgCrudRest {
	rest := &OrgCrudRest{orgSvc: params.Engine.ApplicationService().(itOrg.OrgCrudApplicationService)}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// OrgCrudRest serves the built-in CRUD of the iam_org resource and its custom actions.
type OrgCrudRest struct {
	composable.CrudRestBase
	orgSvc itOrg.OrgCrudApplicationService
}

func (this *OrgCrudRest) ManageUsers(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "manage organization users", payload, this.orgSvc.ManageUsers,
		composable.MutateResponse)
}
