package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A settings schema owns no table, so it must never ask for one.
func TestOrgSettingsSchemaIsMetadataOnly(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	assert.Equal(t, OrgSettingsSchemaName, schema.Name())
	assert.Empty(t, schema.TableName())
	assert.Empty(t, schema.PrimaryKeys())
}

// Every unit setting has a default unit, and nothing else does.
func TestEveryUnitSettingHasADefaultSymbol(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	require.Len(t, DefaultUomSymbolOf, len(schema.Fields()))
	for name := range DefaultUomSymbolOf {
		_, declared := schema.Field(name)
		assert.Truef(t, declared, "%q has a default but is not declared", name)
	}
}
