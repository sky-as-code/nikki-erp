package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	modconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// AssertTemplateNotUsedExternally asks every dependant whether it still names this template.
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
	return usagecheck.AssertDeletable(ctx, dispatcher, modconstants.InventoryModuleName,
		models.ProductTemplateFieldId, usagecheck.ResourceRef{
			ResourceName: usagecheck.ResourceProductTemplate,
			Identifier: usagecheck.NewIdentifier(
				templateId, readStringParam(params, basemodel.FieldOrgId)),
		}, vErrs)
}
