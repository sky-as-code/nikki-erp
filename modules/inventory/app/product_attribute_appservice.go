package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductAttributeApplicationService is handed the composable default by the product_attribute onion.
func NewProductAttributeApplicationService(base composable.CrudApplicationService) itProduct.ProductAttributeApplicationService {
	return &ProductAttributeApplicationServiceImpl{CrudApplicationService: base}
}

// ProductAttributeApplicationServiceImpl is the authorized CRUD of the resource.
type ProductAttributeApplicationServiceImpl struct {
	composable.CrudApplicationService
}
