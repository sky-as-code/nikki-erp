package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/essential/infra/repository"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
)

type currencyEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_currency"`
}

// registerCurrencyEngine declares the currency onion and publishes its typed layers.
//
// CurrencyAppService is published as well as CurrencyApplicationService: it is the narrow
// capability sales, purchase and accounting already bind to, and it must keep resolving.
func registerCurrencyEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.CurrencySchemaName),
		func(
			param dynamicresource.BuildParam, activation *services.CurrencyActivation,
		) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.CurrencySchemaName,
				IsOrgBound: true,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewCurrencyRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewCurrencyDomainService(base, activation)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewCurrencyApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p currencyEngineParam) itCurrency.CurrencyRepository {
			return p.Engine.Repository().(itCurrency.CurrencyRepository)
		},
		func(p currencyEngineParam) itCurrency.CurrencyDomainService {
			return p.Engine.DomainService().(itCurrency.CurrencyDomainService)
		},
		func(p currencyEngineParam) itCurrency.CurrencyApplicationService {
			return p.Engine.ApplicationService().(itCurrency.CurrencyApplicationService)
		},
		func(p currencyEngineParam) itCurrency.CurrencyAppService {
			return p.Engine.ApplicationService().(itCurrency.CurrencyAppService)
		},
	))
}
