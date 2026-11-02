package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	accsettings "github.com/sky-as-code/nikki-erp/modules/accounting/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

type fakeCurrencies map[string]itExt.GetCurrencyResultData

func (this fakeCurrencies) GetCurrency(_ corectx.Context, query itExt.GetCurrencyQuery) (*itExt.GetCurrencyResult, error) {
	data, ok := this[string(query.Id)]
	return &itExt.GetCurrencyResult{Data: data, HasData: ok}, nil
}

func (this fakeCurrencies) GetCurrencyByCode(
	corectx.Context, itExt.GetCurrencyByCodeQuery,
) (*itExt.GetCurrencyByCodeResult, error) {
	return &itExt.GetCurrencyByCodeResult{}, nil
}

// A system currency is checked against the catalogue on write, at both levels; clearing it is not.
func TestSystemCurrencyRule(t *testing.T) {
	rule := newSystemCurrencyRule(fakeCurrencies{
		"vnd": {Id: "vnd", IsActive: true},
		"old": {Id: "old", IsActive: false},
	})
	testCases := []struct {
		name   string
		item   coreSettings.SetSettingItem
		errors int
	}{
		{"usable org currency", coreSettings.SetSettingItem{Name: accsettings.OrgSettingSystemCurrency, Value: "vnd"}, 0},
		{"cleared org currency", coreSettings.SetSettingItem{Name: accsettings.OrgSettingSystemCurrency, Value: ""}, 0},
		{"unknown tenant currency", coreSettings.SetSettingItem{Name: accsettings.TenantSettingSystemCurrency, Value: "xyz"}, 1},
		{"inactive org currency", coreSettings.SetSettingItem{Name: accsettings.OrgSettingSystemCurrency, Value: "old"}, 1},
		{"another setting is not checked", coreSettings.SetSettingItem{Name: accsettings.OrgSettingDefaultPriceMode, Value: "xyz"}, 0},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			vErrs, err := rule(nil, coreSettings.LevelOrg, []coreSettings.SetSettingItem{testCase.item})
			assert.NoError(t, err)
			assert.Equal(t, testCase.errors, vErrs.Count())
		})
	}
}
