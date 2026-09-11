package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ProductAttributeValueEngineName is the container name of the product_attribute_value onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const ProductAttributeValueEngineName = "dynengine_inventory_product_attribute_value"

type productAttributeValueRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_attribute_value"`
}

func NewProductAttributeValueRest(params productAttributeValueRestParams) *ProductAttributeValueRest {
	rest := &ProductAttributeValueRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.productAttributeValueSvc = params.Engine.ApplicationService().(itProduct.ProductAttributeValueApplicationService)
	return rest
}

// ProductAttributeValueRest serves the built-in CRUD of the product_attribute_value resource.
type ProductAttributeValueRest struct {
	composable.CrudRestBase
	productAttributeValueSvc itProduct.ProductAttributeValueApplicationService
}
