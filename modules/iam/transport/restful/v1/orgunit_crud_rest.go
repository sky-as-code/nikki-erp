package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
)

// OrgUnitEngineName is the container name of the iam_orgunit onion, declared once for the tag.
const OrgUnitEngineName = "dynengine_iam_orgunit"

type orgUnitCrudRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_orgunit"`
}

func NewOrgUnitCrudRest(params orgUnitCrudRestParams) *OrgUnitCrudRest {
	svc := params.Engine.ApplicationService().(itOrgUnit.OrgUnitCrudApplicationService)
	rest := &OrgUnitCrudRest{orgUnitSvc: svc}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// OrgUnitCrudRest serves the built-in CRUD of the iam_orgunit resource and its custom actions.
type OrgUnitCrudRest struct {
	composable.CrudRestBase
	orgUnitSvc itOrgUnit.OrgUnitCrudApplicationService
}

func (this *OrgUnitCrudRest) ManageUsers(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "manage org unit users", payload, this.orgUnitSvc.ManageUsers,
		composable.MutateResponse)
}
