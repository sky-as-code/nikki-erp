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

type taxRuleResultEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rule_result"`
}

func registerTaxRuleResultEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxRuleResultSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxRuleResultSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxRuleResultRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxRuleResultDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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
