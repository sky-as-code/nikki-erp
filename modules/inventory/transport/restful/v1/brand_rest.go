package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// BrandEngineName is the container name of the brand onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const BrandEngineName = "dynengine_inventory_brand"

type brandRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_brand"`
}

func NewBrandRest(params brandRestParams) *BrandRest {
	rest := &BrandRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.brandSvc = params.Engine.ApplicationService().(itProduct.BrandApplicationService)
	return rest
}

// BrandRest serves the built-in CRUD of the brand resource.
type BrandRest struct {
	composable.CrudRestBase
	brandSvc itProduct.BrandApplicationService
}
