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
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

type warehouseSupplyRelationEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_warehouse_supply_relation"`
}

// registerWarehouseSupplyRelationEngine declares the warehouse_supply_relation onion and publishes its typed layers.
func registerWarehouseSupplyRelationEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.WarehouseSupplyRelationSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.WarehouseSupplyRelationSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewWarehouseSupplyRelationRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSupplyRelationDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewWarehouseSupplyRelationApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p warehouseSupplyRelationEngineParam) itWarehouse.WarehouseSupplyRelationRepository {
			return p.Engine.Repository().(itWarehouse.WarehouseSupplyRelationRepository)
		},
		func(p warehouseSupplyRelationEngineParam) itWarehouse.WarehouseSupplyRelationDomainService {
			return p.Engine.DomainService().(itWarehouse.WarehouseSupplyRelationDomainService)
		},
		func(p warehouseSupplyRelationEngineParam) itWarehouse.WarehouseSupplyRelationApplicationService {
			return p.Engine.ApplicationService().(itWarehouse.WarehouseSupplyRelationApplicationService)
		},
	))
}
