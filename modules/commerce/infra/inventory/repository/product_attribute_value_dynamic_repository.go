package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewProductAttributeValueRepository is handed the composable default by the product_attribute_value onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductAttributeValueRepository(base dynamicresource.CrudRepository) itProduct.ProductAttributeValueRepository {
	return &ProductAttributeValueRepositoryImpl{CrudRepository: base}
}

type ProductAttributeValueRepositoryImpl struct {
	dynamicresource.CrudRepository
}
