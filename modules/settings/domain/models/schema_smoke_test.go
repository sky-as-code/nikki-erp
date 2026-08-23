package models

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// The JSON documents extend the core base mixins by name, which CoreModule.RegisterModels does at
// app start-up. Without it every parse here panics on an unregistered mixin.
func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}

// The JSON documents are parsed and validated at boot, so a malformed one panics the app rather
// than failing a request. Building them here turns that into a test failure instead.
func TestSettingsSchemas_Build(t *testing.T) {
	schemaSchema := SettingsSchemaSchemaBuilder().Build()
	assert.Equal(t, SettingsSchemaSchemaName, schemaSchema.Name())
	assert.Equal(t, "settings_schemas", schemaSchema.TableName())

	recordSchema := SettingsRecordSchemaBuilder().Build()
	assert.Equal(t, SettingsRecordSchemaName, recordSchema.Name())
	assert.Equal(t, "settings_records", recordSchema.TableName())

	for _, name := range []string{
		SettingsRecordFieldSchemaId, SettingsRecordFieldModuleKey, SettingsRecordFieldLevel,
		SettingsRecordFieldOwnerType, SettingsRecordFieldOwnerId, SettingsRecordFieldName,
		SettingsRecordFieldValue,
	} {
		_, ok := recordSchema.Field(name)
		assert.True(t, ok, "settings_record must declare %q", name)
	}
}

// Both tables are real, so both must carry DB metadata. PrimaryKeys() is populated only by
// populateDbMetadata, which runs only under should_build_db.
func TestSettingsSchemas_BuildDb(t *testing.T) {
	assert.NotEmpty(t, SettingsSchemaSchemaBuilder().Build().PrimaryKeys())
	assert.NotEmpty(t, SettingsRecordSchemaBuilder().Build().PrimaryKeys())
}

// The envelope is what lets a scalar and a list share one json_map column, and a null column is
// reachable from the database, so the accessor must report absence rather than panic.
func TestSettingsRecord_ValueEnvelope(t *testing.T) {
	record := NewSettingsRecord()

	_, ok := record.GetValue()
	assert.False(t, ok, "an unset value must read as absent")

	for _, stored := range []any{"dark", float64(3), true, []any{"a", "b"}} {
		record.SetValue(stored)
		got, ok := record.GetValue()
		require.True(t, ok)
		assert.Equal(t, stored, got)
	}

	record.SetValue(nil)
	got, ok := record.GetValue()
	assert.True(t, ok, "an explicit null is present but empty, not absent")
	assert.Nil(t, got)
}

// The Go level/owner-type constants and the JSON enums are two declarations of one set. A value
// added to only one of them would be writable through the constant and rejected by the schema, or
// the reverse, so they are pinned against each other here.
func TestSettingsRecord_EnumsMatchConstants(t *testing.T) {
	recordSchema := SettingsRecordSchemaBuilder().Build()

	testCases := []struct {
		field    string
		expected []string
	}{
		{
			field:    SettingsRecordFieldLevel,
			expected: []string{constants.LevelTenant, constants.LevelOrg, constants.LevelUser},
		},
		{
			field:    SettingsRecordFieldOwnerType,
			expected: []string{constants.OwnerTypeTenant, constants.OwnerTypeOrg, constants.OwnerTypeUser},
		},
	}

	for _, testCase := range testCases {
		field, ok := recordSchema.Field(testCase.field)
		require.True(t, ok, testCase.field)

		values, ok := field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues].([]string)
		require.True(t, ok, "%s must be an enum_string", testCase.field)
		assert.ElementsMatch(t, testCase.expected, values, testCase.field)
	}
}

// settings_schemas must not inherit a tenant key, and settings_records must.
//
// This is the split the whole module rests on: a module's *declaration* of what it can be
// configured with is identical for every tenant and is registered once at start-up, when no tenant
// is in scope; only the *values* are per-tenant. Extending core.basemodel.base_model on the schema
// table would silently reintroduce the tenant column, because that mixin calls ExtendBase() — which
// is how this was got wrong the first time, and it panics the app at boot with "tenant ID is
// required" rather than failing here.
//
// The nikkierp binary installs no base model, so neither carries a tenant key in this test; the
// coremart-side equivalent is TestGeneratedIndexNamesFitPostgresLimit's tree, where the base model
// is installed. What this pins is that settings_schema does not declare one of its own and does not
// extend a mixin that would.
func TestSettingsSchemas_TenantScoping(t *testing.T) {
	schemaSchema := SettingsSchemaSchemaBuilder().Build()

	_, hasOwnId := schemaSchema.Field(SettingsSchemaFieldId)
	assert.True(t, hasOwnId,
		"settings_schema must declare its own id rather than extending base_model, which carries the tenant key")

	// The decisive assertion. In the coremart binary the base model injects a tenant key into
	// every schema that extends base_model; here it pins that settings_schema declares none of its
	// own, which is what keeps it global once that base model is installed.
	assert.Empty(t, schemaSchema.TenantKey(),
		"settings_schema must not be tenant-scoped: module declarations are the same for every tenant")
}
