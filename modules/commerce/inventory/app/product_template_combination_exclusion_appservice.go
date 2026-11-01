package app

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewProductTemplateCombinationExclusionApplicationService is handed the composable default by the
// product_template_combination_exclusion onion.
func NewProductTemplateCombinationExclusionApplicationService(
	base dynamicresource.CrudApplicationService,
) itProduct.ProductTemplateCombinationExclusionApplicationService {
	return &ProductTemplateCombinationExclusionApplicationServiceImpl{CrudApplicationService: base}
}

// ProductTemplateCombinationExclusionApplicationServiceImpl is the authorized CRUD of the resource.
type ProductTemplateCombinationExclusionApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
