package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itPaymentMethod "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
)

func NewPaymentMethodRepository(base dynamicresource.CrudRepository) itPaymentMethod.PaymentMethodCrudRepository {
	return &PaymentMethodRepositoryImpl{CrudRepository: base}
}

type PaymentMethodRepositoryImpl struct {
	dynamicresource.CrudRepository
}
