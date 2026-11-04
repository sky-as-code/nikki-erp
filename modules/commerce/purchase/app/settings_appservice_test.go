package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pursettings "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/settings"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

// A client may not repoint a code series: the org lifecycle owns every sequence id setting.
func TestSequenceSettingsRuleRefusesEverySequenceId(t *testing.T) {
	items := []coreSettings.SetSettingItem{}
	for _, name := range pursettings.SequenceIdSettings() {
		items = append(items, coreSettings.SetSettingItem{Name: name, Value: "01OTHER"})
	}

	vErrs, err := SequenceSettingsRule(corectx.NewRequestContext(t.Context()), "", items)

	require.NoError(t, err)
	require.Equal(t, len(items), vErrs.Count())
	assert.Equal(t, pursettings.OrgSettingOrderCodeSequenceId, vErrs[0].Field)
}
