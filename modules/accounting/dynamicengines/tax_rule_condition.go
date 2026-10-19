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

type taxRuleConditionEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rule_condition"`
}

func registerTaxRuleConditionEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxRuleConditionSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxRuleConditionSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxRuleConditionRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxRuleConditionDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxRuleConditionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxRuleConditionEngineParam) taxconfig.TaxRuleConditionRepository {
			return p.Engine.Repository().(taxconfig.TaxRuleConditionRepository)
		},
		func(p taxRuleConditionEngineParam) taxconfig.TaxRuleConditionDomainService {
			return p.Engine.DomainService().(taxconfig.TaxRuleConditionDomainService)
		},
		func(p taxRuleConditionEngineParam) taxconfig.TaxRuleConditionApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxRuleConditionApplicationService)
		},
	))
}
