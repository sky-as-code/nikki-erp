package repository

import (
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewWarehouseSupplyRelationRepository is handed the composable default by the warehouse_supply_relation onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewWarehouseSupplyRelationRepository(base composable.CrudRepository) itWarehouse.WarehouseSupplyRelationRepository {
	return &WarehouseSupplyRelationRepositoryImpl{CrudRepository: base}
}

type WarehouseSupplyRelationRepositoryImpl struct {
	composable.CrudRepository
}
