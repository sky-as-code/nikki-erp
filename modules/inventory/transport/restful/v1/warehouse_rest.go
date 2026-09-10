package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// WarehouseEngineName is the container name of the warehouse onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const WarehouseEngineName = "dynengine_inventory_warehouse"

type warehouseRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_warehouse"`
}

func NewWarehouseRest(params warehouseRestParams) *WarehouseRest {
	rest := &WarehouseRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.warehouseSvc = params.Engine.ApplicationService().(itWarehouse.WarehouseApplicationService)
	return rest
}

// WarehouseRest serves the built-in CRUD of the warehouse resource.
type WarehouseRest struct {
	composable.CrudRestBase
	warehouseSvc itWarehouse.WarehouseApplicationService
}
