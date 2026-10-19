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

type taxJurisdictionEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_jurisdiction"`
}

func registerTaxJurisdictionEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxJurisdictionSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxJurisdictionSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxJurisdictionRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxJurisdictionDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxJurisdictionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxJurisdictionEngineParam) taxconfig.TaxJurisdictionRepository {
			return p.Engine.Repository().(taxconfig.TaxJurisdictionRepository)
		},
		func(p taxJurisdictionEngineParam) taxconfig.TaxJurisdictionDomainService {
			return p.Engine.DomainService().(taxconfig.TaxJurisdictionDomainService)
		},
		func(p taxJurisdictionEngineParam) taxconfig.TaxJurisdictionApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxJurisdictionApplicationService)
		},
	))
}
