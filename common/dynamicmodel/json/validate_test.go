package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSchemaJson_Valid(t *testing.T) {
	modelJson := `{
		"name": "iam_role",
		"label": "iam_role.label",
		"table_name": "iam_roles",
		"should_build_db": true,
		"record_label_field": "name",
		"extend_before": ["core.basemodel.base_model"],
		"fields": [
			{"name": "name", "data_type": {"type": "string", "min": 1, "max": 200}, "required_for_create": true},
			{"name": "is_private", "data_type": "boolean", "required_for_create": true, "default_value": false}
		],
		"extend_after": ["core.basemodel.archivable_model"],
		"partial_uniques": [{"not_null_fields": ["name"], "nullable_field": "org_id"}],
		"edges_to": [
			{
				"edge": "owner_group",
				"type": "many:one",
				"dest_schema": "iam_group",
				"key_map": {"owner_group_id": "id"}
			},
			{
				"edge": "assigned_users",
				"type": "many:many",
				"peer_schema": "iam_user",
				"through_schema": "iam_role_user_assignment",
				"src_field_prefix": "role"
			}
		],
		"edges_from": [
			{"edge": "entitlements", "existing": {"src_schema": "iam_entitlement", "src_edge": "role"}}
		]
	}`

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.NotNil(t, errs)
	assert.Equal(t, 0, errs.Count(), "expected no errors, got: %v", errs)
}

func TestValidateSchemaJson_Malformed(t *testing.T) {
	errs := ValidateSchemaJson(`{"name": `, ModelJsonSchema)

	assert.Equal(t, 1, errs.Count())
	assert.Equal(t, ErrKeyMalformed, errs[0].Key)
}

func TestValidateSchemaJson_MissingName(t *testing.T) {
	errs := ValidateSchemaJson(`{"table_name": "iam_roles"}`, ModelJsonSchema)

	assert.Greater(t, errs.Count(), 0)
}

func TestValidateSchemaJson_ReportsDottedFieldPath(t *testing.T) {
	// fields.1 has no data_type, which is required.
	modelJson := `{
		"name": "iam_role",
		"fields": [
			{"name": "ok", "data_type": "boolean"},
			{"name": "broken"}
		]
	}`

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.Greater(t, errs.Count(), 0)
	assert.True(t, errs.Has("fields.1"), "expected a violation on fields.1, got: %v", errs)
}

func TestValidateSchemaJson_RejectsTupleDataType(t *testing.T) {
	// The tuple form from the original requirement is deliberately unsupported.
	modelJson := `{"name": "iam_role", "fields": [{"name": "n", "data_type": ["string", 0, 200]}]}`

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.Greater(t, errs.Count(), 0)
}

func TestValidateSchemaJson_RejectsUnknownDataType(t *testing.T) {
	modelJson := `{"name": "m", "fields": [{"name": "n", "data_type": "bogus"}]}`

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.Greater(t, errs.Count(), 0)
}

func TestValidateSchemaJson_RejectsJsonMapArray(t *testing.T) {
	// FieldDataTypeJsonMap().ArrayType() panics, so the schema rejects it up front.
	modelJson := `{"name": "m", "fields": [{"name": "n", "data_type": {"type": "jsonmap", "array": true}}]}`

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.Greater(t, errs.Count(), 0)
}

func TestValidateSchemaJson_RequiresDecimalStringBounds(t *testing.T) {
	// Decimal min/max are strings in the Go constructor; numbers must be rejected.
	modelJson := `{"name": "m", "fields": [
		{"name": "n", "data_type": {"type": "decimal", "min": 0, "max": 10, "scale": 2}}
	]}`

	errs := ValidateSchemaJson(modelJson, ModelJsonSchema)

	assert.Greater(t, errs.Count(), 0)
}

func TestValidateSchemaJson_RejectsUnknownProperty(t *testing.T) {
	errs := ValidateSchemaJson(`{"name": "m", "tableName": "x"}`, ModelJsonSchema)

	assert.Greater(t, errs.Count(), 0)
}
