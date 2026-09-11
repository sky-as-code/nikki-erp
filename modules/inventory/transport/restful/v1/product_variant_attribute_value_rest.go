package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ProductVariantAttributeValueEngineName is the container name of the product_variant_attribute_value onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const ProductVariantAttributeValueEngineName = "dynengine_inventory_product_variant_attribute_value"

type productVariantAttributeValueRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_variant_attribute_value"`
}

func NewProductVariantAttributeValueRest(params productVariantAttributeValueRestParams) *ProductVariantAttributeValueRest {
	rest := &ProductVariantAttributeValueRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.productVariantAttributeValueSvc = params.Engine.ApplicationService().(itProduct.ProductVariantAttributeValueApplicationService)
	return rest
}

// ProductVariantAttributeValueRest serves the built-in CRUD of the product_variant_attribute_value resource.
type ProductVariantAttributeValueRest struct {
	composable.CrudRestBase
	productVariantAttributeValueSvc itProduct.ProductVariantAttributeValueApplicationService
}
