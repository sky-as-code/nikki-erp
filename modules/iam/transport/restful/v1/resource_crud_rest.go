package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ResourceEngineName and ActionEngineName are the container names of the iam_resource and
// iam_action onions, declared once for the struct tags.
const (
	ResourceEngineName = "dynengine_iam_resource"
	ActionEngineName   = "dynengine_iam_action"
)

type resourceCrudRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_resource"`
}

func NewResourceCrudRest(params resourceCrudRestParams) *ResourceCrudRest {
	rest := &ResourceCrudRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// ResourceCrudRest serves the built-in CRUD of the iam_resource resource.
type ResourceCrudRest struct {
	composable.CrudRestBase
}

type actionCrudRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_action"`
}

func NewActionCrudRest(params actionCrudRestParams) *ActionCrudRest {
	rest := &ActionCrudRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// ActionCrudRest serves the built-in CRUD of the iam_action resource.
type ActionCrudRest struct {
	composable.CrudRestBase
}
