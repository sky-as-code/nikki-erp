package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"
)

func NewOrderRepository(base dynamicresource.CrudRepository) itOrder.OrderCrudRepository {
	return &OrderRepositoryImpl{CrudRepository: base}
}

type OrderRepositoryImpl struct {
	dynamicresource.CrudRepository
}
