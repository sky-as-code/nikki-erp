package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
)

func NewInvoiceRepository(base composable.CrudRepository) itInvoice.InvoiceCrudRepository {
	return &InvoiceRepositoryImpl{CrudRepository: base}
}

type InvoiceRepositoryImpl struct {
	composable.CrudRepository
}
