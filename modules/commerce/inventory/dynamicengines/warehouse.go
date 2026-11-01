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
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type warehouseEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_warehouse"`
}

// registerWarehouseEngine declares the warehouse onion and publishes its typed layers.
func registerWarehouseEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.WarehouseSchemaName),
		func(param dynamicresource.BuildParam, locationSvc *services.InventoryLocationDomainServiceImpl) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.WarehouseSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewWarehouseRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewWarehouseDomainService(base, locationSvc)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
