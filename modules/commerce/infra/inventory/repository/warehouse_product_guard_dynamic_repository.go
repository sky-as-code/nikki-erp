package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewWarehouseProductGuardRepository(base composable.CrudRepository) itStock.WarehouseProductGuardRepository {
	return &WarehouseProductGuardRepositoryImpl{CrudRepository: base}
}

type WarehouseProductGuardRepositoryImpl struct {
	composable.CrudRepository
}
