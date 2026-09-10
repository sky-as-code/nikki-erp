package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTypeRepository is handed the composable default by the product_type onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductTypeRepository(base composable.CrudRepository) itProduct.ProductTypeRepository {
	return &ProductTypeRepositoryImpl{CrudRepository: base}
}

type ProductTypeRepositoryImpl struct {
	composable.CrudRepository
}
