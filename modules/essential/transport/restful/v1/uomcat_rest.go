package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// UomCatEngineName is the container name of the UoM Category onion, declared once for the tag.
const UomCatEngineName = "dynengine_essential_uomcat"

type uomCatRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_uomcat"`
}

func NewUomCatRest(params uomCatRestParams) *UomCatRest {
	rest := &UomCatRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// UomCatRest serves the built-in CRUD of the UoM Category resource.
type UomCatRest struct {
	composable.CrudRestBase
}
