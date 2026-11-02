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

type taxRuleResultEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rule_result"`
}

func registerTaxRuleResultEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.TaxRuleResultSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxRuleResultSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewTaxRuleResultRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewTaxRuleResultDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewTaxRuleResultApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxRuleResultEngineParam) taxconfig.TaxRuleResultRepository {
			return p.Engine.Repository().(taxconfig.TaxRuleResultRepository)
		},
		func(p taxRuleResultEngineParam) taxconfig.TaxRuleResultDomainService {
			return p.Engine.DomainService().(taxconfig.TaxRuleResultDomainService)
		},
		func(p taxRuleResultEngineParam) taxconfig.TaxRuleResultApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxRuleResultApplicationService)
		},
	))
}
