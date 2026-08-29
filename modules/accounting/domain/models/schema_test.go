package models

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// allSchemas builds every tax schema once, so a module-wide invariant is asserted against the whole
// set rather than whichever resource a test remembered.
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

	// Build panics on a malformed schema rather than returning an error, so a bad data_type fails
	// here instead of at application boot.
	schemas := make(map[string]*dmodel.ModelSchema, len(builders))
	for name, builder := range builders {
		schemas[name] = builder.Build()
	}
	return schemas
}

// Every schema's JSON parses and builds. Catches a data_type spelled with its canonical Go name
// rather than its JSON one, which otherwise panics at boot.
func TestEverySchemaBuilds(t *testing.T) {
	if schemas := allSchemas(t); len(schemas) != 13 {
		t.Fatalf("expected 13 tax schemas, got %d", len(schemas))
	}
}

// Tax configuration belongs to an organization. The org_base_model mixin injects the org_id the
// engine scopes every query by, so a schema that lost it would silently serve another
// organization's rates.
func TestEverySchemaIsOrgScoped(t *testing.T) {
	for name, schema := range allSchemas(t) {
		if _, present := schema.Fields()[basemodel.FieldOrgId]; !present {
			t.Errorf("schema %q has no %s: it would not be scoped to an organization",
				name, basemodel.FieldOrgId)
		}
	}
}

// The schema name is both the IAM resource code and the route path segment, so drifting from the
// "accounting_" prefix breaks the IAM seed and the URL at once.
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

// Every resource carries the base model's audit trail, proving who changed the configuration behind
// a charge and when.
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

// The five versioned resources are those the lifecycle rules govern. A missing lifecycle field
// leaves a resource publishable in name only, its material fields never freezing.
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

// Effective dating is what stops a rate change reinterpreting a historical sale, so every versioned
// resource must say when it applies from.
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

// Money and rates must never be float-typed: a float64 cannot hold a decimal fraction exactly, so
// every amount computed from such a rate is subtly wrong.
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
