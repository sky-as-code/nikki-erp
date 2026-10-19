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

type taxRuleEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rule"`
}

func registerTaxRuleEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxRuleSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxRuleSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxRuleRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxRuleDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxRuleApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxRuleEngineParam) taxconfig.TaxRuleRepository {
			return p.Engine.Repository().(taxconfig.TaxRuleRepository)
		},
		func(p taxRuleEngineParam) taxconfig.TaxRuleDomainService {
			return p.Engine.DomainService().(taxconfig.TaxRuleDomainService)
		},
		func(p taxRuleEngineParam) taxconfig.TaxRuleApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxRuleApplicationService)
		},
	))
}
