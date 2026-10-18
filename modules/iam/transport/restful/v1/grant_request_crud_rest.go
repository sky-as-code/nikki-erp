package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// GrantRequestEngineName is the container name of the iam_grant_request onion, declared once for
// the struct tag.
const GrantRequestEngineName = "dynengine_iam_grant_request"

type grantRequestCrudRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_grant_request"`
}

func NewGrantRequestCrudRest(params grantRequestCrudRestParams) *GrantRequestCrudRest {
	rest := &GrantRequestCrudRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// GrantRequestCrudRest serves the built-in CRUD of the iam_grant_request resource.
type GrantRequestCrudRest struct {
	composable.CrudRestBase
}
