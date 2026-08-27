package json

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The function kind shares the "function" key with the aggregate kind, so these tests pin both
// sides of that overload: an engine-function name must be accepted, and the six-name aggregate
// enum must still be enforced on the aggregate branch where it moved to.

func computedFieldJson(computed string) string {
	return fmt.Sprintf(`{
		"name": "inventory_product_variant",
		"table_name": "inventory_product_variants",
		"fields": [
			{"name": "sales_tax_mode", "data_type": {"type": "string", "min": 1, "max": 20}},
			{"name": "derived", "data_type": {"type": "ulid", "array": true}, "computed": %s}
		]
	}`, computed)
}

func TestValidateSchemaJson_AcceptsFunctionKind(t *testing.T) {
	modelJson := computedFieldJson(`{
		"kind": "function",
		"is_stored": false,
		"function": "inventory.effective_sales_tax_ids",
		"depends_on": "sales_tax_mode"
	}`)

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.Equal(t, 0, errs.Count(), "expected no errors, got: %v", errs)
}

func TestValidateSchemaJson_FunctionKindRequiresFunctionName(t *testing.T) {
	modelJson := computedFieldJson(`{"kind": "function", "is_stored": false}`)

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.NotEqual(t, 0, errs.Count(), "a function kind without a name must be rejected")
}

// depends_on belongs to the function kind alone; letting it ride along on another kind would
// promise a recompute trigger that nothing implements.
func TestValidateSchemaJson_RejectsDependsOnOutsideFunctionKind(t *testing.T) {
	modelJson := computedFieldJson(`{
		"kind": "related",
		"is_stored": false,
		"field": "template.name",
		"depends_on": "sales_tax_mode"
	}`)

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.NotEqual(t, 0, errs.Count(), "depends_on must be rejected on a related field")
}

// The aggregate enum moved from the shared property into the aggregate branch when the function
// kind started using the same key for a free-form name. This pins that it still bites.
func TestValidateSchemaJson_AggregateStillRejectsUnknownFunction(t *testing.T) {
	modelJson := computedFieldJson(`{
		"kind": "aggregate",
		"is_stored": false,
		"source": "quants",
		"function": "array_agg"
	}`)

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.NotEqual(t, 0, errs.Count(), "array_agg is not one of the six supported aggregates")
}

func TestValidateSchemaJson_FunctionKindRejectsSubqueryProps(t *testing.T) {
	modelJson := computedFieldJson(`{
		"kind": "function",
		"is_stored": false,
		"function": "inventory.effective_sales_tax_ids",
		"source": "quants"
	}`)

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.NotEqual(t, 0, errs.Count(), "a function kind takes no source edge")
}
