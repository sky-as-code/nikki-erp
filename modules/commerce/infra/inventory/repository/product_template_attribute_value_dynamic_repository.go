package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewProductTemplateAttributeValueRepository is handed the composable default by the product_template_attribute_value onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductTemplateAttributeValueRepository(base composable.CrudRepository) itProduct.ProductTemplateAttributeValueRepository {
	return &ProductTemplateAttributeValueRepositoryImpl{CrudRepository: base}
}

type ProductTemplateAttributeValueRepositoryImpl struct {
	composable.CrudRepository
}
