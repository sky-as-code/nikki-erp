package v1

import (
	"go.uber.org/dig"

	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductTemplateAttributeEngineName is the container name of the product_template_attribute onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const ProductTemplateAttributeEngineName = "dynengine_inventory_product_template_attribute"

type productTemplateAttributeRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template_attribute"`
}

func NewProductTemplateAttributeRest(params productTemplateAttributeRestParams) *ProductTemplateAttributeRest {
	rest := &ProductTemplateAttributeRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.productTemplateAttributeSvc = params.Engine.ApplicationService().(itProduct.ProductTemplateAttributeApplicationService)
	return rest
}

// ProductTemplateAttributeRest serves the built-in CRUD of the product_template_attribute resource.
type ProductTemplateAttributeRest struct {
	dynamicresource.CrudRestBase
	productTemplateAttributeSvc itProduct.ProductTemplateAttributeApplicationService
}
