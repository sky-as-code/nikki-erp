package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// ProductTypeEngineName is the container name of the product_type onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const ProductTypeEngineName = "dynengine_inventory_product_type"

type productTypeRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_type"`
}

func NewProductTypeRest(params productTypeRestParams) *ProductTypeRest {
	rest := &ProductTypeRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.productTypeSvc = params.Engine.ApplicationService().(itProduct.ProductTypeApplicationService)
	return rest
}

// ProductTypeRest serves the built-in CRUD of the product_type resource.
type ProductTypeRest struct {
	composable.CrudRestBase
	productTypeSvc itProduct.ProductTypeApplicationService
}
