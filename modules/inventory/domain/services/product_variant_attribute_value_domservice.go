package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductVariantAttributeValueDomainService is handed the composable default by the product_variant_attribute_value onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewProductVariantAttributeValueDomainService(base composable.CrudDomainService) itProduct.ProductVariantAttributeValueDomainService {
	return &ProductVariantAttributeValueDomainServiceImpl{CrudDomainService: base}
}

type ProductVariantAttributeValueDomainServiceImpl struct {
	composable.CrudDomainService
}
