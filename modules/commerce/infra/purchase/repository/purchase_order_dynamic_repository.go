package repository

import (
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/order"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewPurchaseOrderRepository(base dynamicresource.CrudRepository) itOrder.PurchaseOrderCrudRepository {
	return &PurchaseOrderRepositoryImpl{CrudRepository: base}
}

type PurchaseOrderRepositoryImpl struct {
	dynamicresource.CrudRepository
}
