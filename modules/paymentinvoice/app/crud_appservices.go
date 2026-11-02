package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
	itPaymentProfile "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentprofile"
	itTransaction "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/transaction"
)

// The application services of the resources that add nothing to the built-in CRUD. The default
// resolves the org, asserts the permission and confines single-row actions to the org.

func NewPaymentProfileApplicationService(
	base dynamicresource.CrudApplicationService,
) itPaymentProfile.PaymentProfileCrudApplicationService {
	return &PaymentProfileApplicationServiceImpl{CrudApplicationService: base}
}

type PaymentProfileApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewTransactionApplicationService(
	base dynamicresource.CrudApplicationService,
) itTransaction.TransactionCrudApplicationService {
	return &TransactionApplicationServiceImpl{CrudApplicationService: base}
}

type TransactionApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func NewInvoiceLineApplicationService(
	base dynamicresource.CrudApplicationService,
) itInvoice.InvoiceLineCrudApplicationService {
	return &InvoiceLineApplicationServiceImpl{CrudApplicationService: base}
}

type InvoiceLineApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
