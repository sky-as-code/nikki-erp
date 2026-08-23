package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Both authoring forms must produce the same metadata, since a module may declare a schema
// either way.
func TestFieldMetadata_JsonAndBuilderAgree(t *testing.T) {
	fromJson := ParseModelJson(`{
		"name": "test_meta_json",
		"table_name": "test_meta_jsons",
		"fields": [
			{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true},
			{
				"name": "theme_mode",
				"data_type": {"type": "string", "min": 1, "max": 20},
				"metadata": {"allow_override": true}
			}
		]
	}`).Build()

	fromBuilder := DefineModel("test_meta_builder").
		Field(DefineField().
			Name("theme_mode").
			DataType(FieldDataTypeString(1, 20)).
			Metadata(map[string]any{"allow_override": true})).
		Build()

	for name, schema := range map[string]*ModelSchema{"json": fromJson, "builder": fromBuilder} {
		field, ok := schema.Field("theme_mode")
		require.True(t, ok, name)

		val, found := field.MetadataValue("allow_override")
		assert.True(t, found, name)
		assert.Equal(t, true, val, name)
	}
}

// A field declaring no metadata must behave exactly as before: nil, not an empty map, and no
// "metadata" key in the simplized payload the frontend receives.
func TestFieldMetadata_AbsentIsAdditive(t *testing.T) {
	schema := ParseModelJson(`{
		"name": "test_meta_absent",
		"table_name": "test_meta_absents",
		"fields": [
			{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true}
		]
	}`).Build()

	field, ok := schema.Field("id")
	require.True(t, ok)
	assert.Nil(t, field.Metadata())

	_, found := field.MetadataValue("allow_override")
	assert.False(t, found)

	encoded, err := json.Marshal(field.ToSimplized())
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "metadata", "omitempty must keep the payload unchanged")
}

// ToSimplized feeds the frontend's meta/schema endpoint, so metadata has to survive it.
func TestFieldMetadata_ReachesSimplizedPayload(t *testing.T) {
	schema := ParseModelJson(`{
		"name": "test_meta_simplized",
		"table_name": "test_meta_simplizeds",
		"fields": [
			{"name": "id", "data_type": "ulid", "primary_key": true, "use_type_default": true},
			{
				"name": "language",
				"data_type": {"type": "string", "min": 1, "max": 20},
				"metadata": {"allow_override": false, "group": "profile"}
			}
		]
	}`).Build()

	field, ok := schema.Field("language")
	require.True(t, ok)

	encoded, err := json.Marshal(field.ToSimplized())
	require.NoError(t, err)

	var decoded struct {
		Metadata map[string]any `json:"metadata"`
	}
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.Equal(t, false, decoded.Metadata["allow_override"])
	assert.Equal(t, "profile", decoded.Metadata["group"])
}

// Clone backs Extend. A derived schema adding a key must not write through into the mixin's
// shared field definition.
func TestFieldMetadata_CloneDoesNotAliasTheSource(t *testing.T) {
	original := DefineField().
		Name("theme_mode").
		DataType(FieldDataTypeString(1, 20)).
		Metadata(map[string]any{"allow_override": true}).
		Build()

	cloned := original.Clone()
	cloned.metadata["allow_override"] = false
	cloned.metadata["added_later"] = true

	assert.Equal(t, true, original.Metadata()["allow_override"], "clone must not mutate the source")
	_, found := original.MetadataValue("added_later")
	assert.False(t, found)
}

// Repeated calls merge rather than replace, so a schema extending a mixin can add one key
// without discarding what the mixin declared.
func TestFieldMetadata_RepeatedCallsMerge(t *testing.T) {
	field := DefineField().
		Name("theme_mode").
		DataType(FieldDataTypeString(1, 20)).
		Metadata(map[string]any{"allow_override": true, "group": "profile"}).
		Metadata(map[string]any{"group": "appearance"}).
		Build()

	assert.Equal(t, true, field.Metadata()["allow_override"], "earlier keys survive")
	assert.Equal(t, "appearance", field.Metadata()["group"], "later keys win")
}
