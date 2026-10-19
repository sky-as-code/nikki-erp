package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
)

func NewInvoiceLineRepository(base composable.CrudRepository) itInvoice.InvoiceLineCrudRepository {
	return &InvoiceLineRepositoryImpl{CrudRepository: base}
}

type InvoiceLineRepositoryImpl struct {
	composable.CrudRepository
}
