package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

func taxRuleEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxRuleSchemaName,
		DefineActions: defineRuleActions,
	}
}

func defineRuleActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateRuleCreate,
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax rule create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   ruleKeysToFetch,
		ValidateExtra: validateRuleUpdate,
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax rule update validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionDelete,
		KeysToFetch:   ruleKeysToFetch,
		ValidateExtra: validateRuleDelete,
	})
	return errors.Wrap(err, "failed to attach tax rule delete validation")
}

func ruleKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.TaxRuleFieldId: params[models.TaxRuleFieldId]}
}

func validateRuleCreate(
	_ corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	rule := models.NewTaxRuleFrom(params)
	assertWellFormedPeriod(
		rule.GetEffectiveFrom(), rule.GetEffectiveTo(), models.TaxRuleFieldEffectiveTo, vErrs)
	return nil
}

func validateRuleUpdate(
	_ corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	found := entityFields(foundModel)
	assertLifecycleTransition(models.TaxRuleFieldLifecycleStatus, params, found, vErrs)
	assertMaterialFieldsImmutable(
		models.TaxRuleSchemaName, models.TaxRuleFieldLifecycleStatus, params, found, vErrs)
	if vErrs.Count() > 0 {
		return nil
	}

	rule := models.NewTaxRuleFrom(mergeFields(found, params))
	assertWellFormedPeriod(
		rule.GetEffectiveFrom(), rule.GetEffectiveTo(), models.TaxRuleFieldEffectiveTo, vErrs)
	return nil
}

func validateRuleDelete(
	_ corectx.Context, _ *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	found := entityFields(foundModel)
	assertDeletableLifecycle(models.TaxRuleFieldLifecycleStatus, found, vErrs)
	return nil
}

func taxRuleConditionEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxRuleConditionSchemaName,
		DefineActions: defineRuleConditionActions,
	}
}

func defineRuleConditionActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateRuleConditionWrite,
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax rule condition create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   ruleConditionKeysToFetch,
		ValidateExtra: validateRuleConditionUpdate,
	})
	return errors.Wrap(err, "failed to attach tax rule condition update validation")
}

func ruleConditionKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.TaxRuleConditionFieldId: params[models.TaxRuleConditionFieldId],
	}
}

func validateRuleConditionUpdate(
	ctx corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	found := entityFields(foundModel)
	return validateRuleConditionWrite(ctx, drif.NewDynamicEntityFrom(mergeFields(found, params)), foundModel, vErrs)
}

// validateRuleConditionWrite enforces the typed-condition contract: the field key must be one the
// engine can read from the tax context, and the operator and value must agree about arity. The
// whitelist is what keeps a condition a declarative comparison rather than an expression language.
func validateRuleConditionWrite(
	_ corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	condition := models.NewTaxRuleConditionFrom(params)

	fieldKey := condition.GetFieldKey()
	if fieldKey != nil && !models.IsKnownContextField(*fieldKey) {
		vErrs.Append(*ft.NewBusinessViolation(models.TaxRuleConditionFieldFieldKey,
			"tax.unknown_condition_field",
			"'"+*fieldKey+"' is not a field of the tax context that a rule can test"))
	}

	rawOp := condition.GetOperator()
	if rawOp == nil {
		return nil
	}
	operator := models.WrapConditionOperator(*rawOp)
	if operator == nil {
		return nil // the enum constraint already rejects an unknown operator
	}

	value := condition.GetValue()
	switch {
	case models.IsNullaryOperator(*operator):
		if value != nil {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxRuleConditionFieldValue,
				"tax.nullary_operator_takes_no_value",
				"the operator '"+string(*operator)+"' tests presence and takes no comparison value"))
		}
	case models.IsArrayOperator(*operator):
		if _, isArray := value.([]any); !isArray {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxRuleConditionFieldValue,
				"tax.array_operator_requires_list",
				"the operator '"+string(*operator)+"' compares against a list of values"))
		}
	default:
		if value == nil {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxRuleConditionFieldValue,
				"tax.operator_requires_value",
				"the operator '"+string(*operator)+"' needs a comparison value"))
		}
	}
	return nil
}

func taxRuleResultEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.TaxRuleResultSchemaName,
		DefineActions: defineRuleResultActions,
	}
}

func defineRuleResultActions(engine drif.DynamicResourceEngine) error {
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionCreate,
		ValidateExtra: validateRuleResultWrite,
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach tax rule result create validation")
	}

	err = engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionUpdate,
		KeysToFetch:   ruleResultKeysToFetch,
		ValidateExtra: validateRuleResultUpdate,
	})
	return errors.Wrap(err, "failed to attach tax rule result update validation")
}

func ruleResultKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.TaxRuleResultFieldId: params[models.TaxRuleResultFieldId],
	}
}

func validateRuleResultUpdate(
	ctx corectx.Context, inputModel *drif.DynamicEntity, foundModel *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	found := entityFields(foundModel)
	return validateRuleResultWrite(ctx, drif.NewDynamicEntityFrom(mergeFields(found, params)), foundModel, vErrs)
}

// validateRuleResultWrite enforces the per-action required fields. JSON cannot express a
// conditional requirement, so all three companion fields are nullable in the schema and the pairing
// is checked here: an add_tax with no tax would appear to match while contributing nothing.
func validateRuleResultWrite(
	_ corectx.Context, inputModel *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	params := inputModel.GetFieldData()
	result := models.NewTaxRuleResultFrom(params)
	rawAction := result.GetAction()
	if rawAction == nil {
		return nil
	}
	action := models.WrapRuleResultAction(*rawAction)
	if action == nil {
		return nil
	}

	switch *action {
	case models.ActionAddTax, models.ActionRemoveTax:
		if result.GetTaxId() == nil {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxRuleResultFieldTaxId,
				"tax.rule_result_requires_tax",
				"the action '"+string(*action)+"' must name the tax it applies to"))
		}
	case models.ActionApplyMapping:
		if result.GetTaxMappingId() == nil {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxRuleResultFieldTaxMappingId,
				"tax.rule_result_requires_mapping",
				"the action 'apply_mapping' must name the mapping to apply"))
		}
	case models.ActionNoTaxApplicable:
		if result.GetTaxTreatment() == nil {
			vErrs.Append(*ft.NewBusinessViolation(models.TaxRuleResultFieldTaxTreatment,
				"tax.rule_result_requires_treatment",
				"the action 'no_tax_applicable' must state whether the supply is "+
					"non_taxable or out_of_scope"))
		}
	}
	return nil
}
