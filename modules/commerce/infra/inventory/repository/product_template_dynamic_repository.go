package repository

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewProductTemplateRepository is handed the composable default by the product_template onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewProductTemplateRepository(base composable.CrudRepository) itProduct.ProductTemplateRepository {
	return &ProductTemplateRepositoryImpl{CrudRepository: base}
}

type ProductTemplateRepositoryImpl struct {
	composable.CrudRepository
}
