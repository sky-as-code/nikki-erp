package app

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewProductVariantAttributeValueApplicationService is handed the composable default by the product_variant_attribute_value onion.
func NewProductVariantAttributeValueApplicationService(base composable.CrudApplicationService) itProduct.ProductVariantAttributeValueApplicationService {
	return &ProductVariantAttributeValueApplicationServiceImpl{CrudApplicationService: base}
}

// ProductVariantAttributeValueApplicationServiceImpl is the authorized CRUD of the resource.
type ProductVariantAttributeValueApplicationServiceImpl struct {
	composable.CrudApplicationService
}
