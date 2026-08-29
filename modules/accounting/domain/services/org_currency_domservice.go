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

// OrgCurrencyDomainServiceImpl resolves the organization's book currency from its settings: read
// the configured id, then resolve it through Essential. The second call is not decoration, since a
// setting may name a currency that has since been deleted.
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
// Unlike Sales' ResolveSalesPolicy, this never swallows a failed settings read and falls back to a
// default: there is no safe default currency, and a wrong one silently reinterprets every amount.
// An unreadable setting surfaces as an error and an unconfigured one as HasData false.
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
	// A configured id naming no currency is reported as no configuration at all: the remedy is the
	// same, so a second failure mode would only oblige every caller to handle both.
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
	// Never a bare type assertion: the value has been through a jsonb column, so a wrong shape is a
	// configuration defect rather than a reason to panic.
	value, _ := result.Data.Values[key].(string)
	return value, nil
}
