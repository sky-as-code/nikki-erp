package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const conditionalJson = `{
	"name": "test_cond_json",
	"fields": [
		{"name": "kind", "data_type": {"type": "enum_string", "values": ["internal", "vendor", "view"]},
			"required_fields": {"internal": ["capacity", "parent_id"]},
			"optional_fields": {"internal": ["note"], "vendor": ["note"]}},
		{"name": "capacity", "data_type": {"type": "int32", "min": 0, "max": 1000}},
		{"name": "parent_id", "data_type": {"type": "string", "min": 1, "max": 50}},
		{"name": "note", "data_type": {"type": "string", "min": 0, "max": 50}}
	]
}`

func conditionalBuilder(name string) *ModelSchemaBuilder {
	return DefineModel(name).
		Field(DefineField().Name("kind").
			DataType(FieldDataTypeEnumString([]string{"internal", "vendor", "view"})).
			RequiredFields(map[string][]string{"internal": {"capacity", "parent_id"}}).
			OptionalFields(map[string][]string{"internal": {"note"}, "vendor": {"note"}})).
		Field(DefineField().Name("capacity").DataType(FieldDataTypeInt32(0, 1000))).
		Field(DefineField().Name("parent_id").DataType(FieldDataTypeString(1, 50))).
		Field(DefineField().Name("note").DataType(FieldDataTypeString(0, 50)))
}

func TestConditionalFields_JsonAndBuilderAgree(t *testing.T) {
	fromJson := ParseModelJson(conditionalJson).Build()
	fromBuilder := conditionalBuilder("test_cond_builder").Build()

	for name, schema := range map[string]*ModelSchema{"json": fromJson, "builder": fromBuilder} {
		field := schema.MustField("kind")
		assert.Equal(t, map[string][]string{"internal": {"capacity", "parent_id"}}, field.RequiredFields(), name)
		assert.Equal(t, map[string][]string{"internal": {"note"}, "vendor": {"note"}}, field.OptionalFields(), name)
		assert.True(t, schema.HasConditionalRequiredFields(), name)
	}
}

