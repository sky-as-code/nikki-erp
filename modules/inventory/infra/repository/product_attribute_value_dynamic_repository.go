package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductAttributeValueRepository is handed the composable default by the product_attribute_value onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductAttributeValueRepository(base composable.CrudRepository) itProduct.ProductAttributeValueRepository {
	return &ProductAttributeValueRepositoryImpl{CrudRepository: base}
}

type ProductAttributeValueRepositoryImpl struct {
	composable.CrudRepository
}
