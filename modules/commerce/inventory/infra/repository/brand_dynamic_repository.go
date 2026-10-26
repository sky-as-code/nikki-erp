package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewBrandRepository is handed the composable default by the brand onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewBrandRepository(base composable.CrudRepository) itProduct.BrandRepository {
	return &BrandRepositoryImpl{CrudRepository: base}
}

type BrandRepositoryImpl struct {
	composable.CrudRepository
}
