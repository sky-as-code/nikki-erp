package model

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// virtualTestSchemaJson declares one physical field and one virtual field, which is the shape
// every test below needs: something that must be written and something that must not be.
const virtualTestSchemaJson = `{
	"name": "test_virtual",
	"table_name": "test_virtuals",
	"should_build_db": true,
	"fields": [
		{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true},
		{"name": "sku", "data_type": {"type": "string", "min": 1, "max": 100}},
		{"name": "template_name", "data_type": {"type": "string", "min": 1, "max": 200}, "is_virtual": true}
	]
}`

func buildVirtualTestSchema(t *testing.T) *ModelSchema {
	t.Helper()
	return ParseModelJson(virtualTestSchemaJson).Build()
}

func TestVirtualField_ParsesFromJson(t *testing.T) {
	schema := buildVirtualTestSchema(t)

	field, ok := schema.Field("template_name")
	require.True(t, ok, "a virtual field is still a field")
	assert.True(t, field.IsVirtual())
	assert.False(t, field.IsVirtualModelField(), "a scalar is not a model-typed edge field")
	assert.True(t, field.IsNonPhysical())

	sku, ok := schema.Field("sku")
	require.True(t, ok)
	assert.False(t, sku.IsVirtual())
	assert.False(t, sku.IsNonPhysical())
}

// Columns drives DDL and writes; ReadableFields drives projections. A virtual field must appear
// in exactly one of them, which is the whole point of the split.
func TestVirtualField_ExcludedFromColumnsIncludedInReadable(t *testing.T) {
	schema := buildVirtualTestSchema(t)

	assert.Equal(t, []string{"id", "sku"}, fieldNamesOf(schema.Columns()))
	assert.Equal(t, []string{"id", "sku", "template_name"}, fieldNamesOf(schema.ReadableFields()))
	assert.Contains(t, schema.FieldNames(), "template_name")
}

func fieldNamesOf(fields []*ModelField) []string {
	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, field.Name())
	}
	return names
}

// A virtual value must never reach the write. Dropping it in Validate is what makes that true:
// the result map is what the repository writes from.
func TestVirtualField_DroppedByValidateOnCreate(t *testing.T) {
	schema := buildVirtualTestSchema(t)

	result, errs := schema.Validate(DynamicFields{
		"sku":           "SKU-1",
		"template_name": "Classic T-Shirt",
	})

	assert.Zero(t, errs.Count())
	assert.Equal(t, "SKU-1", result["sku"])
	assert.NotContains(t, result, "template_name", "a virtual field is never written")
}

func TestVirtualField_DroppedByValidateOnEdit(t *testing.T) {
	schema := buildVirtualTestSchema(t)

	result, errs := schema.Validate(DynamicFields{
		"sku":           "SKU-2",
		"template_name": "Anything",
	}, true)

	assert.Zero(t, errs.Count())
	assert.NotContains(t, result, "template_name")
}

// The skip must happen before validation, not after: a client round-tripping a GET into a PUT
// sends back whatever it received, and that must never become a validation failure for a field
// it does not control.
func TestVirtualField_InvalidValueStillDoesNotError(t *testing.T) {
	schema := buildVirtualTestSchema(t)

	result, errs := schema.Validate(DynamicFields{
		"sku":           "SKU-3",
		"template_name": 12345, // an int where a string is declared
	})

	assert.Zero(t, errs.Count(), "a virtual field is skipped before its data type is checked")
	assert.NotContains(t, result, "template_name")
}

func TestVirtualField_NotServiceInjected(t *testing.T) {
	schema := buildVirtualTestSchema(t)

	result := DynamicFields{}
	schema.InjectServiceFields(context.Background(), result, false)

	assert.NotContains(t, result, "template_name")
}

// meta/schema is a published contract. is_virtual_model_field must keep meaning exactly what it
// meant before this flag existed, or every client reading it silently changes behaviour.
func TestVirtualField_ToSimplizedKeepsEdgeFlagDistinct(t *testing.T) {
	schema := buildVirtualTestSchema(t)
	field, ok := schema.Field("template_name")
	require.True(t, ok)

	dto := toSimplizedMap(t, field)

	assert.Equal(t, true, dto["is_virtual"])
	assert.Equal(t, false, dto["is_virtual_model_field"])
	assert.Equal(t, true, dto["is_system_field"], "a virtual field is read-only to a client")
}

