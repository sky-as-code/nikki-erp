package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewInventoryLocationRepository is handed the composable default by the inventory_location onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewInventoryLocationRepository(base composable.CrudRepository) itWarehouse.InventoryLocationRepository {
	return &InventoryLocationRepositoryImpl{CrudRepository: base}
}

type InventoryLocationRepositoryImpl struct {
	composable.CrudRepository
}
