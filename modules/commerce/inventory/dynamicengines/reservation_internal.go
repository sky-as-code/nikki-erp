package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The two internal resources of the warehouse reservation feature. Neither is served over REST:
// the guard is a lock target and the outbox is the module's own event queue. They are onions
// only so their repositories exist for the services that use them, and BuildInternalEngines
// forces their construction because no REST handler will.

type warehouseProductGuardEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_warehouse_product_guard"`
}

func registerWarehouseProductGuardEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.WarehouseProductGuardSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.WarehouseProductGuardSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewWarehouseProductGuardRepository(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p warehouseProductGuardEngineParam) itStock.WarehouseProductGuardRepository {
			return p.Engine.Repository().(itStock.WarehouseProductGuardRepository)
		},
	))
}

type inventoryIntegrationOutboxEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_integration_outbox"`
}

func registerInventoryIntegrationOutboxEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.InventoryIntegrationOutboxSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:  models.InventoryIntegrationOutboxSchemaName,
				CrudActions: readOnlyCrudActions(),
				// Not audited: the outbox is delivery plumbing, and every event it carries describes a
				// write the audit log already recorded.
				IsAudited: util.ToPtr(false),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewInventoryIntegrationOutboxRepository(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p inventoryIntegrationOutboxEngineParam) itStock.InventoryIntegrationOutboxRepository {
			return p.Engine.Repository().(itStock.InventoryIntegrationOutboxRepository)
		},
	))
}

// BuildInternalEngines resolves the onions no REST handler resolves, so the boot-time check that
// every declared resource was built holds for them too. Called from the module's Init after the
// REST handlers have been registered.
func BuildInternalEngines() error {
	return deps.Invoke(func(
		guard warehouseProductGuardEngineParam, outbox inventoryIntegrationOutboxEngineParam,
	) error {
		return nil
	})
}
