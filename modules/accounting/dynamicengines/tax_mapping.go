package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

func taxMappingEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxMappingSchemaName,
		DefineActions: defineMappingActions,
	}
}

func defineMappingActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateMappingCreate,
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax mapping create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   mappingKeysToFetch,
		ValidateExtra: validateMappingUpdate,
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax mapping update validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionDelete,
		KeysToFetch:   mappingKeysToFetch,
		ValidateExtra: validateMappingDelete,
	})
	return errors.Wrap(err, "failed to attach tax mapping delete validation")
}

func mappingKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.TaxMappingFieldId: params[models.TaxMappingFieldId]}
}

func validateMappingCreate(
	_ corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	mapping := models.NewTaxMappingFrom(params)
	assertWellFormedPeriod(
		mapping.GetEffectiveFrom(), mapping.GetEffectiveTo(),
		models.TaxMappingFieldEffectiveTo, vErrs)
	return nil
}

func validateMappingUpdate(
	_ corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
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
	_ corectx.Context, _ *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	found := entityFields(foundModel)
	assertDeletableLifecycle(models.TaxMappingFieldLifecycleStatus, found, vErrs)
	return nil
}

func defineMappingLineActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateMappingLineCreate,
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax mapping line create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   mappingLineKeysToFetch,
		ValidateExtra: validateMappingLineUpdate(engine),
	})
	return errors.Wrap(err, "failed to attach tax mapping line update validation")
}

func mappingLineKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.TaxMappingLineFieldId: params[models.TaxMappingLineFieldId],
	}
}

// validateMappingLineCreate rejects a line that maps a tax onto itself.
//
// Substituting a tax for itself is a no-op that reads as a configured rule, so an author debugging
// why an export order still carries domestic VAT would find a mapping that appears to handle it.
func validateMappingLineCreate(
	_ corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	assertMappingLineSubstitutes(models.NewTaxMappingLineFrom(params), vErrs)
	return nil
}

func validateMappingLineUpdate(engine drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		params := inputModel.GetFieldData()
		found := entityFields(foundModel)
		if err := assertParentMappingEditable(ctx, engine, found, vErrs); err != nil {
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
			"tax.mapping_line_maps_to_itself",
			"a mapping line must substitute one tax for a different one"))
	}
}

// assertParentMappingEditable refuses to change a line of a published mapping.
//
// Same reasoning as tax components: a line has no lifecycle of its own, so without this the
// mapping's own immutability would be bypassable by editing what it substitutes rather than the
// mapping row itself (BR-TAX-ESS-SUP-031).
func assertParentMappingEditable(
	ctx corectx.Context,
	engine drif.DynamicResourceEngine,
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

	mapping, err := models.FindMappingById(ctx, engine.ResourceRepository(), *mappingId)
	if err != nil {
		return errors.Wrap(err, "assertParentMappingEditable")
	}
	if mapping == nil {
		return nil
	}

	status := mapping.GetLifecycleStatus()
	if status != nil && *status != string(models.LifecycleDraft) {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxMappingLineFieldTaxMappingId,
			"tax.mapping_line_parent_not_draft",
			"the mapping this line belongs to is no longer a draft; "+
				"create a new mapping version to change what it substitutes"))
	}
	return nil
}

func taxRoundingPolicyEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxRoundingPolicySchemaName,
		DefineActions: defineRoundingPolicyActions,
	}
}

func defineRoundingPolicyActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateRoundingPolicyCreate,
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax rounding policy create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   roundingPolicyKeysToFetch,
		ValidateExtra: validateRoundingPolicyUpdate,
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax rounding policy update validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionDelete,
		KeysToFetch:   roundingPolicyKeysToFetch,
		ValidateExtra: validateRoundingPolicyDelete,
	})
	return errors.Wrap(err, "failed to attach tax rounding policy delete validation")
}

func roundingPolicyKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.TaxRoundingPolicyFieldId: params[models.TaxRoundingPolicyFieldId],
	}
}

func validateRoundingPolicyCreate(
	_ corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
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
	_ corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
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
	_ corectx.Context, _ *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	found := entityFields(foundModel)
	assertDeletableLifecycle(models.TaxRoundingPolicyFieldLifecycleStatus, found, vErrs)
	return nil
}

// assertRoundingIncrementPositive rejects a zero rounding quantum.
//
// The schema's minimum is zero because a decimal bound cannot express "greater than", but rounding
// to the nearest zero has no meaning and would divide by it. Zero is the value a partly-filled form
// produces, so it is worth a message that says what is wrong rather than a later arithmetic error.
func assertRoundingIncrementPositive(
	policy *models.TaxRoundingPolicy, vErrs *ft.ClientErrors,
) {
	increment := policy.GetRoundingIncrement()
	if increment == nil {
		return
	}
	if !increment.IsPositive() {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxRoundingPolicyFieldRoundingIncrement,
			"tax.rounding_increment_not_positive",
			"the rounding increment must be greater than zero"))
	}
}
