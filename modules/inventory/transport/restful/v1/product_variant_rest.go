package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ProductVariantEngineName is the container name of the product_variant onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const ProductVariantEngineName = "dynengine_inventory_product_variant"

type productVariantRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_variant"`
}

func NewProductVariantRest(params productVariantRestParams) *ProductVariantRest {
	rest := &ProductVariantRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.productVariantSvc = params.Engine.ApplicationService().(itProduct.ProductVariantApplicationService)
	return rest
}

// ProductVariantRest serves the built-in CRUD of the product_variant resource.
type ProductVariantRest struct {
	composable.CrudRestBase
	productVariantSvc itProduct.ProductVariantApplicationService
}
