package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewBrandRepository is handed the composable default by the brand onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewBrandRepository(base dynamicresource.CrudRepository) itProduct.BrandRepository {
	return &BrandRepositoryImpl{CrudRepository: base}
}

type BrandRepositoryImpl struct {
	dynamicresource.CrudRepository
}
