package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	pinvsettings "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/settings"
)

// A client may not repoint the invoice series: the org lifecycle owns every sequence id setting.
func TestSequenceSettingsRuleRefusesEverySequenceId(t *testing.T) {
	items := []coreSettings.SetSettingItem{}
	for _, name := range pinvsettings.SequenceIdSettings() {
		items = append(items, coreSettings.SetSettingItem{Name: name, Value: "01OTHER"})
	}

	vErrs, err := SequenceSettingsRule(corectx.NewRequestContext(t.Context()), "", items)

	require.NoError(t, err)
	require.Equal(t, len(items), vErrs.Count())
	assert.Equal(t, pinvsettings.OrgSettingInvoiceNumberSequenceId, vErrs[0].Field)
}
