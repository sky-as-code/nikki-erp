package repository

import (
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/order"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewPurchaseOrderLineRepository(base dynamicresource.CrudRepository) itOrder.PurchaseOrderLineCrudRepository {
	return &PurchaseOrderLineRepositoryImpl{CrudRepository: base}
}

type PurchaseOrderLineRepositoryImpl struct {
	dynamicresource.CrudRepository
}
