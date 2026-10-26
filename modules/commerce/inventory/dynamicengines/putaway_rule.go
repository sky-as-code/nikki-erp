package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type putawayRuleEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_putaway_rule"`
}

// registerPutawayRuleEngine declares the putaway_rule onion and publishes its typed layers.
func registerPutawayRuleEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.PutawayRuleSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.PutawayRuleSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewPutawayRuleRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewPutawayRuleDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewPutawayRuleApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p putawayRuleEngineParam) itWarehouse.PutawayRuleRepository {
			return p.Engine.Repository().(itWarehouse.PutawayRuleRepository)
		},
		func(p putawayRuleEngineParam) itWarehouse.PutawayRuleDomainService {
			return p.Engine.DomainService().(itWarehouse.PutawayRuleDomainService)
		},
		func(p putawayRuleEngineParam) itWarehouse.PutawayRuleApplicationService {
			return p.Engine.ApplicationService().(itWarehouse.PutawayRuleApplicationService)
		},
	))
}
