package services

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	accsettings "github.com/sky-as-code/nikki-erp/modules/accounting/domain/settings"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/currency"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// OrgCurrencyDomainServiceImpl resolves the organization's book currency from its settings.
//
// Two calls, in order: read the setting to learn which currency was configured, then resolve that
// id through Essential to turn it into something a caller can use. The second call is not
// decoration — a setting holds whatever was written into it, and a currency that has since been
// deleted would otherwise be handed out as a valid reference.
type OrgCurrencyDomainServiceImpl struct {
	settingsSvc itExt.EffectiveSettingsExtService
	currencySvc itExt.CurrencyExtService
}

func NewOrgCurrencyDomainServiceImpl(
	settingsSvc itExt.EffectiveSettingsExtService,
	currencySvc itExt.CurrencyExtService,
) itCurrency.OrgCurrencyService {
	return &OrgCurrencyDomainServiceImpl{
		settingsSvc: settingsSvc,
		currencySvc: currencySvc,
	}
}

var _ itCurrency.OrgCurrencyService = (*OrgCurrencyDomainServiceImpl)(nil)

// GetOrgDefaultCurrency implements itCurrency.OrgCurrencyService.
//
// Unlike Sales' ResolveSalesPolicy, this does NOT swallow a failed settings read and fall back to a
// default. The reasoning there was that refusing to sell is a worse response to a misconfiguration
// than selling under a documented default, and every default was a safe reading. Neither holds
// here: there is no safe default currency, and a wrong one silently reinterprets every amount in
// the system — the exact failure BR-PRICE-CUR-004 names. So an unreadable setting surfaces as an
// error and an unconfigured one as HasData false, and the caller decides which it can live with.
func (this *OrgCurrencyDomainServiceImpl) GetOrgDefaultCurrency(
	ctx corectx.Context, query itCurrency.GetOrgDefaultCurrencyQuery,
) (*itCurrency.GetOrgDefaultCurrencyResult, error) {
	currencyId, err := this.readConfiguredCurrencyId(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetOrgDefaultCurrency")
	}
	if currencyId == "" {
		return &itCurrency.GetOrgDefaultCurrencyResult{HasData: false}, nil
	}

	currency, err := this.currencySvc.GetCurrency(ctx, itExt.GetCurrencyQuery{
		Id: model.Id(currencyId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetOrgDefaultCurrency")
	}
	// A configured id naming no currency is reported the same way as no configuration at all.
	// The caller's remedy is identical — configure a usable currency — and inventing a second
	// failure mode would only oblige every caller to handle both.
	if currency == nil || !currency.HasData {
		return &itCurrency.GetOrgDefaultCurrencyResult{HasData: false}, nil
	}

	return &itCurrency.GetOrgDefaultCurrencyResult{
		HasData: true,
		Data: itCurrency.GetOrgDefaultCurrencyResultData{
			CurrencyId:    string(currency.Data.Id),
			Code:          currency.Data.Code,
			Symbol:        currency.Data.Symbol,
			DecimalPlaces: currency.Data.DecimalPlaces,
		},
	}, nil
}

// readConfiguredCurrencyId returns the raw setting value, or empty when it is unset.
func (this *OrgCurrencyDomainServiceImpl) readConfiguredCurrencyId(
	ctx corectx.Context,
) (string, error) {
	result, err := this.settingsSvc.GetEffectiveSettings(ctx, itExt.GetEffectiveSettingsQuery{
		ModuleKeys: []string{modconstants.AccountingModuleName},
	})
	if err != nil {
		return "", err
	}
	if result == nil || !result.HasData || result.ClientErrors.Count() > 0 {
		return "", nil
	}

	key := modconstants.AccountingModuleName + "." + accsettings.OrgSettingOrgDefaultCurrency
	// Never a bare type assertion: the value has been through a jsonb column and back, so a
	// wrong-shaped value is a configuration defect rather than a reason to panic.
	value, _ := result.Data.Values[key].(string)
	return value, nil
}
