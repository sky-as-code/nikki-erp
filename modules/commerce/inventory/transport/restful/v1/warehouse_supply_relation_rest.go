package v1

import (
	"go.uber.org/dig"

	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// WarehouseSupplyRelationEngineName is the container name of the warehouse_supply_relation onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const WarehouseSupplyRelationEngineName = "dynengine_inventory_warehouse_supply_relation"

type warehouseSupplyRelationRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_warehouse_supply_relation"`
}

func NewWarehouseSupplyRelationRest(params warehouseSupplyRelationRestParams) *WarehouseSupplyRelationRest {
	rest := &WarehouseSupplyRelationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.warehouseSupplyRelationSvc = params.Engine.ApplicationService().(itWarehouse.WarehouseSupplyRelationApplicationService)
	return rest
}

// WarehouseSupplyRelationRest serves the built-in CRUD of the warehouse_supply_relation resource.
type WarehouseSupplyRelationRest struct {
	dynamicresource.CrudRestBase
	warehouseSupplyRelationSvc itWarehouse.WarehouseSupplyRelationApplicationService
}
