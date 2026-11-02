package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
)

func NewInvoiceLineRepository(base dynamicresource.CrudRepository) itInvoice.InvoiceLineCrudRepository {
	return &InvoiceLineRepositoryImpl{CrudRepository: base}
}

type InvoiceLineRepositoryImpl struct {
	dynamicresource.CrudRepository
}
