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

type taxRoundingPolicyEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_rounding_policy"`
}

func registerTaxRoundingPolicyEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxRoundingPolicySchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxRoundingPolicySchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxRoundingPolicyRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxRoundingPolicyDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxRoundingPolicyApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxRoundingPolicyEngineParam) taxconfig.TaxRoundingPolicyRepository {
			return p.Engine.Repository().(taxconfig.TaxRoundingPolicyRepository)
		},
		func(p taxRoundingPolicyEngineParam) taxconfig.TaxRoundingPolicyDomainService {
			return p.Engine.DomainService().(taxconfig.TaxRoundingPolicyDomainService)
		},
		func(p taxRoundingPolicyEngineParam) taxconfig.TaxRoundingPolicyApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxRoundingPolicyApplicationService)
		},
	))
}
