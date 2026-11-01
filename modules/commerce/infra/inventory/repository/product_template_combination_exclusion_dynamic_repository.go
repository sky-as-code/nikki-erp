package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewProductTemplateCombinationExclusionRepository is handed the composable default by the
// product_template_combination_exclusion onion. The type exists so a bespoke query has a home.
func NewProductTemplateCombinationExclusionRepository(
	base dynamicresource.CrudRepository,
) itProduct.ProductTemplateCombinationExclusionRepository {
	return &ProductTemplateCombinationExclusionRepositoryImpl{CrudRepository: base}
}

type ProductTemplateCombinationExclusionRepositoryImpl struct {
	dynamicresource.CrudRepository
}