func TestVirtualField_ToSimplizedOnEdgeFieldUnchanged(t *testing.T) {
	schema := ParseModelJson(`{
		"name": "test_virtual_edge_src",
		"table_name": "test_virtual_edge_srcs",
		"fields": [
			{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true},
			{"name": "peer", "data_type": "model"}
		]
	}`).Build()

	field, ok := schema.Field("peer")
	require.True(t, ok)

	dto := toSimplizedMap(t, field)

	assert.Equal(t, true, dto["is_virtual_model_field"], "unchanged meaning")
	assert.Equal(t, false, dto["is_virtual"], "a model-typed field is not a virtual scalar")
}

func toSimplizedMap(t *testing.T, field *ModelField) map[string]any {
	t.Helper()
	raw, err := json.Marshal(field.ToSimplized())
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

// Each combination below is a promise nothing can keep: the field is dropped from every write,
// so requiring it, indexing it or generating it would fail silently and far from its cause.
func TestVirtualField_ContradictionsPanicAtBuild(t *testing.T) {
	testCases := []struct {
		name  string
		attrs string
	}{
		{"primary key", `"primary_key": true`},
		{"tenant key", `"tenant_key": true`},
		{"versioning key", `"versioning_key": true`},
		{"unique", `"unique": true`},
		{"required for create", `"required_for_create": true`},
		{"required for update", `"required_for_update": true`},
		{"required always", `"required_always": true`},
		{"auto generated", `"auto_generated": true`},
		{"no update", `"no_update": true`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Panics(t, func() {
				ParseModelJson(`{
					"name": "test_virtual_bad",
					"table_name": "test_virtual_bads",
					"fields": [
						{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true},
						{"name": "bad", "data_type": {"type": "string", "min": 0, "max": 10},
							"is_virtual": true, ` + testCase.attrs + `}
					]
				}`).Build()
			})
		})
	}
}

func TestVirtualField_RequiredWithPanicsAtBuild(t *testing.T) {
	assert.Panics(t, func() {
		ParseModelJson(`{
			"name": "test_virtual_reqwith",
			"table_name": "test_virtual_reqwiths",
			"fields": [
				{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true},
				{"name": "other", "data_type": {"type": "string", "min": 0, "max": 10}},
				{"name": "bad", "data_type": {"type": "string", "min": 0, "max": 10},
					"is_virtual": true, "required_with": "other"}
			]
		}`).Build()
	})
}

func TestVirtualField_AsRecordLabelPanicsAtBuild(t *testing.T) {
	assert.Panics(t, func() {
		ParseModelJson(`{
			"name": "test_virtual_label",
			"table_name": "test_virtual_labels",
			"record_label_field": "bad",
			"fields": [
				{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true},
				{"name": "bad", "data_type": {"type": "string", "min": 0, "max": 10}, "is_virtual": true}
			]
		}`).Build()
	})
}

func TestVirtualField_InIndexOrConstraintPanicsAtBuild(t *testing.T) {
	testCases := []struct {
		name   string
		clause string
	}{
		{
			"composite unique",
			`"composite_uniques": [{"index_name": "tvi_bad_key", "fields": ["sku", "bad"]}],`,
		},
		{
			"search index",
			`"search_indexes": [{"index_name": "tvi_bad_idx", "fields": ["bad"]}],`,
		},
		{
			"exclusive required",
			`"exclusive_required_fields": [["sku", "bad"]],`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Panics(t, func() {
				ParseModelJson(`{
					"name": "test_virtual_idx",
					"table_name": "test_virtual_idxs",
					` + testCase.clause + `
					"fields": [
						{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true},
						{"name": "sku", "data_type": {"type": "string", "min": 1, "max": 100},
							"required_for_create": true},
						{"name": "bad", "data_type": {"type": "string", "min": 0, "max": 10}, "is_virtual": true}
					]
				}`).Build()
			})
		})
	}
}

// additionalProperties is false on the field schema, so a typo'd flag must be rejected rather
// than silently ignored -- which is also what forces is_virtual to be declared in the meta-schema.
func TestVirtualField_TypoedFlagIsRejected(t *testing.T) {
	assert.Panics(t, func() {
		ParseModelJson(`{
			"name": "test_virtual_typo",
			"table_name": "test_virtual_typos",
			"fields": [
				{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true},
				{"name": "bad", "data_type": {"type": "string", "min": 0, "max": 10}, "is_virtua1": true}
			]
		}`).Build()
	})
}

func TestVirtualField_ClonePreservesFlag(t *testing.T) {
	schema := buildVirtualTestSchema(t)
	field, ok := schema.Field("template_name")
	require.True(t, ok)

	assert.True(t, field.Clone().IsVirtual())
}
