package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// UomEngineName is the container name of the UoM onion, declared once for the struct tag.
const UomEngineName = "dynengine_essential_uom"

type uomRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_uom"`
}

func NewUomRest(params uomRestParams) *UomRest {
	rest := &UomRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// UomRest serves the built-in CRUD of the UoM resource. Conversion is a calculation over two
// records rather than an action on one, so it stays on UomConversionRest.
type UomRest struct {
	composable.CrudRestBase
}
