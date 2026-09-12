package settings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}

// A settings schema owns no table. Declaring one, or extending a base mixin, would inject columns
// with nowhere to live and panic the app at boot.
func TestOrgSettingsSchemaBuildsAndOwnsNoTable(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	assert.Equal(t, OrgSettingsSchemaName, schema.Name())
	assert.Empty(t, schema.TableName())

	for _, name := range []string{
		OrgSettingWebEnabled,
		OrgSettingStreamHeartbeatIntervalSeconds,
		OrgSettingStreamMaxBufferedEvents,
	} {
		_, found := schema.Field(name)
		assert.True(t, found, "org settings must declare %q", name)
	}
}
