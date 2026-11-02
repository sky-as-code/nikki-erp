package v1

import (
	"go.uber.org/dig"

	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// InventoryLocationEngineName is the container name of the inventory_location onion, declared once for the struct tag and
// checked against dynamicresource.EngineDependencyName by a test.
const InventoryLocationEngineName = "dynengine_inventory_location"

type inventoryLocationRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_inventory_location"`
}

func NewInventoryLocationRest(params inventoryLocationRestParams) *InventoryLocationRest {
	rest := &InventoryLocationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.inventoryLocationSvc = params.Engine.ApplicationService().(itWarehouse.InventoryLocationApplicationService)
	return rest
}

// InventoryLocationRest serves the built-in CRUD of the inventory_location resource.
type InventoryLocationRest struct {
	dynamicresource.CrudRestBase
	inventoryLocationSvc itWarehouse.InventoryLocationApplicationService
}
