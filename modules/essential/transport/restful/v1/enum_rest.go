package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// EnumEngineName is the container name of the onion, checked by a test.
const EnumEngineName = "dynengine_essential_enum"

type enumRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_enum"`
}

func NewEnumRest(params enumRestParams) *EnumRest {
	rest := &EnumRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// EnumRest serves the built-in CRUD of the resource.
type EnumRest struct {
	composable.CrudRestBase
}
