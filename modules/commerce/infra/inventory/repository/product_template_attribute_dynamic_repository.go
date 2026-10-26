package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewProductTemplateAttributeRepository is handed the composable default by the product_template_attribute onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductTemplateAttributeRepository(base composable.CrudRepository) itProduct.ProductTemplateAttributeRepository {
	return &ProductTemplateAttributeRepositoryImpl{CrudRepository: base}
}

type ProductTemplateAttributeRepositoryImpl struct {
	composable.CrudRepository
}
