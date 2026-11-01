package app

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewProductAttributeValueApplicationService is handed the composable default by the product_attribute_value onion.
func NewProductAttributeValueApplicationService(base dynamicresource.CrudApplicationService) itProduct.ProductAttributeValueApplicationService {
	return &ProductAttributeValueApplicationServiceImpl{CrudApplicationService: base}
}

// ProductAttributeValueApplicationServiceImpl is the authorized CRUD of the resource.
type ProductAttributeValueApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
