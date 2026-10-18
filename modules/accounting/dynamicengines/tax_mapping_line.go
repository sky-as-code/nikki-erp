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

type taxMappingLineEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_mapping_line"`
}

// taxMappingLineBuildParam is composable.BuildParam plus the sibling repository the rules read.
type taxMappingLineBuildParam struct {
	composable.BuildParam

	Mappings taxconfig.TaxMappingRepository
}

func registerTaxMappingLineEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxMappingLineSchemaName),
		func(param taxMappingLineBuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxMappingLineSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxMappingLineRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxMappingLineDomainService(base, param.Mappings)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxMappingLineApplicationService(base)
				},
			}, param.BuildParam)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxMappingLineEngineParam) taxconfig.TaxMappingLineRepository {
			return p.Engine.Repository().(taxconfig.TaxMappingLineRepository)
		},
		func(p taxMappingLineEngineParam) taxconfig.TaxMappingLineDomainService {
			return p.Engine.DomainService().(taxconfig.TaxMappingLineDomainService)
		},
		func(p taxMappingLineEngineParam) taxconfig.TaxMappingLineApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxMappingLineApplicationService)
		},
	))
}
