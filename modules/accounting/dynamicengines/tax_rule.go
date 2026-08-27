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
	_ corectx.Context, params dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	rule := models.NewTaxRuleFrom(params)
	assertWellFormedPeriod(
		rule.GetEffectiveFrom(), rule.GetEffectiveTo(), models.TaxRuleFieldEffectiveTo, vErrs)
	return nil
}

func validateRuleUpdate(
	_ corectx.Context, params dmodel.DynamicFields, found *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
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
	_ corectx.Context, _ dmodel.DynamicFields, found *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
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
	ctx corectx.Context, params dmodel.DynamicFields, found *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	return validateRuleConditionWrite(ctx, mergeFields(found, params), found, vErrs)
}

// validateRuleConditionWrite enforces the typed-condition contract of BR-TAX-ESS-SUP-007.
//
// Two things are checked here that the schema cannot state: that the field key is one the engine
// actually knows how to read from the tax context, and that the operator and the value agree about
// arity. The whitelist matters most — it is what keeps a condition a declarative comparison rather
// than an expression language, which the requirement forbids outright.
func validateRuleConditionWrite(
	_ corectx.Context, params dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
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
	ctx corectx.Context, params dmodel.DynamicFields, found *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	return validateRuleResultWrite(ctx, mergeFields(found, params), found, vErrs)
}

// validateRuleResultWrite enforces the per-action required fields of BR-TAX-ESS-SUP-008.
//
// Each action needs a different companion field and JSON cannot express a conditional requirement,
// so all three are nullable in the schema and the pairing is checked here. A result missing its
// field is not a harmless no-op: an add_tax with no tax silently contributes nothing, and the rule
// appears to have matched while changing nothing at all.
func validateRuleResultWrite(
	_ corectx.Context, params dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
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
