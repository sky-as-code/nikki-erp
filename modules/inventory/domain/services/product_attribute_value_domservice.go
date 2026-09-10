package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductAttributeValueDomainService is handed the composable default by the product_attribute_value onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewProductAttributeValueDomainService(base composable.CrudDomainService) itProduct.ProductAttributeValueDomainService {
	return &ProductAttributeValueDomainServiceImpl{CrudDomainService: base}
}

type ProductAttributeValueDomainServiceImpl struct {
	composable.CrudDomainService
}
