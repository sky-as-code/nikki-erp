package dynamicengines

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// The lifecycle rules shared by every versioned tax configuration.
//
// Four resources carry a lifecycle_status — definition versions, rate versions, rules and mappings
// — plus rounding policies. They all obey the same three rules, so the rules live here once rather
// than in four near-identical copies that would drift the first time one of them gained a case.
//
// The rules (BR-TAX-ESS-SUP-002, SUP-027):
//
//  1. draft may be edited freely and deleted.
//  2. published may not have a material field changed, and may not be deleted. A material field is
//     anything that could alter what a past calculation produced.
//  3. withdrawn may not return to draft, and may not be deleted.
//
// Why immutability is enforced here rather than by trusting callers: a published rate is cited by
// name and version in every Tax Snapshot calculated against it. Editing 8% to 10% in place would
// silently restate what every one of those historical documents says was charged, and there is no
// way to detect it afterwards — the snapshot and the master would simply agree on the wrong number.

// materialFieldsBySchema lists the fields that freeze on publication, per resource.
//
// Everything that feeds determination, the formula, the rate, the base, the treatment, the
// jurisdiction or the applicable period is material. Descriptive fields — a name, a description,
// a legal reference correcting a typo — are not, so that fixing a label does not force a new
// version nobody needs.
var materialFieldsBySchema = map[string][]string{
	models.TaxDefinitionVersionSchemaName: {
		models.TaxDefinitionVersionFieldTaxId,
		models.TaxDefinitionVersionFieldUsage,
		models.TaxDefinitionVersionFieldJurisdictionId,
		models.TaxDefinitionVersionFieldTaxGroupId,
		models.TaxDefinitionVersionFieldCalculationType,
		models.TaxDefinitionVersionFieldTaxTreatment,
		models.TaxDefinitionVersionFieldPriceInclusionMode,
		models.TaxDefinitionVersionFieldSequence,
		models.TaxDefinitionVersionFieldAffectSubsequentBase,
		models.TaxDefinitionVersionFieldBaseAffectedByPrevious,
		models.TaxDefinitionVersionFieldEffectiveFrom,
		models.TaxDefinitionVersionFieldEffectiveTo,
		models.TaxDefinitionVersionFieldVersionNo,
	},
	models.TaxRateVersionSchemaName: {
		models.TaxRateVersionFieldTaxId,
		models.TaxRateVersionFieldRate,
		models.TaxRateVersionFieldFixedAmount,
		models.TaxRateVersionFieldCurrencyCode,
		models.TaxRateVersionFieldRateUomId,
		models.TaxRateVersionFieldEffectiveFrom,
		models.TaxRateVersionFieldEffectiveTo,
		models.TaxRateVersionFieldVersionNo,
	},
	models.TaxRuleSchemaName: {
		models.TaxRuleFieldJurisdictionId,
		models.TaxRuleFieldPriority,
		models.TaxRuleFieldStopProcessing,
		models.TaxRuleFieldEffectiveFrom,
		models.TaxRuleFieldEffectiveTo,
		models.TaxRuleFieldVersionNo,
	},
	models.TaxMappingSchemaName: {
		models.TaxMappingFieldPriority,
		models.TaxMappingFieldEffectiveFrom,
		models.TaxMappingFieldEffectiveTo,
		models.TaxMappingFieldVersionNo,
	},
	models.TaxRoundingPolicySchemaName: {
		models.TaxRoundingPolicyFieldJurisdictionId,
		models.TaxRoundingPolicyFieldCurrencyCode,
		models.TaxRoundingPolicyFieldRoundingScope,
		models.TaxRoundingPolicyFieldRoundingMethod,
		models.TaxRoundingPolicyFieldRoundingIncrement,
		models.TaxRoundingPolicyFieldEffectiveFrom,
		models.TaxRoundingPolicyFieldEffectiveTo,
		models.TaxRoundingPolicyFieldVersionNo,
	},
}

// lifecycleStatusOf reads the lifecycle status out of a stored row, or nil when absent.
func lifecycleStatusOf(found *dmodel.DynamicFields, field string) *models.LifecycleStatus {
	if found == nil {
		return nil
	}
	raw := found.GetString(field)
	if raw == nil {
		return nil
	}
	return models.WrapLifecycleStatus(*raw)
}

// assertMaterialFieldsImmutable rejects an edit to a material field of a published record.
//
// It reports every offending field rather than stopping at the first, so that a user editing a
// form is told everything they must undo in one response instead of discovering it a field at a
// time.
func assertMaterialFieldsImmutable(
	schemaName string,
	statusField string,
	params dmodel.DynamicFields,
	found *dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) {
	status := lifecycleStatusOf(found, statusField)
	if status == nil || *status != models.LifecyclePublished {
		return
	}

	for _, field := range materialFieldsBySchema[schemaName] {
		if _, submitted := params[field]; !submitted {
			continue
		}
		vErrs.Append(*ft.NewBusinessViolation(field, "tax.published_field_immutable",
			"this configuration is published; create a new version instead of changing "+
				"a field that decides how tax is calculated"))
	}
}

// assertLifecycleTransition rejects a status change that the state machine does not allow.
//
// Only two transitions are legal, and both move forward: draft to published, and either of those
// to withdrawn. Everything else is refused, most importantly withdrawn back to draft — a withdrawn
// configuration is kept for audit, and letting it become editable again would reopen exactly the
// history that withdrawing it was meant to close.
func assertLifecycleTransition(
	statusField string,
	params dmodel.DynamicFields,
	found *dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) {
	raw := params.GetString(statusField)
	if raw == nil {
		return
	}
	next := models.WrapLifecycleStatus(*raw)
	if next == nil {
		return // the enum constraint in the schema already rejects an unknown value
	}
	current := lifecycleStatusOf(found, statusField)
	if current == nil || *current == *next {
		return
	}

	allowed := false
	switch *current {
	case models.LifecycleDraft:
		allowed = *next == models.LifecyclePublished || *next == models.LifecycleWithdrawn
	case models.LifecyclePublished:
		allowed = *next == models.LifecycleWithdrawn
	case models.LifecycleWithdrawn:
		allowed = false
	}

	if !allowed {
		vErrs.Append(*ft.NewBusinessViolation(statusField, "tax.invalid_lifecycle_transition",
			"a tax configuration cannot move from "+string(*current)+" to "+string(*next)))
	}
}

// assertDeletableLifecycle rejects deletion of anything that has ever been published.
//
// Note what this does NOT do: ask any downstream module whether the configuration is in use.
// BR-TAX-ESS-SUP-026 replaced that question with this one deliberately. Tax cannot enumerate its
// consumers without depending on them, and a delete permitted because no consumer answered would
// be exactly as destructive as one permitted because nobody asked. Publication is the boundary Tax
// owns, so publication is the boundary it enforces.
func assertDeletableLifecycle(
	statusField string, found *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) {
	status := lifecycleStatusOf(found, statusField)
	if status == nil || *status == models.LifecycleDraft {
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(statusField, "tax.published_not_deletable",
		"a configuration that has been published cannot be deleted; withdraw or archive it instead"))
}
