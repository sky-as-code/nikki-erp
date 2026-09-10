package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewBrandDomainService is handed the composable default by the brand onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewBrandDomainService(base composable.CrudDomainService) itProduct.BrandDomainService {
	return &BrandDomainServiceImpl{CrudDomainService: base}
}

type BrandDomainServiceImpl struct {
	composable.CrudDomainService
}
