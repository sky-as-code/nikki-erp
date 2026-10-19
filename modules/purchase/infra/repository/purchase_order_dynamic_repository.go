package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrder "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/order"
)

func NewPurchaseOrderRepository(base composable.CrudRepository) itOrder.PurchaseOrderCrudRepository {
	return &PurchaseOrderRepositoryImpl{CrudRepository: base}
}

type PurchaseOrderRepositoryImpl struct {
	composable.CrudRepository
}
