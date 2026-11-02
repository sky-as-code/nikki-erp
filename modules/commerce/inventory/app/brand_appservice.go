package app

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewBrandApplicationService is handed the composable default by the brand onion.
func NewBrandApplicationService(base dynamicresource.CrudApplicationService) itProduct.BrandApplicationService {
	return &BrandApplicationServiceImpl{CrudApplicationService: base}
}

// BrandApplicationServiceImpl is the authorized CRUD of the resource.
type BrandApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