func TestConditionalFields_SimplizedOmitsWhenEmpty(t *testing.T) {
	schema := ParseModelJson(conditionalJson).Build()

	raw, err := json.Marshal(schema.MustField("kind").ToSimplized())
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"required_fields":{"internal":["capacity","parent_id"]}`)
	assert.Contains(t, string(raw), `"optional_fields":{"internal":["note"],"vendor":["note"]}`)

	raw, err = json.Marshal(schema.MustField("note").ToSimplized())
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "required_fields")
	assert.NotContains(t, string(raw), "optional_fields")
}

func TestConditionalFields_ModelJsonRoundTrip(t *testing.T) {
	original := ParseModelJson(conditionalJson).Build()
	document, err := original.ToModelJson()
	require.NoError(t, err)
	document["name"] = "test_cond_round_trip"
	raw, err := json.Marshal(document)
	require.NoError(t, err)

	rebuilt := ParseModelJson(string(raw)).Build()
	assert.Equal(t, original.MustField("kind").RequiredFields(), rebuilt.MustField("kind").RequiredFields())
	assert.Equal(t, original.MustField("kind").OptionalFields(), rebuilt.MustField("kind").OptionalFields())
}

func TestConditionalFields_CloneIsDeep(t *testing.T) {
	field := ParseModelJson(conditionalJson).Build().MustField("kind")
	cloned := field.Clone()
	cloned.requiredFields["internal"][0] = "changed"
	assert.Equal(t, "capacity", field.RequiredFields()["internal"][0])
}

func TestConditionalFields_BuildPanics(t *testing.T) {
	enumField := func() *FieldBuilder {
		return DefineField().Name("kind").DataType(FieldDataTypeEnumString([]string{"a", "b"}))
	}
	other := DefineField().Name("other").DataType(FieldDataTypeString(0, 50))
	cases := map[string]*FieldBuilder{
		"non-enum source": DefineField().Name("kind").DataType(FieldDataTypeString(0, 50)).
			RequiredFields(map[string][]string{"a": {"other"}}),
		"enum_int32 source": DefineField().Name("kind").DataType(FieldDataTypeEnumInt32([]int32{1})).
			OptionalFields(map[string][]string{"1": {"other"}}),
		"array source": DefineField().Name("kind").
			DataType(FieldDataTypeEnumString([]string{"a"}).ArrayType()).
			RequiredFields(map[string][]string{"a": {"other"}}),
		"unknown key":   enumField().RequiredFields(map[string][]string{"zzz": {"other"}}),
		"unknown field": enumField().OptionalFields(map[string][]string{"a": {"missing"}}),
		"self":          enumField().RequiredFields(map[string][]string{"a": {"kind"}}),
		"both maps": enumField().RequiredFields(map[string][]string{"a": {"other"}}).
			OptionalFields(map[string][]string{"a": {"other"}}),
	}
	for name, source := range cases {
		assert.Panics(t, func() {
			DefineModel("test_cond_panic").Field(source).Field(other).Build()
		}, name)
	}
}

func TestConditionalFields_SameFieldInDifferentKeysIsAllowed(t *testing.T) {
	assert.NotPanics(t, func() {
		DefineModel("test_cond_ok").
			Field(DefineField().Name("kind").DataType(FieldDataTypeEnumString([]string{"a", "b"})).
				RequiredFields(map[string][]string{"a": {"other"}}).
				OptionalFields(map[string][]string{"b": {"other"}})).
			Field(DefineField().Name("other").DataType(FieldDataTypeString(0, 50))).
			Build()
	})
}

func TestConditionalFields_ValidateOnCreate(t *testing.T) {
	schema := conditionalBuilder("test_cond_create").Build()

	_, errs := schema.Validate(DynamicFields{"kind": "internal"})
	assert.True(t, errs.Has("capacity"), "%v", errs)
	assert.True(t, errs.Has("parent_id"), "%v", errs)
	assert.False(t, errs.Has("note"), "optional fields are never enforced")

	_, errs = schema.Validate(DynamicFields{"kind": "internal", "capacity": 0, "parent_id": "P1"})
	assert.Equal(t, 0, errs.Count(), "zero is a value: %v", errs)

	_, errs = schema.Validate(DynamicFields{"kind": "vendor"})
	assert.Equal(t, 0, errs.Count(), "%v", errs)

	_, errs = schema.Validate(DynamicFields{})
	assert.Equal(t, 0, errs.Count(), "rules skip an empty source: %v", errs)
}

func TestConditionalFields_InvalidSourceSkipsRules(t *testing.T) {
	schema := conditionalBuilder("test_cond_bad_source").Build()

	_, errs := schema.Validate(DynamicFields{"kind": "nonsense"})
	assert.True(t, errs.Has("kind"))
	assert.False(t, errs.Has("capacity"), "%v", errs)
}

func TestConditionalFields_EditDefersToStoredRow(t *testing.T) {
	schema := conditionalBuilder("test_cond_edit").Build()

	_, errs := schema.Validate(DynamicFields{"kind": "internal"}, true)
	assert.Equal(t, 0, errs.Count(), "Validate alone cannot judge a partial edit: %v", errs)

	stored := DynamicFields{"kind": "vendor", "capacity": 5, "parent_id": nil}
	errs = schema.ValidateConditionalFields(DynamicFields{"kind": "internal"}, stored)
	assert.True(t, errs.Has("parent_id"), "%v", errs)
	assert.False(t, errs.Has("capacity"), "stored value counts: %v", errs)

	errs = schema.ValidateConditionalFields(DynamicFields{"kind": "internal", "parent_id": "P1"}, stored)
	assert.Equal(t, 0, errs.Count(), "%v", errs)

	stored = DynamicFields{"kind": "internal", "capacity": 5, "parent_id": "P1"}
	errs = schema.ValidateConditionalFields(DynamicFields{"capacity": nil}, stored)
	assert.True(t, errs.Has("capacity"), "clearing a required field on a stored source: %v", errs)

	errs = schema.ValidateConditionalFields(DynamicFields{"note": "x"}, stored)
	assert.Equal(t, 0, errs.Count(), "%v", errs)
}
