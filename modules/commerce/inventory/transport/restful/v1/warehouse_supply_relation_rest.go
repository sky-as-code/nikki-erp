package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// WarehouseSupplyRelationEngineName is the container name of the warehouse_supply_relation onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const WarehouseSupplyRelationEngineName = "dynengine_inventory_warehouse_supply_relation"

type warehouseSupplyRelationRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_warehouse_supply_relation"`
}

func NewWarehouseSupplyRelationRest(params warehouseSupplyRelationRestParams) *WarehouseSupplyRelationRest {
	rest := &WarehouseSupplyRelationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.warehouseSupplyRelationSvc = params.Engine.ApplicationService().(itWarehouse.WarehouseSupplyRelationApplicationService)
	return rest
}

// WarehouseSupplyRelationRest serves the built-in CRUD of the warehouse_supply_relation resource.
type WarehouseSupplyRelationRest struct {
	composable.CrudRestBase
	warehouseSupplyRelationSvc itWarehouse.WarehouseSupplyRelationApplicationService
}
