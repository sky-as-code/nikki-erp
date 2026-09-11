package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductTemplateAttributeValueApplicationService is handed the composable default by the product_template_attribute_value onion.
func NewProductTemplateAttributeValueApplicationService(base composable.CrudApplicationService) itProduct.ProductTemplateAttributeValueApplicationService {
	return &ProductTemplateAttributeValueApplicationServiceImpl{CrudApplicationService: base}
}

// ProductTemplateAttributeValueApplicationServiceImpl is the authorized CRUD of the resource.
type ProductTemplateAttributeValueApplicationServiceImpl struct {
	composable.CrudApplicationService
}
