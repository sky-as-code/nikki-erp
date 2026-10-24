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
//
// The inline `values` delta on Create and Update is authorized by the attribute's own create or
// update permission: the values belong to the attribute and are edited as part of it, so the
// domain service writes them through the value domain service without a second permission
// check. The standalone product_attribute_value routes keep their own permissions.
type ProductAttributeApplicationServiceImpl struct {
	composable.CrudApplicationService
}
