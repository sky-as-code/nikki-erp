package app

import (
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewProductTemplateAttributeApplicationService is handed the composable default by the product_template_attribute onion.
func NewProductTemplateAttributeApplicationService(base dynamicresource.CrudApplicationService) itProduct.ProductTemplateAttributeApplicationService {
	return &ProductTemplateAttributeApplicationServiceImpl{CrudApplicationService: base}
}

// ProductTemplateAttributeApplicationServiceImpl is the authorized CRUD of the resource.
type ProductTemplateAttributeApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
