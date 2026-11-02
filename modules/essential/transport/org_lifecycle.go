package transport

import (
	"context"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	c "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
	itLanguage "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

type orgLifecycleDeps struct {
	dig.In

	UomCatSvc    itUomCat.UomCatDomainService
	UomSvc       itUom.UomDomainService
	LanguageSvc  itLanguage.LanguageDomainService
	CurrencySvc  itCurrency.CurrencyDomainService
	Languages    *services.LanguageActivation
	Currencies   *services.CurrencyActivation
	SettingsSeed itExt.OrgSettingsSeedExtService
}

func initOrgLifecycleHandlers() error {
	svc := orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(param orgLifecycleDeps) {
			built = app.NewOrgLifecycleAppService(app.OrgLifecycleParam{
				UomCatSvc: param.UomCatSvc, UomSvc: param.UomSvc,
				LanguageSvc: param.LanguageSvc, CurrencySvc: param.CurrencySvc,
				Languages: param.Languages, Currencies: param.Currencies, SettingsSeed: param.SettingsSeed,
			})
		})
		return built, err
	})
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return orglifecycle.RegisterHandlers(context.Background(), bus, c.EssentialModuleName, svc)
	})
}
