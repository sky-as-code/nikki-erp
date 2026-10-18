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
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type taxRateVersionEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rate_version"`
}

func registerTaxRateVersionEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxRateVersionSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxRateVersionSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxRateVersionRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxRateVersionDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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
