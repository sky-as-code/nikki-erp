package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// TagEngineName is the container name of the onion, checked by a test.
const TagEngineName = "dynengine_essential_tag"

type tagRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_tag"`
}

func NewTagRest(params tagRestParams) *TagRest {
	rest := &TagRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// TagRest serves the built-in CRUD of the resource.
type TagRest struct {
	composable.CrudRestBase
}
