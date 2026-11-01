package app

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	accsettings "github.com/sky-as-code/nikki-erp/modules/accounting/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

var systemCurrencySettings = map[string]bool{
	accsettings.TenantSettingSystemCurrency: true,
	accsettings.OrgSettingSystemCurrency:    true,
}

// newSystemCurrencyRule checks a written system currency against Essential's catalogue, so a typo
// is refused at write time rather than surfacing later as an organization with no currency.
// Clearing the value is allowed: an empty org value falls back to the tenant's.
func newSystemCurrencyRule(currencySvc itExt.CurrencyExtService) coreSettings.SettingsRule {
	return func(ctx corectx.Context, _ string, items []coreSettings.SetSettingItem) (ft.ClientErrors, error) {
		vErrs := ft.ClientErrors{}
		for _, item := range items {
			currencyId, _ := item.Value.(string)
			if !systemCurrencySettings[item.Name] || currencyId == "" {
				continue
			}
			found, err := currencySvc.GetCurrency(ctx, itExt.GetCurrencyQuery{Id: model.Id(currencyId)})
			if err != nil {
				return nil, errors.Wrap(err, "system currency rule")
			}
			if cErr := unusableCurrencyError(item.Name, found); cErr != nil {
				vErrs.Append(*cErr)
			}
		}
		return vErrs, nil
	}
}

// unusableCurrencyError reuses Essential's currency keys: they describe the same three situations.
func unusableCurrencyError(field string, found *itExt.GetCurrencyResult) *ft.ClientErrorItem {
	switch {
	case found == nil || !found.HasData:
		return ft.NewValidationError(field, ft.ErrorKey("currency.not_found", "essential"),
			"the currency does not exist")
	case found.Data.IsArchived:
		return ft.NewValidationError(field, ft.ErrorKey("currency.archived", "essential"),
			"the currency is archived")
	case !found.Data.IsActive:
		return ft.NewValidationError(field, ft.ErrorKey("currency.not_active", "essential"),
			"the currency is not active")
	}
	return nil
}
