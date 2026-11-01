package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewProductVariantRepository is handed the composable default by the product_variant onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductVariantRepository(base dynamicresource.CrudRepository) itProduct.ProductVariantRepository {
	return &ProductVariantRepositoryImpl{CrudRepository: base}
}

type ProductVariantRepositoryImpl struct {
	dynamicresource.CrudRepository
}
