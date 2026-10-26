package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewProductTemplateCombinationExclusionRepository is handed the composable default by the
// product_template_combination_exclusion onion. The type exists so a bespoke query has a home.
func NewProductTemplateCombinationExclusionRepository(
	base composable.CrudRepository,
) itProduct.ProductTemplateCombinationExclusionRepository {
	return &ProductTemplateCombinationExclusionRepositoryImpl{CrudRepository: base}
}

type ProductTemplateCombinationExclusionRepositoryImpl struct {
	composable.CrudRepository
}
