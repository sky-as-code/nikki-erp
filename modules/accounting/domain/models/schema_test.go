package models

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// allSchemas builds every tax schema once, so an invariant that must hold across the module is
// asserted against the whole set rather than against whichever resource a test remembered.
func allSchemas(t *testing.T) map[string]*dmodel.ModelSchema {
	t.Helper()

	builders := map[string]*dmodel.ModelSchemaBuilder{
		TaxJurisdictionSchemaName:          TaxJurisdictionSchemaBuilder(),
		TaxGroupSchemaName:                 TaxGroupSchemaBuilder(),
		TaxRoundingPolicySchemaName:        TaxRoundingPolicySchemaBuilder(),
		TaxProductClassificationSchemaName: TaxProductClassificationSchemaBuilder(),
		TaxSchemaName:                      TaxSchemaBuilder(),
		TaxDefinitionVersionSchemaName:     TaxDefinitionVersionSchemaBuilder(),
		TaxRateVersionSchemaName:           TaxRateVersionSchemaBuilder(),
		TaxComponentSchemaName:             TaxComponentSchemaBuilder(),
		TaxMappingSchemaName:               TaxMappingSchemaBuilder(),
		TaxMappingLineSchemaName:           TaxMappingLineSchemaBuilder(),
		TaxRuleSchemaName:                  TaxRuleSchemaBuilder(),
		TaxRuleConditionSchemaName:         TaxRuleConditionSchemaBuilder(),
		TaxRuleResultSchemaName:            TaxRuleResultSchemaBuilder(),
	}

	// Build panics on a malformed schema rather than returning an error, which is what makes
	// TestEverySchemaBuilds meaningful: a bad data_type fails here instead of at application boot.
	schemas := make(map[string]*dmodel.ModelSchema, len(builders))
	for name, builder := range builders {
		schemas[name] = builder.Build()
	}
	return schemas
}

// Every schema's JSON parses and builds. This is the assertion that catches a data_type spelled
// with its canonical Go name rather than its JSON one — a mistake that otherwise panics at boot,
// long after the change that caused it.
func TestEverySchemaBuilds(t *testing.T) {
	if schemas := allSchemas(t); len(schemas) != 13 {
		t.Fatalf("expected 13 tax schemas, got %d", len(schemas))
	}
}

// AC-TAX-SUP-19: tax configuration belongs to an organization, and a resource from another org must
// never be resolved. The enforcement is the org_base_model mixin, which injects org_id and is what
// the engine scopes every query by — so a schema that lost the mixin would silently serve another
// organization's rates, with nothing in the code that reads it looking wrong.
func TestEverySchemaIsOrgScoped(t *testing.T) {
	for name, schema := range allSchemas(t) {
		if _, present := schema.Fields()[basemodel.FieldOrgId]; !present {
			t.Errorf("schema %q has no %s: it would not be scoped to an organization",
				name, basemodel.FieldOrgId)
		}
	}
}

// The schema name is the resource code the engine asserts permissions against and the path segment
// its routes hang off, so a name that drifts from the "accounting_" prefix breaks the IAM seed and
// the URL at once, in a way that reads as a permission problem rather than a naming one.
func TestSchemaNamesCarryTheModulePrefix(t *testing.T) {
	for name, schema := range allSchemas(t) {
		if schema.Name() != name {
			t.Errorf("schema built as %q but registered as %q", schema.Name(), name)
		}
		if len(name) < 11 || name[:11] != "accounting_" {
			t.Errorf("schema %q does not carry the accounting_ prefix", name)
		}
	}
}

// Every resource carries the audit trail the base model provides. A snapshot proves what was
// charged; these prove who changed the configuration behind it and when.
func TestEverySchemaHasTheBaseAuditFields(t *testing.T) {
	for name, schema := range allSchemas(t) {
		fields := schema.Fields()
		for _, required := range []string{basemodel.FieldId, basemodel.FieldCreatedAt, basemodel.FieldEtag} {
			if _, present := fields[required]; !present {
				t.Errorf("schema %q is missing the base field %q", name, required)
			}
		}
	}
}

// The five versioned resources are the ones the engine's lifecycle rules govern. A lifecycle field
// missing here would leave a resource publishable in name only — its material fields never freezing,
// because the rules key off exactly this field.
func TestVersionedResourcesCarryALifecycleStatus(t *testing.T) {
	lifecycleSchemas := map[string]string{
		TaxDefinitionVersionSchemaName: TaxDefinitionVersionFieldLifecycleStatus,
		TaxRateVersionSchemaName:       TaxRateVersionFieldLifecycleStatus,
		TaxRoundingPolicySchemaName:    TaxRoundingPolicyFieldLifecycleStatus,
		TaxMappingSchemaName:           TaxMappingFieldLifecycleStatus,
		TaxRuleSchemaName:              TaxRuleFieldLifecycleStatus,
	}

	schemas := allSchemas(t)
	for name, statusField := range lifecycleSchemas {
		if _, present := schemas[name].Fields()[statusField]; !present {
			t.Errorf("versioned schema %q has no %q field", name, statusField)
		}
	}
}

// Effective dating is what makes a rate change not reinterpret a historical sale, so every
// versioned resource has to be able to say when it applies from.
func TestVersionedResourcesCarryAnEffectiveFrom(t *testing.T) {
	effectiveFrom := map[string]string{
		TaxDefinitionVersionSchemaName: TaxDefinitionVersionFieldEffectiveFrom,
		TaxRateVersionSchemaName:       TaxRateVersionFieldEffectiveFrom,
		TaxRoundingPolicySchemaName:    TaxRoundingPolicyFieldEffectiveFrom,
		TaxMappingSchemaName:           TaxMappingFieldEffectiveFrom,
		TaxRuleSchemaName:              TaxRuleFieldEffectiveFrom,
	}

	schemas := allSchemas(t)
	for name, field := range effectiveFrom {
		if _, present := schemas[name].Fields()[field]; !present {
			t.Errorf("versioned schema %q has no %q field", name, field)
		}
	}
}

// Money and rates must never be float-typed. A float64 cannot hold a decimal fraction exactly, and
// a rate stored as one would make every amount computed from it subtly wrong — the kind of defect
// that shows up as a one-cent discrepancy an auditor finds years later.
func TestMonetaryFieldsAreDecimal(t *testing.T) {
	schemas := allSchemas(t)

	monetary := map[string][]string{
		TaxRateVersionSchemaName: {
			TaxRateVersionFieldRate,
			TaxRateVersionFieldFixedAmount,
		},
		TaxRoundingPolicySchemaName: {
			TaxRoundingPolicyFieldRoundingIncrement,
		},
	}

	for schemaName, fieldNames := range monetary {
		for _, fieldName := range fieldNames {
			field, present := schemas[schemaName].Fields()[fieldName]
			if !present {
				t.Errorf("schema %q has no field %q", schemaName, fieldName)
				continue
			}
			if got := field.DataType().String(); got != dmodel.FieldDataTypeNameDecimal {
				t.Errorf("%s.%s is %q, but money and rates must be decimal",
					schemaName, fieldName, got)
			}
		}
	}
}
