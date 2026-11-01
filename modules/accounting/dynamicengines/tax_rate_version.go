package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/accounting/app"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/accounting/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type taxRateVersionEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rate_version"`
}

func registerTaxRateVersionEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.TaxRateVersionSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxRateVersionSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewTaxRateVersionRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewTaxRateVersionDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewTaxRateVersionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxRateVersionEngineParam) taxconfig.TaxRateVersionRepository {
			return p.Engine.Repository().(taxconfig.TaxRateVersionRepository)
		},
		func(p taxRateVersionEngineParam) taxconfig.TaxRateVersionDomainService {
			return p.Engine.DomainService().(taxconfig.TaxRateVersionDomainService)
		},
		func(p taxRateVersionEngineParam) taxconfig.TaxRateVersionApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxRateVersionApplicationService)
		},
	))
}
