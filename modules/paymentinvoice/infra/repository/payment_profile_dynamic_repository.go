package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itPaymentProfile "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentprofile"
)

func NewPaymentProfileRepository(base dynamicresource.CrudRepository) itPaymentProfile.PaymentProfileCrudRepository {
	return &PaymentProfileRepositoryImpl{CrudRepository: base}
}

type PaymentProfileRepositoryImpl struct {
	dynamicresource.CrudRepository
}
