package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductAttributeDomainService is handed the composable default by the product_attribute onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewProductAttributeDomainService(base composable.CrudDomainService) itProduct.ProductAttributeDomainService {
	return &ProductAttributeDomainServiceImpl{CrudDomainService: base}
}

type ProductAttributeDomainServiceImpl struct {
	composable.CrudDomainService
}
