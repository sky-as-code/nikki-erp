package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductCategoryApplicationService is handed the composable default by the product_category onion.
func NewProductCategoryApplicationService(base composable.CrudApplicationService) itProduct.ProductCategoryApplicationService {
	return &ProductCategoryApplicationServiceImpl{CrudApplicationService: base}
}

// ProductCategoryApplicationServiceImpl is the authorized CRUD of the resource.
type ProductCategoryApplicationServiceImpl struct {
	composable.CrudApplicationService
}
