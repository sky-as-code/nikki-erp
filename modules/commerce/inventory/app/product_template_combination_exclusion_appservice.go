package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTemplateCombinationExclusionApplicationService is handed the composable default by the
// product_template_combination_exclusion onion.
func NewProductTemplateCombinationExclusionApplicationService(
	base composable.CrudApplicationService,
) itProduct.ProductTemplateCombinationExclusionApplicationService {
	return &ProductTemplateCombinationExclusionApplicationServiceImpl{CrudApplicationService: base}
}

// ProductTemplateCombinationExclusionApplicationServiceImpl is the authorized CRUD of the resource.
type ProductTemplateCombinationExclusionApplicationServiceImpl struct {
	composable.CrudApplicationService
}
