package v1

import (
	"go.uber.org/dig"

	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductTemplateCombinationExclusionEngineName is the container name of the exclusion onion,
// declared once for the struct tag and checked against dynamicresource.EngineDependencyName by a test.
const ProductTemplateCombinationExclusionEngineName = "dynengine_inventory_product_template_combination_exclusion"

type productTemplateCombinationExclusionRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_product_template_combination_exclusion"`
}

func NewProductTemplateCombinationExclusionRest(
	params productTemplateCombinationExclusionRestParams,
) *ProductTemplateCombinationExclusionRest {
	appSvc := params.Engine.ApplicationService()
	rest := &ProductTemplateCombinationExclusionRest{}
	rest.SetApplicationService(appSvc)
	rest.exclusionSvc = appSvc.(itProduct.ProductTemplateCombinationExclusionApplicationService)
	return rest
}

// ProductTemplateCombinationExclusionRest serves the reads of the exclusion resource. Writes go
// through the template's configuration payload, so the route table registers no write.
type ProductTemplateCombinationExclusionRest struct {
	dynamicresource.CrudRestBase
	exclusionSvc itProduct.ProductTemplateCombinationExclusionApplicationService
}
