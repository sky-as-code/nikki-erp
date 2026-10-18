package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func validateMappingCreate(
	_ corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	mapping := models.NewTaxMappingFrom(params)
	assertWellFormedPeriod(
		mapping.GetEffectiveFrom(), mapping.GetEffectiveTo(),
		models.TaxMappingFieldEffectiveTo, vErrs)
	return nil
}

func validateMappingUpdate(
	_ corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	found := entityFields(foundModel)
	assertLifecycleTransition(models.TaxMappingFieldLifecycleStatus, params, found, vErrs)
	assertMaterialFieldsImmutable(
		models.TaxMappingSchemaName, models.TaxMappingFieldLifecycleStatus, params, found, vErrs)
	if vErrs.Count() > 0 {
		return nil
	}

	mapping := models.NewTaxMappingFrom(mergeFields(found, params))
	assertWellFormedPeriod(
		mapping.GetEffectiveFrom(), mapping.GetEffectiveTo(),
		models.TaxMappingFieldEffectiveTo, vErrs)
	return nil
}

func validateMappingDelete(
	_ corectx.Context, _ *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	found := entityFields(foundModel)
	assertDeletableLifecycle(models.TaxMappingFieldLifecycleStatus, found, vErrs)
	return nil
}

// validateMappingLineCreate rejects a line that maps a tax onto itself: a no-op that reads as a
// configured rule.
func validateMappingLineCreate(
	_ corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	assertMappingLineSubstitutes(models.NewTaxMappingLineFrom(params), vErrs)
	return nil
}

func validateMappingLineUpdate(mappingRepo models.TaxSearcher) composable.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity,
		vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		found := entityFields(foundModel)
		if err := assertParentMappingEditable(ctx, mappingRepo, found, vErrs); err != nil {
			return err
		}
		if vErrs.Count() > 0 {
			return nil
		}
		assertMappingLineSubstitutes(models.NewTaxMappingLineFrom(mergeFields(found, params)), vErrs)
		return nil
	}
}

func assertMappingLineSubstitutes(line *models.TaxMappingLine, vErrs *ft.ClientErrors) {
	source, target := line.GetSourceTaxId(), line.GetTargetTaxId()
	if source == nil || target == nil {
		return
	}
	if *source == *target {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxMappingLineFieldTargetTaxId,
			ft.ErrorKey("tax.mapping_line_maps_to_itself", "accounting"),
			"a mapping line must substitute one tax for a different one"))
	}
}

// assertParentMappingEditable refuses to change a line of a published mapping. As with tax
// components, a line has no lifecycle of its own, so without this the mapping's immutability would
// be bypassable by editing what it substitutes.
func assertParentMappingEditable(
	ctx corectx.Context,
	repo models.TaxSearcher,
	found *dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) error {
	if found == nil {
		return nil
	}
	mappingId := models.NewTaxMappingLineFrom(*found).GetTaxMappingId()
	if mappingId == nil {
		return nil
	}

	mapping, err := models.FindMappingById(ctx, repo, *mappingId)
	if err != nil {
		return errors.Wrap(err, "assertParentMappingEditable")
	}
	if mapping == nil {
		return nil
	}

	status := mapping.GetLifecycleStatus()
	if status != nil && *status != string(models.LifecycleDraft) {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxMappingLineFieldTaxMappingId,
			ft.ErrorKey("tax.mapping_line_parent_not_draft", "accounting"),
			"the mapping this line belongs to is no longer a draft; "+
				"create a new mapping version to change what it substitutes"))
	}
	return nil
}

func validateRoundingPolicyCreate(
	_ corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	policy := models.NewTaxRoundingPolicyFrom(params)
	assertWellFormedPeriod(
		policy.GetEffectiveFrom(), policy.GetEffectiveTo(),
		models.TaxRoundingPolicyFieldEffectiveTo, vErrs)
	assertRoundingIncrementPositive(policy, vErrs)
	return nil
}

func validateRoundingPolicyUpdate(
	_ corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	found := entityFields(foundModel)
	assertLifecycleTransition(models.TaxRoundingPolicyFieldLifecycleStatus, params, found, vErrs)
	assertMaterialFieldsImmutable(
		models.TaxRoundingPolicySchemaName,
		models.TaxRoundingPolicyFieldLifecycleStatus, params, found, vErrs)
	if vErrs.Count() > 0 {
		return nil
	}

	policy := models.NewTaxRoundingPolicyFrom(mergeFields(found, params))
	assertWellFormedPeriod(
		policy.GetEffectiveFrom(), policy.GetEffectiveTo(),
		models.TaxRoundingPolicyFieldEffectiveTo, vErrs)
	assertRoundingIncrementPositive(policy, vErrs)
	return nil
}

func validateRoundingPolicyDelete(
	_ corectx.Context, _ *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	found := entityFields(foundModel)
	assertDeletableLifecycle(models.TaxRoundingPolicyFieldLifecycleStatus, found, vErrs)
	return nil
}

// assertRoundingIncrementPositive rejects a zero rounding quantum. The schema's minimum is zero
// because a decimal bound cannot express "greater than", but rounding to the nearest zero would
// divide by zero.
func assertRoundingIncrementPositive(
	policy *models.TaxRoundingPolicy, vErrs *ft.ClientErrors,
) {
	increment := policy.GetRoundingIncrement()
	if increment == nil {
		return
	}
	if !increment.IsPositive() {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxRoundingPolicyFieldRoundingIncrement,
			ft.ErrorKey("tax.rounding_increment_not_positive", "accounting"),
			"the rounding increment must be greater than zero"))
	}
}
