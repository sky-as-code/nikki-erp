package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewProductAttributeRepository is handed the composable default by the product_attribute onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductAttributeRepository(base dynamicresource.CrudRepository) itProduct.ProductAttributeRepository {
	return &ProductAttributeRepositoryImpl{CrudRepository: base}
}

type ProductAttributeRepositoryImpl struct {
	dynamicresource.CrudRepository
}
