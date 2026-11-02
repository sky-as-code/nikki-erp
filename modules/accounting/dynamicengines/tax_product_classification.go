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

type taxProductClassificationEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_product_classification"`
}

func registerTaxProductClassificationEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.TaxProductClassificationSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxProductClassificationSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewTaxProductClassificationRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewTaxProductClassificationDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
