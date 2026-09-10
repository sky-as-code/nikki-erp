package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/inventory/infra/repository"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

type inventoryLocationEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_location"`
}

// registerInventoryLocationEngine declares the inventory_location onion and publishes its typed layers.
func registerInventoryLocationEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.InventoryLocationSchemaName),
		func(param composable.BuildParam, usage itStock.LocationUsageReadService) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.InventoryLocationSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewInventoryLocationRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewInventoryLocationDomainService(base, usage)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewInventoryLocationApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p inventoryLocationEngineParam) itWarehouse.InventoryLocationRepository {
			return p.Engine.Repository().(itWarehouse.InventoryLocationRepository)
		},
		func(p inventoryLocationEngineParam) itWarehouse.InventoryLocationDomainService {
			return p.Engine.DomainService().(itWarehouse.InventoryLocationDomainService)
		},
		func(p inventoryLocationEngineParam) itWarehouse.InventoryLocationApplicationService {
			return p.Engine.ApplicationService().(itWarehouse.InventoryLocationApplicationService)
		},
		// The warehouse orchestration provisions locations through the concrete service.
		func(p inventoryLocationEngineParam) *services.InventoryLocationDomainServiceImpl {
			return p.Engine.DomainService().(*services.InventoryLocationDomainServiceImpl)
		},
	))
}
