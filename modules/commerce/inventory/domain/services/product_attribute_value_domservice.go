package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewProductAttributeValueDomainService is handed the composable default by the
// product_attribute_value onion.
func NewProductAttributeValueDomainService(base composable.CrudDomainService) itProduct.ProductAttributeValueDomainService {
	return &ProductAttributeValueDomainServiceImpl{CrudDomainService: base}
}

// ProductAttributeValueDomainServiceImpl holds the value rules. The attribute's inline values
// delta writes through this service too, so a value written either way obeys the same rules.
type ProductAttributeValueDomainServiceImpl struct {
	composable.CrudDomainService
}

// Create trims the name and refuses a blank one, one that clashes with a sibling, and a value
// under an archived attribute.
func (this *ProductAttributeValueDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	trimLangJsonInPlace(cmd, models.ProductAttributeValueFieldName)
	opts := attrFirstOption(options)
	opts.ValidateExtra = composable.ChainValidateExtra(assertValueWritable, opts.ValidateExtra)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

// Update applies the same rules as Create; the name rules only when the name changes.
func (this *ProductAttributeValueDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	trimLangJsonInPlace(cmd, models.ProductAttributeValueFieldName)
	opts := attrFirstOption(options)
	opts.ValidateExtra = composable.ChainValidateExtra(assertValueWritable, opts.ValidateExtra)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// Delete refuses a value any template line ever offered: variants were built from that row, so
// the value can only be archived.
func (this *ProductAttributeValueDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	opts := attrFirstOption(options)
	opts.ValidateExtra = composable.ChainValidateExtra(assertValueHasNoHistory, opts.ValidateExtra)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

// SetArchived refuses to archive a value an enabled template line still selects, and to
// unarchive one whose attribute is archived.
func (this *ProductAttributeValueDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd composable.SetArchivedCommand,
) (*composable.MutateResult, error) {
	archive, _ := composable.ParamBool(cmd, basemodel.FieldIsArchived)
	valueId := string(composable.ParamId(cmd))
	var cErrs ft.ClientErrors
	var err error
	if archive {
		cErrs, err = assertValueNotSelected(ctx, valueId)
	} else {
		cErrs, err = this.assertParentActive(ctx, valueId)
	}
	if err != nil {
		return nil, errors.Wrap(err, "SetArchived product attribute value")
	}
	if cErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: cErrs}, nil
	}
	return this.CrudDomainService.SetArchived(ctx, cmd)
}

// assertValueWritable stops at the first refusal: a value of an archived attribute is read-only,
// so its name is not worth checking.
func assertValueWritable(
	ctx corectx.Context, input *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	attributeId := composable.ParamString(input.GetFieldData(), models.ProductAttributeValueFieldAttributeId)
	if found != nil {
		attributeId = composable.ParamString(found.GetFieldData(), models.ProductAttributeValueFieldAttributeId)
	}
	if attributeId != "" {
		archived, err := isAttributeArchived(ctx, model.Id(attributeId))
		if err != nil {
			return errors.Wrap(err, "assertValueWritable")
		}
		if archived {
			vErrs.Append(*newValueParentArchivedError(models.ProductAttributeValueFieldAttributeId))
			return nil
		}
	}
	if name := input.GetFieldData().GetLangJson(models.ProductAttributeValueFieldName); name != nil &&
		isLangJsonBlank(*name) {
		vErrs.Append(*newValueNameBlankError(models.ProductAttributeValueFieldName))
		return nil
	}
	return assertValueNameUnique(ctx, input, found, vErrs)
}

func newValueParentArchivedError(field string) *ft.ClientErrorItem {
	return ft.NewBusinessViolation(field, ft.ErrorKey("product_attribute_value.parent_archived", "inventory"),
		"the attribute of this value is archived; unarchive the attribute first")
}

func assertValueNameUnique(
	ctx corectx.Context, input *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	fields := input.GetFieldData()
	name := fields.GetLangJson(models.ProductAttributeValueFieldName)
	if name == nil {
		return nil
	}
	candidate := attributeValueName{name: *name}
	attributeId := composable.ParamString(fields, models.ProductAttributeValueFieldAttributeId)
	if found != nil {
		candidate.id = composable.ParamId(found.GetFieldData())
		attributeId = composable.ParamString(found.GetFieldData(), models.ProductAttributeValueFieldAttributeId)
	}
	if attributeId == "" {
		return nil
	}
	siblings, err := loadAttributeValueNames(ctx, attributeId)
	if err != nil {
		return errors.Wrap(err, "assertValueNameUnique")
	}
	if clash := findNameClash(candidate, siblings); clash != nil {
		vErrs.Append(*newValueNameDuplicateError(models.ProductAttributeValueFieldName, *clash))
	}
	return nil
}

func assertValueHasNoHistory(
	ctx corectx.Context, _ *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	valueId := string(composable.ParamId(found.GetFieldData()))
	if valueId == "" {
		return nil
	}
	hasHistory, err := valueHasTemplateHistory(ctx, valueId)
	if err != nil {
		return errors.Wrap(err, "assertValueHasNoHistory")
	}
	if hasHistory {
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldId,
			ft.ErrorKey("product_attribute_value.has_history", "inventory"),
			"a product template has used this value, so it cannot be deleted; archive it instead"))
	}
	return nil
}

func assertValueNotSelected(ctx corectx.Context, valueId string) (ft.ClientErrors, error) {
	if valueId == "" {
		return nil, nil
	}
	templateIds, err := findTemplatesSelectingValue(ctx, valueId)
	if err != nil || len(templateIds) == 0 {
		return nil, err
	}
	return ft.ClientErrors{newTemplateUsageViolation(basemodel.FieldIsArchived,
		"product_attribute_value.in_use",
		"{{count}} product template(s) still select this value ({{template_ids}}); unselect it there first",
		templateIds)}, nil
}

// assertParentActive refuses to bring a value back under an archived attribute: an unarchived
// value of an archived attribute could be offered by nothing.
func (this *ProductAttributeValueDomainServiceImpl) assertParentActive(
	ctx corectx.Context, valueId string,
) (ft.ClientErrors, error) {
	if valueId == "" {
		return nil, nil
	}
	values, err := searchAttributeRows(ctx, models.ProductAttributeValueSchemaName,
		attrAndGraph(attrCondition(basemodel.FieldId, dmodel.Equals, valueId)),
		[]string{basemodel.FieldId, models.ProductAttributeValueFieldAttributeId}, 1)
	if err != nil || len(values) == 0 {
		return nil, err
	}
	attributeId := composable.ParamString(values[0], models.ProductAttributeValueFieldAttributeId)
	archived, err := isAttributeArchived(ctx, model.Id(attributeId))
	if err != nil || !archived {
		return nil, err
	}
	return ft.ClientErrors{*newValueParentArchivedError(basemodel.FieldIsArchived)}, nil
}

func isAttributeArchived(ctx corectx.Context, attributeId model.Id) (bool, error) {
	rows, err := searchAttributeRows(ctx, models.ProductAttributeSchemaName,
		attrAndGraph(attrCondition(basemodel.FieldId, dmodel.Equals, string(attributeId))),
		[]string{basemodel.FieldId, basemodel.FieldIsArchived}, 1)
	if err != nil || len(rows) == 0 {
		return false, err
	}
	archived, _ := composable.ParamBool(rows[0], basemodel.FieldIsArchived)
	return archived, nil
}

func attrFirstOption[T any](options []T) T {
	var zero T
	if len(options) > 0 {
		return options[0]
	}
	return zero
}
