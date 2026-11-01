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

type warehouseSupplyRelationEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_warehouse_supply_relation"`
}

// registerWarehouseSupplyRelationEngine declares the warehouse_supply_relation onion and publishes its typed layers.
func registerWarehouseSupplyRelationEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.WarehouseSupplyRelationSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.WarehouseSupplyRelationSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewWarehouseSupplyRelationRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSupplyRelationDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
