package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTemplateAttributeDomainService is handed the composable default by the product_template_attribute onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewProductTemplateAttributeDomainService(base composable.CrudDomainService) itProduct.ProductTemplateAttributeDomainService {
	return &ProductTemplateAttributeDomainServiceImpl{CrudDomainService: base}
}

type ProductTemplateAttributeDomainServiceImpl struct {
	composable.CrudDomainService
}
