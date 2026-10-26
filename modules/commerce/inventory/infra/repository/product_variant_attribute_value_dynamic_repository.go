package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductVariantAttributeValueRepository is handed the composable default by the product_variant_attribute_value onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductVariantAttributeValueRepository(base composable.CrudRepository) itProduct.ProductVariantAttributeValueRepository {
	return &ProductVariantAttributeValueRepositoryImpl{CrudRepository: base}
}

type ProductVariantAttributeValueRepositoryImpl struct {
	composable.CrudRepository
}
