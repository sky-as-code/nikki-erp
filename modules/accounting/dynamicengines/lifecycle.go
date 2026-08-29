package dynamicengines

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// The lifecycle rules shared by every versioned tax configuration: draft may be edited and deleted;
// published may not have a material field changed and may not be deleted (a material field is
// anything that could alter what a past calculation produced); withdrawn may not return to draft
// and may not be deleted.
//
// Immutability is enforced here rather than trusted to callers because a published rate is cited by
// name and version in every Tax Snapshot calculated against it, so an in-place edit would silently
// restate what those historical documents charged, undetectably.

// materialFieldsBySchema lists the fields that freeze on publication, per resource. Anything
// feeding determination, the formula, rate, base, treatment, jurisdiction or applicable period is
// material; purely descriptive fields are not.
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

// assertMaterialFieldsImmutable rejects an edit to a material field of a published record. It
// reports every offending field rather than stopping at the first.
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

// assertLifecycleTransition rejects a status change the state machine disallows. Only forward moves
// are legal: draft to published, and either to withdrawn. Withdrawn never returns to draft, which
// would reopen the history withdrawing it closed.
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

// assertDeletableLifecycle rejects deletion of anything that has ever been published. It
// deliberately does not ask a downstream module whether the configuration is in use: Tax cannot
// enumerate its consumers without depending on them, so publication is the boundary it enforces.
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
