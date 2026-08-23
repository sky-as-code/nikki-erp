package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// A settings schema describes values; it must not own a table. PrimaryKeys is populated only by
// populateDbMetadata, which runs only under ShouldBuildDb, so an empty one is the check.
func TestTenantSettingsSchema_IsMetadataOnly(t *testing.T) {
	schema := TenantSettingsSchemaBuilder().Build()

	assert.Empty(t, schema.PrimaryKeys(),
		"a settings schema must not declare should_build_db: settings owns the only tables")
	assert.Empty(t, schema.TableName())
}

// The schema name is the contract with the settings module and with the frontend, which both key
// off the string rather than off this package.
func TestTenantSettingsSchema_Name(t *testing.T) {
	schema := TenantSettingsSchemaBuilder().Build()

	assert.Equal(t, TenantSettingsSchemaName, schema.Name())
	assert.Equal(t, "iam_tenant_settings", schema.Name())
}

// Both fields must declare allow_override: false, and must declare it explicitly.
//
// Absent metadata reads as overridable, so a dropped entry would not fail anything — it would
// quietly turn a tenant-wide security policy into a per-user suggestion.
func TestTenantSettingsSchema_ForbidsOverride(t *testing.T) {
	schema := TenantSettingsSchemaBuilder().Build()

	for _, name := range []string{TenantSettingSessionTimeoutMinutes, TenantSettingRequireMfa} {
		field, ok := schema.Field(name)
		require.True(t, ok, name)

		allow, found := field.MetadataValue("allow_override")
		assert.True(t, found, "%s must declare allow_override explicitly", name)
		assert.Equal(t, false, allow,
			"%s is a tenant-wide policy: an individual opt-out would make it a suggestion", name)
	}
}

// The floor is 1 rather than 0 because 0 would mean "expire immediately", locking every account in
// the tenant out. The ceiling is one week.
func TestTenantSettingsSchema_SessionTimeoutBounds(t *testing.T) {
	schema := TenantSettingsSchemaBuilder().Build()

	field, ok := schema.Field(TenantSettingSessionTimeoutMinutes)
	require.True(t, ok)

	bounds, ok := field.DataType().Options()[dmodel.FieldDataTypeOptRange].([]int32)
	require.True(t, ok, "session_timeout_minutes must be an int32 with a range")
	require.Len(t, bounds, 2)

	assert.Equal(t, SessionTimeoutMinMinutes, bounds[0],
		"a floor of 0 would mean expire-immediately and lock the tenant out")
	assert.Equal(t, SessionTimeoutMaxMinutes, bounds[1])
	assert.Equal(t, int32(10080), bounds[1], "one week, in minutes")
}

// No default. A tenant that has not answered has not opted out of MFA; defaulting to false would
// silently record that it had.
func TestTenantSettingsSchema_RequireMfaHasNoDefault(t *testing.T) {
	schema := TenantSettingsSchemaBuilder().Build()

	field, ok := schema.Field(TenantSettingRequireMfa)
	require.True(t, ok)

	assert.Nil(t, field.Default(),
		"a tenant that has not answered the MFA question has not answered it 'no'")
}
