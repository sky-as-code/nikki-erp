package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// assertAttributeCreateAllowed is the create-time rule: only the display/mode pairing applies,
// since a new attribute has no template using it.
func assertAttributeCreateAllowed(
	_ corectx.Context, input *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	fields := input.GetFieldData()
	assertDisplayTypeFitsMode(
		composable.ParamString(fields, models.ProductAttributeFieldDisplayType),
		composable.ParamString(fields, models.ProductAttributeFieldVariantCreationMode), vErrs)
	return nil
}

// assertAttributeUpdateAllowed checks the merged state of stored and submitted fields, so an
// unused attribute may change display type and mode together in one save.
func assertAttributeUpdateAllowed(
	ctx corectx.Context, input *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	stored := found.GetFieldData()
	if archived, _ := composable.ParamBool(stored, basemodel.FieldIsArchived); archived {
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldIsArchived,
			ft.ErrorKey("product_attribute.archived_read_only", "inventory"),
			"this attribute is archived; unarchive it before editing"))
		return nil
	}
	submitted := input.GetFieldData()
	storedMode := composable.ParamString(stored, models.ProductAttributeFieldVariantCreationMode)
	finalMode := mergedString(submitted, stored, models.ProductAttributeFieldVariantCreationMode)
	finalDisplay := mergedString(submitted, stored, models.ProductAttributeFieldDisplayType)
	if assertDisplayTypeFitsMode(finalDisplay, finalMode, vErrs) {
		return nil
	}
	if finalMode == storedMode {
		return nil
	}
	attributeId := string(composable.ParamId(stored))
	lines, err := findAttributeLines(ctx, attributeId, util.ToPtr(true), usageScanLimit)
	if err != nil {
		return errors.Wrap(err, "assertAttributeUpdateAllowed")
	}
	templateIds := attrDistinctStrings(lines, models.ProductTemplateAttributeFieldProductTemplateId)
	if len(templateIds) > 0 {
		vErrs.Append(newTemplateUsageViolation(models.ProductAttributeFieldVariantCreationMode,
			"product_attribute.mode_locked",
			"{{count}} product template(s) apply this attribute ({{template_ids}}), so its variant creation "+
				"mode cannot change; create a new attribute instead", templateIds))
	}
	return nil
}

// assertDisplayTypeFitsMode refuses multi_checkbox outside mode never: several values picked at
// once describe no single variant. Reports whether it appended an error.
func assertDisplayTypeFitsMode(displayType string, mode string, vErrs *ft.ClientErrors) bool {
	if displayType != models.AttributeDisplayTypeMultiCheckbox.String() {
		return false
	}
	if mode == "" {
		mode = models.VariantCreationModeDynamic.String()
	}
	if mode == models.VariantCreationModeNever.String() {
		return false
	}
	vErrs.Append(*ft.NewValidationError(models.ProductAttributeFieldDisplayType,
		ft.ErrorKey("product_attribute.multi_checkbox_requires_never", "inventory"),
		"the multi checkbox display type is allowed only when variant creation mode is never"))
	return true
}

func mergedString(submitted, stored dmodel.DynamicFields, field string) string {
	if _, sent := submitted[field]; sent {
		return composable.ParamString(submitted, field)
	}
	return composable.ParamString(stored, field)
}

// assertAttributeNotInUse refuses to archive an attribute an enabled template line applies,
// archived templates included: unarchiving such a template would bring back a line on an
// archived attribute.
func assertAttributeNotInUse(ctx corectx.Context, attributeId string) (ft.ClientErrors, error) {
	if attributeId == "" {
		return nil, nil
	}
	lines, err := findAttributeLines(ctx, attributeId, util.ToPtr(true), usageScanLimit)
	if err != nil {
		return nil, errors.Wrap(err, "assertAttributeNotInUse")
	}
	templateIds := attrDistinctStrings(lines, models.ProductTemplateAttributeFieldProductTemplateId)
	if len(templateIds) == 0 {
		return nil, nil
	}
	return ft.ClientErrors{newTemplateUsageViolation(basemodel.FieldIsArchived, "product_attribute.in_use",
		"{{count}} product template(s) still apply this attribute ({{template_ids}}); remove it there first",
		templateIds)}, nil
}

// AssertProductAttributeDeletable refuses to delete an attribute any template line references,
// enabled or not: a disabled line is history variants were built from. lineRepo searches
// product_template_attribute.
func AssertProductAttributeDeletable(
	ctx corectx.Context, lineRepo models.ProductSearcher, attributeId string, vErrs *ft.ClientErrors,
) error {
	if attributeId == "" {
		return nil
	}
	found, err := lineRepo.Search(ctx, dyn.RepoSearchParam{Graph: attrAndGraph(
		attrCondition(models.ProductTemplateAttributeFieldAttributeId, dmodel.Equals, attributeId),
	), Fields: []string{basemodel.FieldId}, Page: 0, Size: 1})
	if err != nil {
		return errors.Wrap(err, "AssertProductAttributeDeletable")
	}
	if found.ClientErrors.Count() > 0 {
		return errors.Wrap(found.ClientErrors.ToError(), "AssertProductAttributeDeletable")
	}
	if found.HasData && len(found.Data.Items) > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductAttributeFieldId,
			ft.ErrorKey("product_attribute.has_template_history", "inventory"),
			"a product template has used this attribute, so it cannot be deleted; archive it instead"))
	}
	return nil
}
