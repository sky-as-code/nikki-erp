package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// AssertTemplateNotUsedExternally asks every dependant, and the Purchase and Sales parts of Commerce,
// whether they still name this template.
//
// The variant check does not cover it: Sales pricelist items and Purchase vendor prices name a
// template directly, to price all its variants at once, and neither column carries a foreign key,
// so this check is the only thing keeping such a rule from pointing at nothing.
func AssertTemplateNotUsedExternally(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher,
	params dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	templateId := readStringParam(params, models.ProductTemplateFieldId)
	if templateId == "" {
		return nil
	}
	return assertDeletable(ctx, dispatcher,
		models.ProductTemplateFieldId, usagecheck.ResourceRef{
			ResourceName: usagecheck.ResourceProductTemplate,
			Identifier: usagecheck.NewIdentifier(
				templateId, readStringParam(params, basemodel.FieldOrgId)),
		}, vErrs)
}
