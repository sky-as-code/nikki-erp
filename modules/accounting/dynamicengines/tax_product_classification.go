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

type taxProductClassificationEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_product_classification"`
}

func registerTaxProductClassificationEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxProductClassificationSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxProductClassificationSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxProductClassificationRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxProductClassificationDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxProductClassificationApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxProductClassificationEngineParam) taxconfig.TaxProductClassificationRepository {
			return p.Engine.Repository().(taxconfig.TaxProductClassificationRepository)
		},
		func(p taxProductClassificationEngineParam) taxconfig.TaxProductClassificationDomainService {
			return p.Engine.DomainService().(taxconfig.TaxProductClassificationDomainService)
		},
		func(p taxProductClassificationEngineParam) taxconfig.TaxProductClassificationApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxProductClassificationApplicationService)
		},
	))
}
