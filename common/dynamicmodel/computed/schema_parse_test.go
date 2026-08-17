package computed_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func TestParseDefinitionJson_ExpressionMatchesChainedForm(t *testing.T) {
	raw := []byte(`{
		"kind": "expression",
		"is_stored": false,
		"expression": {
			"op": "subtract",
			"args": [
				{"function": "coalesce", "args": [{"field": "on_hand_quantity"}, {"value": 0}]},
				{"function": "coalesce", "args": [{"field": "reserved_quantity"}, {"value": 0}]}
			]
		}
	}`)

	expr, isStored, err := computed.ParseDefinitionJson(raw, "available_quantity")
	require.NoError(t, err)
	assert.False(t, isStored)

	chained := computed.Sub(
		computed.Fn("coalesce", computed.F("on_hand_quantity"), computed.Lit(int64(0))),
		computed.Fn("coalesce", computed.F("reserved_quantity"), computed.Lit(int64(0))),
	)
	assert.Equal(t, chained, expr)
}

func TestParseDefinitionJson_RelatedKind(t *testing.T) {
	raw := []byte(`{"kind": "related", "is_stored": false, "field": "template.name"}`)

	expr, isStored, err := computed.ParseDefinitionJson(raw, "template_name")
	require.NoError(t, err)
	assert.False(t, isStored)
	assert.Equal(t, computed.Related("template.name"), expr)
}

func TestParseDefinitionJson_CaseAndOperators(t *testing.T) {
	raw := []byte(`{
		"kind": "expression",
		"is_stored": false,
		"expression": {
			"case": {
				"when": [
					{"if": {"op": "lte", "args": [{"field": "qty"}, {"value": 0}]}, "then": {"value": "out"}},
					{"if": {"op": "is_null", "args": [{"field": "qty"}]}, "then": {"value": "unknown"}}
				],
				"else": {"value": "in"}
			}
		}
	}`)

	expr, _, err := computed.ParseDefinitionJson(raw, "status")
	require.NoError(t, err)

	chained := computed.Case().
		When(computed.Le(computed.F("qty"), computed.Lit(int64(0))), computed.Lit("out")).
		When(computed.IsNull(computed.F("qty")), computed.Lit("unknown")).
		Else(computed.Lit("in"))
	assert.Equal(t, chained, expr)
}

func TestParseDefinitionJson_LiteralNumbersNeverBecomeFloats(t *testing.T) {
	raw := []byte(`{
		"kind": "expression", "is_stored": false,
		"expression": {"op": "multiply", "args": [{"field": "price"}, {"value": 0.1}]}
	}`)

	expr, _, err := computed.ParseDefinitionJson(raw, "fee")
	require.NoError(t, err)

	lit := expr.(computed.BinaryExpr).Right.(computed.LiteralExpr)
	assert.Equal(t, "decimal.Decimal", fmt.Sprintf("%T", lit.Value),
		"fractional literals must decode as decimal, not float64")
}

func TestParseDefinitionJson_Rejections(t *testing.T) {
	cases := map[string]string{
		"missing is_stored":       `{"kind": "expression", "expression": {"field": "a"}}`,
		"unknown kind":            `{"kind": "aggregate", "is_stored": false, "field": "x"}`,
		"related without field":   `{"kind": "related", "is_stored": false}`,
		"expression without expr": `{"kind": "expression", "is_stored": false}`,
		"unknown operator":        `{"kind": "expression", "is_stored": false, "expression": {"op": "xor", "args": [{"field": "a"}, {"field": "b"}]}}`,
		"subtract with 3 args":    `{"kind": "expression", "is_stored": false, "expression": {"op": "subtract", "args": [{"field": "a"}, {"field": "b"}, {"field": "c"}]}}`,
		"unary with 2 args":       `{"kind": "expression", "is_stored": false, "expression": {"op": "not", "args": [{"field": "a"}, {"field": "b"}]}}`,
		"empty node":              `{"kind": "expression", "is_stored": false, "expression": {}}`,
		"case without else":       `{"kind": "expression", "is_stored": false, "expression": {"case": {"when": [{"if": {"field": "a"}, "then": {"value": 1}}]}}}`,
		"object literal":          `{"kind": "expression", "is_stored": false, "expression": {"value": {"nested": true}}}`,
		"malformed json":          `{"kind": `,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			_, _, err := computed.ParseDefinitionJson([]byte(raw), "f")
			require.Error(t, err)
		})
	}
}

