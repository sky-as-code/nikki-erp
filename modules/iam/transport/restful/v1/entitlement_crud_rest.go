package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// EntitlementEngineName is the container name of the iam_entitlement onion, declared once for the
// struct tag.
const EntitlementEngineName = "dynengine_iam_entitlement"

type entitlementCrudRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_entitlement"`
}

func NewEntitlementCrudRest(params entitlementCrudRestParams) *EntitlementCrudRest {
	rest := &EntitlementCrudRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// EntitlementCrudRest serves the built-in CRUD of the iam_entitlement resource.
type EntitlementCrudRest struct {
	composable.CrudRestBase
}
