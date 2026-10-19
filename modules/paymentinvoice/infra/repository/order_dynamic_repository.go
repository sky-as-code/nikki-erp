package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"
)

func NewOrderRepository(base composable.CrudRepository) itOrder.OrderCrudRepository {
	return &OrderRepositoryImpl{CrudRepository: base}
}

type OrderRepositoryImpl struct {
	composable.CrudRepository
}
