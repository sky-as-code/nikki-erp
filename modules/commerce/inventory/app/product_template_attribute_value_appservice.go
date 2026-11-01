package app

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/constants"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// templateValueUpdatePermission is what PATCH product_template_attribute_value/:id asserts. The
// domain service accepts only sales_price_extra there, and pricing an attribute value is a power
// over the attribute, so it is granted on the attribute resource rather than as an update of a
// configuration row that is otherwise written only through the template.
var templateValueUpdatePermission = dynamicresource.ActionPermission{
	ActionCode:   PermissionUpdateValue,
	ResourceCode: constants.ResourceProductAttribute,
}

// NewProductTemplateAttributeValueApplicationService is handed the composable default by the product_template_attribute_value onion.
func NewProductTemplateAttributeValueApplicationService(base dynamicresource.CrudApplicationService) itProduct.ProductTemplateAttributeValueApplicationService {
	return &ProductTemplateAttributeValueApplicationServiceImpl{CrudApplicationService: base}
}

// ProductTemplateAttributeValueApplicationServiceImpl is the authorized CRUD of the resource.
type ProductTemplateAttributeValueApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

// Update asserts update_value on inventory_product_attribute instead of the default update on this
// resource, keeps the record-in-org check, then reaches the domain service directly: delegating to
// the embedded Update would assert the default permission as well.
func (this *ProductTemplateAttributeValueApplicationServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand,
) (*dynamicresource.MutateResult, error) {
	cErrs, err := dynamicresource.AssertRecordAction(ctx, this.CrudApplicationService, cmd, templateValueUpdatePermission)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &dynamicresource.MutateResult{ClientErrors: *cErrs}, nil
	}
	return this.DomainService().Update(ctx, cmd)
}
