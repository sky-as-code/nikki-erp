package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// Clients never write an exclusion directly: the template's own configuration payload is the one
// write path, and it reaches this service internally. The rules still live here, so every writer
// gets them.

// NewProductTemplateCombinationExclusionDomainService is handed the composable default by the
// product_template_combination_exclusion onion.
func NewProductTemplateCombinationExclusionDomainService(
	base composable.CrudDomainService,
) itProduct.ProductTemplateCombinationExclusionDomainService {
	return &ProductTemplateCombinationExclusionDomainServiceImpl{CrudDomainService: base}
}

type ProductTemplateCombinationExclusionDomainServiceImpl struct {
	composable.CrudDomainService
}

// Create normalizes the pair before schema validation, so the unique index sees one spelling of
// {a, b}, then checks that the two values may be excluded from each other at all.
func (this *ProductTemplateCombinationExclusionDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.BeforeValidation = chainBeforeValidation(opts.BeforeValidation, normalizeExclusionEntity)
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, validateExclusionCreate)
	return this.CrudDomainService.Create(ctx, params, opts)
}

// Update refuses re-pointing an exclusion at other values: a different pair is a different rule,
// and the configuration writer archives one and creates the other.
func (this *ProductTemplateCombinationExclusionDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.UpdateOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, validateExclusionUpdate)
	return this.CrudDomainService.Update(ctx, params, opts)
}

const (
	exclusionFieldLeft  = models.ProductTemplateCombinationExclusionFieldLeftValueId
	exclusionFieldRight = models.ProductTemplateCombinationExclusionFieldRightValueId
)

// NormalizeExclusionPair orders a pair so the smaller id is on the left. ULIDs sort as strings.
func NormalizeExclusionPair(left string, right string) (string, string) {
	if right < left {
		return right, left
	}
	return left, right
}

func normalizeExclusionEntity(
	_ corectx.Context, entity *composable.DynamicEntity, _ *ft.ClientErrors,
) (*composable.DynamicEntity, error) {
	data := entity.GetFieldData()
	left, leftOk := data[exclusionFieldLeft].(string)
	right, rightOk := data[exclusionFieldRight].(string)
	if leftOk && rightOk {
		data[exclusionFieldLeft], data[exclusionFieldRight] = NormalizeExclusionPair(left, right)
	}
	return entity, nil
}

// ExclusionEnd is one side of an exclusion, resolved to the template and attribute it belongs to
// through its template line. An empty TemplateId means the value does not exist.
type ExclusionEnd struct {
	ValueId     string
	TemplateId  string
	AttributeId string
}

// AppendExclusionPairViolations checks a normalized pair against the rules the schema cannot
// express: two distinct values, both on this template, on two different attributes.
func AppendExclusionPairViolations(
	templateId string, left ExclusionEnd, right ExclusionEnd, vErrs *ft.ClientErrors,
) {
	if left.ValueId == right.ValueId {
		vErrs.Append(*ft.NewBusinessViolation(exclusionFieldRight,
			ft.ErrorKey("product_template_combination_exclusion.same_value", "inventory"),
			"a value cannot be excluded from itself"))
		return
	}
	for _, end := range []struct {
		field string
		value ExclusionEnd
	}{{exclusionFieldLeft, left}, {exclusionFieldRight, right}} {
		if end.value.TemplateId != templateId {
			vErrs.Append(*ft.NewBusinessViolation(end.field,
				ft.ErrorKey("product_template_combination_exclusion.value_not_in_template", "inventory"),
				"value '"+end.value.ValueId+"' is not a value of this template"))
		}
	}
	if vErrs.Count() == 0 && left.AttributeId == right.AttributeId {
		vErrs.Append(*ft.NewBusinessViolation(exclusionFieldRight,
			ft.ErrorKey("product_template_combination_exclusion.same_attribute", "inventory"),
			"an exclusion pairs values of two different attributes"))
	}
}

func validateExclusionCreate(
	ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	exclusion := models.NewProductTemplateCombinationExclusionFrom(inputModel.GetFieldData())
	templateId := derefId(exclusion.GetProductTemplateId())
	leftId, rightId := derefId(exclusion.GetLeftValueId()), derefId(exclusion.GetRightValueId())
	if templateId == "" || leftId == "" || rightId == "" {
		return nil
	}
	left, err := loadExclusionEnd(ctx, leftId)
	if err != nil {
		return err
	}
	right, err := loadExclusionEnd(ctx, rightId)
	if err != nil {
		return err
	}
	AppendExclusionPairViolations(templateId, left, right, vErrs)
	if vErrs.Count() > 0 {
		return nil
	}
	return assertExclusionNotDuplicated(ctx, inputModel.GetFieldData(), vErrs)
}

// loadExclusionEnd resolves a template value to its line's template and attribute.
func loadExclusionEnd(ctx corectx.Context, valueId string) (ExclusionEnd, error) {
	end := ExclusionEnd{ValueId: valueId}
	value, err := findRecord(ctx, models.ProductTemplateAttributeValueSchemaName,
		models.ProductTemplateAttributeValueFieldId, valueId)
	if err != nil || value == nil {
		return end, err
	}
	lineId := derefId(models.NewProductTemplateAttributeValueFrom(value).GetTemplateAttributeId())
	line, err := findRecord(ctx, models.ProductTemplateAttributeSchemaName, models.ProductTemplateAttributeFieldId, lineId)
	if err != nil || line == nil {
		return end, err
	}
	typed := models.NewProductTemplateAttributeFrom(line)
	end.TemplateId = derefId(typed.GetProductTemplateId())
	end.AttributeId = derefId(typed.GetAttributeId())
	return end, nil
}

// assertExclusionNotDuplicated reports the pair on a field before the unique index refuses it as
// an unnamed constraint error. Archived rows count: the index does not skip them, and a disabled
// rule is re-enabled by unarchiving it, not by recording it twice.
func assertExclusionNotDuplicated(ctx corectx.Context, data dmodel.DynamicFields, vErrs *ft.ClientErrors) error {
	templateField := models.ProductTemplateCombinationExclusionFieldProductTemplateId
	found, err := findRecordWhere(ctx, models.ProductTemplateCombinationExclusionSchemaName, dmodel.DynamicFields{
		templateField:       data[templateField],
		exclusionFieldLeft:  data[exclusionFieldLeft],
		exclusionFieldRight: data[exclusionFieldRight],
	})
	if err != nil {
		return errors.Wrap(err, "assertExclusionNotDuplicated")
	}
	if found != nil {
		vErrs.Append(*ft.NewBusinessViolation(exclusionFieldLeft,
			ft.ErrorKey("product_template_combination_exclusion.duplicate_pair", "inventory"),
			"this template already excludes the same pair of values"))
	}
	return nil
}

func validateExclusionUpdate(
	_ corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if foundModel == nil {
		return nil
	}
	input := models.NewProductTemplateCombinationExclusionFrom(inputModel.GetFieldData())
	stored := models.NewProductTemplateCombinationExclusionFrom(foundModel.GetFieldData())
	changes := map[string][2]string{
		exclusionFieldLeft:  {derefId(input.GetLeftValueId()), derefId(stored.GetLeftValueId())},
		exclusionFieldRight: {derefId(input.GetRightValueId()), derefId(stored.GetRightValueId())},
	}
	for field, pair := range changes {
		if pair[0] != "" && pair[0] != pair[1] {
			vErrs.Append(*ft.NewBusinessViolation(field,
				ft.ErrorKey("product_template_combination_exclusion.pair_immutable", "inventory"),
				"the values of an exclusion cannot be changed; archive it and add the new pair instead"))
		}
	}
	return nil
}
