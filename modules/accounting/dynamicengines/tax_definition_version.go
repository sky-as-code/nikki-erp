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

type taxDefinitionVersionEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_definition_version"`
}

func registerTaxDefinitionVersionEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxDefinitionVersionSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxDefinitionVersionSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxDefinitionVersionRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxDefinitionVersionDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxDefinitionVersionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxDefinitionVersionEngineParam) taxconfig.TaxDefinitionVersionRepository {
			return p.Engine.Repository().(taxconfig.TaxDefinitionVersionRepository)
		},
		func(p taxDefinitionVersionEngineParam) taxconfig.TaxDefinitionVersionDomainService {
			return p.Engine.DomainService().(taxconfig.TaxDefinitionVersionDomainService)
		},
		func(p taxDefinitionVersionEngineParam) taxconfig.TaxDefinitionVersionApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxDefinitionVersionApplicationService)
		},
	))
}
