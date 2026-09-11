package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTemplateAttributeValueDomainService is handed the composable default by the product_template_attribute_value onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewProductTemplateAttributeValueDomainService(base composable.CrudDomainService) itProduct.ProductTemplateAttributeValueDomainService {
	return &ProductTemplateAttributeValueDomainServiceImpl{CrudDomainService: base}
}

type ProductTemplateAttributeValueDomainServiceImpl struct {
	composable.CrudDomainService
}
