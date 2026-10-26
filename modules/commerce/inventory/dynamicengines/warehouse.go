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

type warehouseEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_warehouse"`
}

// registerWarehouseEngine declares the warehouse onion and publishes its typed layers.
func registerWarehouseEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.WarehouseSchemaName),
		func(param composable.BuildParam, locationSvc *services.InventoryLocationDomainServiceImpl) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.WarehouseSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewWarehouseRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewWarehouseDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewWarehouseApplicationService(base, locationSvc)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p warehouseEngineParam) itWarehouse.WarehouseRepository {
			return p.Engine.Repository().(itWarehouse.WarehouseRepository)
		},
		func(p warehouseEngineParam) itWarehouse.WarehouseDomainService {
			return p.Engine.DomainService().(itWarehouse.WarehouseDomainService)
		},
		func(p warehouseEngineParam) itWarehouse.WarehouseApplicationService {
			return p.Engine.ApplicationService().(itWarehouse.WarehouseApplicationService)
		},
		// The orchestration port: creating a warehouse with its locations and reconfiguring flows.
		func(p warehouseEngineParam) itWarehouse.WarehouseAppService {
			return p.Engine.ApplicationService().(itWarehouse.WarehouseAppService)
		},
		// The read port: listing warehouses for selection, narrower than WarehouseAppService on
		// purpose — filling a selection list must not carry the power to create a warehouse or
		// reconfigure its flows.
		func(p warehouseEngineParam) itWarehouse.WarehouseReadService {
			return p.Engine.DomainService().(itWarehouse.WarehouseReadService)
		},
	))
}
