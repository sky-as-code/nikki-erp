package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itPaymentMethod "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
)

func NewPaymentMethodRepository(base composable.CrudRepository) itPaymentMethod.PaymentMethodCrudRepository {
	return &PaymentMethodRepositoryImpl{CrudRepository: base}
}

type PaymentMethodRepositoryImpl struct {
	composable.CrudRepository
}
