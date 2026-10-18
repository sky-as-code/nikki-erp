package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ModelMetadataEngineName is the container name of the onion, checked by a test.
const ModelMetadataEngineName = "dynengine_essential_model_metadata"

type modelmetadataRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_model_metadata"`
}

func NewModelMetadataRest(params modelmetadataRestParams) *ModelMetadataRest {
	rest := &ModelMetadataRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// ModelMetadataRest serves the built-in CRUD of the resource.
type ModelMetadataRest struct {
	composable.CrudRestBase
}
