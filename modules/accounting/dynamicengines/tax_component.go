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

type taxComponentEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_component"`
}

// taxComponentBuildParam is composable.BuildParam plus the sibling repository the rules read.
type taxComponentBuildParam struct {
	composable.BuildParam

	Versions taxconfig.TaxDefinitionVersionRepository
}

func registerTaxComponentEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxComponentSchemaName),
		func(param taxComponentBuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxComponentSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxComponentRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxComponentDomainService(base, param.Versions)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxComponentApplicationService(base)
				},
			}, param.BuildParam)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxComponentEngineParam) taxconfig.TaxComponentRepository {
			return p.Engine.Repository().(taxconfig.TaxComponentRepository)
		},
		func(p taxComponentEngineParam) taxconfig.TaxComponentDomainService {
			return p.Engine.DomainService().(taxconfig.TaxComponentDomainService)
		},
		func(p taxComponentEngineParam) taxconfig.TaxComponentApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxComponentApplicationService)
		},
	))
}
