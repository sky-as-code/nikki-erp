package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ProductTemplateAttributeValueEngineName is the container name of the product_template_attribute_value onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const ProductTemplateAttributeValueEngineName = "dynengine_inventory_product_template_attribute_value"

type productTemplateAttributeValueRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template_attribute_value"`
}

func NewProductTemplateAttributeValueRest(params productTemplateAttributeValueRestParams) *ProductTemplateAttributeValueRest {
	rest := &ProductTemplateAttributeValueRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.productTemplateAttributeValueSvc = params.Engine.ApplicationService().(itProduct.ProductTemplateAttributeValueApplicationService)
	return rest
}

// ProductTemplateAttributeValueRest serves the built-in CRUD of the product_template_attribute_value resource.
type ProductTemplateAttributeValueRest struct {
	composable.CrudRestBase
	productTemplateAttributeValueSvc itProduct.ProductTemplateAttributeValueApplicationService
}
