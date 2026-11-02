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

type taxMappingLineEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_mapping_line"`
}

// taxMappingLineBuildParam is dynamicresource.BuildParam plus the sibling repository the rules read.
type taxMappingLineBuildParam struct {
	dynamicresource.BuildParam

	Mappings taxconfig.TaxMappingRepository
}

func registerTaxMappingLineEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.TaxMappingLineSchemaName),
		func(param taxMappingLineBuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxMappingLineSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewTaxMappingLineRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewTaxMappingLineDomainService(base, param.Mappings)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
