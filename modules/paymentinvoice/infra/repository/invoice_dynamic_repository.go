package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
)

func NewInvoiceRepository(base dynamicresource.CrudRepository) itInvoice.InvoiceCrudRepository {
	return &InvoiceRepositoryImpl{CrudRepository: base}
}

type InvoiceRepositoryImpl struct {
	dynamicresource.CrudRepository
}
