package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTemplateAttributeApplicationService is handed the composable default by the product_template_attribute onion.
func NewProductTemplateAttributeApplicationService(base composable.CrudApplicationService) itProduct.ProductTemplateAttributeApplicationService {
	return &ProductTemplateAttributeApplicationServiceImpl{CrudApplicationService: base}
}

// ProductTemplateAttributeApplicationServiceImpl is the authorized CRUD of the resource.
type ProductTemplateAttributeApplicationServiceImpl struct {
	composable.CrudApplicationService
}
