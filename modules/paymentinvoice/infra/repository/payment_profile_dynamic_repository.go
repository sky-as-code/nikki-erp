package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itPaymentProfile "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentprofile"
)

func NewPaymentProfileRepository(base composable.CrudRepository) itPaymentProfile.PaymentProfileCrudRepository {
	return &PaymentProfileRepositoryImpl{CrudRepository: base}
}

type PaymentProfileRepositoryImpl struct {
	composable.CrudRepository
}
