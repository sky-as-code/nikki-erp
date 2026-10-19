package app

import (
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/common/model"
	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	orgSettings "github.com/sky-as-code/nikki-erp/modules/accounting/domain/settings"
	orgDefaults "github.com/sky-as-code/nikki-erp/modules/accounting/infra/defaults"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	itTax "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle/defaults"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The currency and rounding policy every new organization starts in. They match the defaults'
// tax tree, which is Vietnamese.
const (
	defaultCurrencyCode       = "VND"
	defaultRoundingPolicyCode = "VN-VND-DOC"
)

type OrgLifecycleParam struct {
	dig.In

	Jurisdiction      itTax.TaxJurisdictionDomainService
	Group             itTax.TaxGroupDomainService
	RoundingPolicy    itTax.TaxRoundingPolicyDomainService
	Classification    itTax.TaxProductClassificationDomainService
	Tax               itTax.TaxDomainService
	DefinitionVersion itTax.TaxDefinitionVersionDomainService
	RateVersion       itTax.TaxRateVersionDomainService
	Rule              itTax.TaxRuleDomainService
	RuleCondition     itTax.TaxRuleConditionDomainService
	RuleResult        itTax.TaxRuleResultDomainService
	Component         itTax.TaxComponentDomainService
	Mapping           itTax.TaxMappingDomainService
	MappingLine       itTax.TaxMappingLineDomainService

	CurrencySvc itExt.CurrencyExtService
	SettingsSvc itExt.OrgSettingsSeedExtService
}

func NewOrgLifecycleAppService(param OrgLifecycleParam) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{param: param}
}

// OrgLifecycleAppServiceImpl gives a new org the Vietnamese tax baseline and its currency.
type OrgLifecycleAppServiceImpl struct {
	param OrgLifecycleParam
}

// writeOrder lists the defaulted resources parents first.
func (this *OrgLifecycleAppServiceImpl) writeOrder() []composable.CrudDomainService {
	p := this.param
	return []composable.CrudDomainService{
		p.Jurisdiction, p.Group, p.RoundingPolicy, p.Classification, p.Tax,
		p.DefinitionVersion, p.RateVersion, p.Rule, p.RuleCondition, p.RuleResult,
	}
}

func (this *OrgLifecycleAppServiceImpl) InitOrg(ctx corectx.Context, orgId model.Id) error {
	return composable.InTransaction(ctx, this.param.Tax.Repository(), func(tranxCtx corectx.Context) error {
		aliases := defaults.NewAliases()
		for _, svc := range this.writeOrder() {
			rows, err := defaults.Load(orgDefaults.Files, svc.Schema().Name())
			if err != nil {
				return err
			}
			if err := composable.CreateDefaults(tranxCtx, svc, rows, aliases, orgId); err != nil {
				return err
			}
		}
		return this.seedSettings(tranxCtx)
	})
}

func (this *OrgLifecycleAppServiceImpl) seedSettings(ctx corectx.Context) error {
	currency, err := this.param.CurrencySvc.GetCurrencyByCode(ctx, itExt.GetCurrencyByCodeQuery{Code: defaultCurrencyCode})
	if err != nil {
		return err
	}
	if currency.ClientErrors.Count() > 0 || !currency.HasData {
		return errors.Errorf("default currency %s is not available: %v",
			defaultCurrencyCode, currency.ClientErrors.ToError())
	}

	result, err := this.param.SettingsSvc.SeedOrgSettings(ctx, itExt.SetSettingsCommand{
		ModuleKey: modconstants.AccountingModuleName,
		Items: []itExt.SetSettingItem{
			{Name: orgSettings.OrgSettingOrgDefaultCurrency, Value: string(currency.Data.Id)},
			{Name: orgSettings.OrgSettingDefaultRoundingPolicyCode, Value: defaultRoundingPolicyCode},
		},
	})
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "seed accounting settings")
	}
	return nil
}

// CleanUpOrg deletes children before the parents they reference.
func (this *OrgLifecycleAppServiceImpl) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	p := this.param
	return composable.InTransaction(ctx, p.Tax.Repository(), func(tranxCtx corectx.Context) error {
		return composable.DeleteOrgData(tranxCtx, orgId,
			p.RuleResult, p.RuleCondition, p.Rule, p.MappingLine, p.Mapping, p.Component,
			p.RateVersion, p.DefinitionVersion, p.Tax, p.Classification, p.RoundingPolicy, p.Group, p.Jurisdiction,
		)
	})
}
