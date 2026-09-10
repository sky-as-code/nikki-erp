package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTypeDomainService is handed the composable default by the product_type onion. The resource is plain
// CRUD; the type exists so a rule has a home when one is needed.
func NewProductTypeDomainService(base composable.CrudDomainService) itProduct.ProductTypeDomainService {
	return &ProductTypeDomainServiceImpl{CrudDomainService: base}
}

type ProductTypeDomainServiceImpl struct {
	composable.CrudDomainService
}
