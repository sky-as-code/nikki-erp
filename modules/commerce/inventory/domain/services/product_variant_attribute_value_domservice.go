package services

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewProductVariantAttributeValueDomainService is handed the composable default by the product_variant_attribute_value onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewProductVariantAttributeValueDomainService(base dynamicresource.CrudDomainService) itProduct.ProductVariantAttributeValueDomainService {
	return &ProductVariantAttributeValueDomainServiceImpl{CrudDomainService: base}
}

type ProductVariantAttributeValueDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}
