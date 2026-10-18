package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	itRecord "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsrecord"
)

func TestCopyTenantRowsSqlWithTenant(t *testing.T) {
	query, args, err := buildCopyTenantRowsSql("settings_records", true, itRecord.CopyTenantRowsParam{
		Level: "user", OwnerType: "user", OwnerId: "U1", TenantId: "T1",
	})
	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO settings_records (id, schema_id, module_key, level, owner_type, owner_id, "+
		"name, value, allow_override, created_at, tenant_id) "+
		"SELECT gen_random_uuid()::text, schema_id, module_key, $1, $2, $3, name, value, allow_override, now(), $4 "+
		"FROM settings_records WHERE level = $5 AND tenant_id = $6 ON CONFLICT DO NOTHING", query)
	assert.Equal(t, []any{"user", "user", "U1", "T1", "tenant", "T1"}, args)
}

// A table with a tenant column and a request with no tenant must refuse rather than copy every
// tenant's rows onto the new owner.
func TestCopyTenantRowsSqlRefusesAMissingTenant(t *testing.T) {
	_, _, err := buildCopyTenantRowsSql("settings_records", true, itRecord.CopyTenantRowsParam{
		Level: "org", OwnerType: "org", OwnerId: "O1",
	})
	assert.Error(t, err)
}

func TestCopyTenantRowsSqlWithoutTenantColumn(t *testing.T) {
	query, args, err := buildCopyTenantRowsSql("settings_records", false, itRecord.CopyTenantRowsParam{
		Level: "org", OwnerType: "org", OwnerId: "O1",
	})
	require.NoError(t, err)
	assert.NotContains(t, query, "tenant_id")
	assert.Equal(t, []any{"org", "org", "O1", "tenant"}, args)
}
