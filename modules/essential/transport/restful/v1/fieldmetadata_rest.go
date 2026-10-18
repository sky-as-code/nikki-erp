package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// FieldMetadataEngineName is the container name of the onion, checked by a test.
const FieldMetadataEngineName = "dynengine_essential_field_metadata"

type fieldmetadataRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_field_metadata"`
}

func NewFieldMetadataRest(params fieldmetadataRestParams) *FieldMetadataRest {
	rest := &FieldMetadataRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// FieldMetadataRest serves the built-in CRUD of the resource.
type FieldMetadataRest struct {
	composable.CrudRestBase
}
