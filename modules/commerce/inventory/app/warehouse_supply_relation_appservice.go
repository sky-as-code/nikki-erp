package app

import (
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewWarehouseSupplyRelationApplicationService is handed the composable default by the warehouse_supply_relation onion.
func NewWarehouseSupplyRelationApplicationService(base composable.CrudApplicationService) itWarehouse.WarehouseSupplyRelationApplicationService {
	return &WarehouseSupplyRelationApplicationServiceImpl{CrudApplicationService: base}
}

// WarehouseSupplyRelationApplicationServiceImpl is the authorized CRUD of the resource.
type WarehouseSupplyRelationApplicationServiceImpl struct {
	composable.CrudApplicationService
}
