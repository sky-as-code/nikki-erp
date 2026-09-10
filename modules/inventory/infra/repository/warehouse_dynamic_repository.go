package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewWarehouseRepository is handed the composable default by the warehouse onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewWarehouseRepository(base composable.CrudRepository) itWarehouse.WarehouseRepository {
	return &WarehouseRepositoryImpl{CrudRepository: base}
}

type WarehouseRepositoryImpl struct {
	composable.CrudRepository
}
