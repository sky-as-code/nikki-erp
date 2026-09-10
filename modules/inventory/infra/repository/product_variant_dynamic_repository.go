package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductVariantRepository is handed the composable default by the product_variant onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductVariantRepository(base composable.CrudRepository) itProduct.ProductVariantRepository {
	return &ProductVariantRepositoryImpl{CrudRepository: base}
}

type ProductVariantRepositoryImpl struct {
	composable.CrudRepository
}
