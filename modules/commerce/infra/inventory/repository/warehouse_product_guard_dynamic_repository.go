package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewWarehouseProductGuardRepository(base dynamicresource.CrudRepository) itStock.WarehouseProductGuardRepository {
	return &WarehouseProductGuardRepositoryImpl{CrudRepository: base}
}

type WarehouseProductGuardRepositoryImpl struct {
	dynamicresource.CrudRepository
}
