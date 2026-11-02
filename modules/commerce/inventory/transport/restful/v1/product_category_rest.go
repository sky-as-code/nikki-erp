package v1

import (
	"go.uber.org/dig"

	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductCategoryEngineName is the container name of the product_category onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const ProductCategoryEngineName = "dynengine_inventory_product_category"

type productCategoryRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_category"`
}

func NewProductCategoryRest(params productCategoryRestParams) *ProductCategoryRest {
	rest := &ProductCategoryRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.productCategorySvc = params.Engine.ApplicationService().(itProduct.ProductCategoryApplicationService)
	return rest
}

// ProductCategoryRest serves the built-in CRUD of the product_category resource.
type ProductCategoryRest struct {
	dynamicresource.CrudRestBase
	productCategorySvc itProduct.ProductCategoryApplicationService
}
