package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewWarehouseSupplyRelationApplicationService is handed the composable default by the warehouse_supply_relation onion.
func NewWarehouseSupplyRelationApplicationService(base composable.CrudApplicationService) itWarehouse.WarehouseSupplyRelationApplicationService {
	return &WarehouseSupplyRelationApplicationServiceImpl{CrudApplicationService: base}
}

// WarehouseSupplyRelationApplicationServiceImpl is the authorized CRUD of the resource.
type WarehouseSupplyRelationApplicationServiceImpl struct {
	composable.CrudApplicationService
}
