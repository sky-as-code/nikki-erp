package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

func NewWarehouseProductGuardRepository(base composable.CrudRepository) itStock.WarehouseProductGuardRepository {
	return &WarehouseProductGuardRepositoryImpl{CrudRepository: base}
}

type WarehouseProductGuardRepositoryImpl struct {
	composable.CrudRepository
}
