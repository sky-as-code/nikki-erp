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

type taxMappingEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_mapping"`
}

func registerTaxMappingEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxMappingSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxMappingSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxMappingRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxMappingDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxMappingApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxMappingEngineParam) taxconfig.TaxMappingRepository {
			return p.Engine.Repository().(taxconfig.TaxMappingRepository)
		},
		func(p taxMappingEngineParam) taxconfig.TaxMappingDomainService {
			return p.Engine.DomainService().(taxconfig.TaxMappingDomainService)
		},
		func(p taxMappingEngineParam) taxconfig.TaxMappingApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxMappingApplicationService)
		},
	))
}
