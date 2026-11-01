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

type taxJurisdictionEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_jurisdiction"`
}

func registerTaxJurisdictionEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.TaxJurisdictionSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxJurisdictionSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewTaxJurisdictionRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewTaxJurisdictionDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
