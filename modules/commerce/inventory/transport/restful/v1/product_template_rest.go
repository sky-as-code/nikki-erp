package v1

import (
	"go.uber.org/dig"

	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductTemplateEngineName is the container name of the product_template onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const ProductTemplateEngineName = "dynengine_inventory_product_template"

type productTemplateRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template"`
}

func NewProductTemplateRest(params productTemplateRestParams) *ProductTemplateRest {
	svc := params.Engine.ApplicationService().(itProduct.ProductTemplateApplicationService)
	rest := &ProductTemplateRest{productTemplateSvc: svc}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// ProductTemplateRest serves the built-in CRUD of the product_template resource.
type ProductTemplateRest struct {
	composable.CrudRestBase
	productTemplateSvc itProduct.ProductTemplateApplicationService
}
