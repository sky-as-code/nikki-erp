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

type taxRuleConditionEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rule_condition"`
}

func registerTaxRuleConditionEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.TaxRuleConditionSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxRuleConditionSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewTaxRuleConditionRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewTaxRuleConditionDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
