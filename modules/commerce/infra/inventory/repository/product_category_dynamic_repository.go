package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewProductCategoryRepository is handed the composable default by the product_category onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductCategoryRepository(base composable.CrudRepository) itProduct.ProductCategoryRepository {
	return &ProductCategoryRepositoryImpl{CrudRepository: base}
}

type ProductCategoryRepositoryImpl struct {
	composable.CrudRepository
}
