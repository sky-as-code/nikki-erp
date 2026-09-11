package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductAttributeValueApplicationService is handed the composable default by the product_attribute_value onion.
func NewProductAttributeValueApplicationService(base composable.CrudApplicationService) itProduct.ProductAttributeValueApplicationService {
	return &ProductAttributeValueApplicationServiceImpl{CrudApplicationService: base}
}

// ProductAttributeValueApplicationServiceImpl is the authorized CRUD of the resource.
type ProductAttributeValueApplicationServiceImpl struct {
	composable.CrudApplicationService
}
