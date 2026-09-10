package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ProductAttributeEngineName is the container name of the product_attribute onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const ProductAttributeEngineName = "dynengine_inventory_product_attribute"

type productAttributeRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_attribute"`
}

func NewProductAttributeRest(params productAttributeRestParams) *ProductAttributeRest {
	rest := &ProductAttributeRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.productAttributeSvc = params.Engine.ApplicationService().(itProduct.ProductAttributeApplicationService)
	return rest
}

// ProductAttributeRest serves the built-in CRUD of the product_attribute resource.
type ProductAttributeRest struct {
	composable.CrudRestBase
	productAttributeSvc itProduct.ProductAttributeApplicationService
}
