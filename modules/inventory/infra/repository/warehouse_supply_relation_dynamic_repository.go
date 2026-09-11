package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewWarehouseSupplyRelationRepository is handed the composable default by the warehouse_supply_relation onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewWarehouseSupplyRelationRepository(base composable.CrudRepository) itWarehouse.WarehouseSupplyRelationRepository {
	return &WarehouseSupplyRelationRepositoryImpl{CrudRepository: base}
}

type WarehouseSupplyRelationRepositoryImpl struct {
	composable.CrudRepository
}
