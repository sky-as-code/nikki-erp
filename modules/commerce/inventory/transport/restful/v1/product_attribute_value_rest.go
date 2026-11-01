package v1

import (
	"go.uber.org/dig"

	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductAttributeValueEngineName is the container name of the product_attribute_value onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const ProductAttributeValueEngineName = "dynengine_inventory_product_attribute_value"

type productAttributeValueRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_attribute_value"`
}

func NewProductAttributeValueRest(params productAttributeValueRestParams) *ProductAttributeValueRest {
	rest := &ProductAttributeValueRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.productAttributeValueSvc = params.Engine.ApplicationService().(itProduct.ProductAttributeValueApplicationService)
	return rest
}

// ProductAttributeValueRest serves the built-in CRUD of the product_attribute_value resource.
type ProductAttributeValueRest struct {
	dynamicresource.CrudRestBase
	productAttributeValueSvc itProduct.ProductAttributeValueApplicationService
}
