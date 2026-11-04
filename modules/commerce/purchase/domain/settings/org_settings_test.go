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

// Every declared setting is a sequence id today; a new one must be added to the list or to a rule.
func TestEverySequenceSettingIsDeclared(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	require.Len(t, schema.Fields(), len(SequenceIdSettings()))
	for _, name := range SequenceIdSettings() {
		_, declared := schema.Field(name)
		assert.Truef(t, declared, "%q is not declared", name)
	}
}