func TestParseModelJson_ComputedFieldEndToEnd(t *testing.T) {
	modelJson := `{
		"name": "cf_json_quant",
		"fields": [
			{"name": "id", "data_type": "ulid", "primary_key": true},
			{"name": "on_hand_quantity", "data_type": {"type": "decimal", "min": "0", "max": "999999", "scale": 4}},
			{"name": "reserved_quantity", "data_type": {"type": "decimal", "min": "0", "max": "999999", "scale": 4}},
			{"name": "available_quantity",
			 "data_type": {"type": "decimal", "min": "-999999", "max": "999999", "scale": 4},
			 "computed": {
				"kind": "expression",
				"is_stored": false,
				"expression": {
					"op": "subtract",
					"args": [
						{"function": "coalesce", "args": [{"field": "on_hand_quantity"}, {"value": 0}]},
						{"function": "coalesce", "args": [{"field": "reserved_quantity"}, {"value": 0}]}
					]
				}
			 }}
		]
	}`

	schema := dmodel.ParseModelJson(modelJson).Build()
	field, ok := schema.Field("available_quantity")
	require.True(t, ok)
	assert.True(t, field.IsComputed())
	assert.True(t, field.IsVirtual())

	def, err := computed.DefOf(field)
	require.NoError(t, err)
	assert.Equal(t, computed.ComputeExpression, def.Kind)
}

func TestParseModelJson_ComputedSchemaValidation(t *testing.T) {
	template := `{
		"name": "cf_json_invalid",
		"fields": [
			{"name": "id", "data_type": "ulid", "primary_key": true},
			{"name": "broken", "data_type": {"type": "string", "min": 0, "max": 100}, "computed": %s}
		]
	}`
	invalidBlocks := map[string]string{
		"kind outside enum":        `{"kind": "aggregate", "is_stored": false, "field": "x"}`,
		"missing is_stored":        `{"kind": "related", "field": "template.name"}`,
		"related with expression":  `{"kind": "related", "is_stored": false, "field": "a", "expression": {"field": "b"}}`,
		"expression with field":    `{"kind": "expression", "is_stored": false, "field": "a", "expression": {"field": "b"}}`,
		"unknown property":         `{"kind": "related", "is_stored": false, "field": "a", "sql": "DROP TABLE x"}`,
		"node with field and op":   `{"kind": "expression", "is_stored": false, "expression": {"field": "a", "op": "add", "args": [{"field": "b"}, {"field": "c"}]}}`,
		"op outside enum":          `{"kind": "expression", "is_stored": false, "expression": {"op": "exec", "args": [{"field": "a"}]}}`,
		"case branch missing then": `{"kind": "expression", "is_stored": false, "expression": {"case": {"when": [{"if": {"field": "a"}}], "else": {"value": 1}}}}`,
	}
	for name, block := range invalidBlocks {
		t.Run(name, func(t *testing.T) {
			_, errs := dmodel.ParseModelJsonSafe(fmt.Sprintf(template, block))
			assert.Greater(t, errs.Count(), 0, "the JSON Schema must reject this block")
		})
	}

	t.Run("valid related block passes the same template", func(t *testing.T) {
		// Control: proves the rejections above come from the computed block, not the scaffold.
		_, errs := dmodel.ParseModelJsonSafe(fmt.Sprintf(template,
			`{"kind": "related", "is_stored": false, "field": "template.name"}`))
		assert.Equal(t, 0, errs.Count(), errs)
	})
}
