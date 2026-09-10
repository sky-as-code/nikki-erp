package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTypeApplicationService is handed the composable default by the product_type onion.
func NewProductTypeApplicationService(base composable.CrudApplicationService) itProduct.ProductTypeApplicationService {
	return &ProductTypeApplicationServiceImpl{CrudApplicationService: base}
}

// ProductTypeApplicationServiceImpl is the authorized CRUD of the resource.
type ProductTypeApplicationServiceImpl struct {
	composable.CrudApplicationService
}
