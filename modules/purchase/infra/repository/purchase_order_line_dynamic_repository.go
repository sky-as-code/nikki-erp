package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrder "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/order"
)

func NewPurchaseOrderLineRepository(base composable.CrudRepository) itOrder.PurchaseOrderLineCrudRepository {
	return &PurchaseOrderLineRepositoryImpl{CrudRepository: base}
}

type PurchaseOrderLineRepositoryImpl struct {
	composable.CrudRepository
}
