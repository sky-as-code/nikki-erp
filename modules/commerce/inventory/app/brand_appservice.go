package app

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewBrandApplicationService is handed the composable default by the brand onion.
func NewBrandApplicationService(base composable.CrudApplicationService) itProduct.BrandApplicationService {
	return &BrandApplicationServiceImpl{CrudApplicationService: base}
}

// BrandApplicationServiceImpl is the authorized CRUD of the resource.
type BrandApplicationServiceImpl struct {
	composable.CrudApplicationService
}
