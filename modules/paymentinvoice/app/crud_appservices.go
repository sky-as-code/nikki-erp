package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
	itPaymentProfile "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentprofile"
	itTransaction "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/transaction"
)

// The application services of the resources that add nothing to the built-in CRUD. The default
// resolves the org, asserts the permission and confines single-row actions to the org.

func NewPaymentProfileApplicationService(
	base composable.CrudApplicationService,
) itPaymentProfile.PaymentProfileCrudApplicationService {
	return &PaymentProfileApplicationServiceImpl{CrudApplicationService: base}
}

type PaymentProfileApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewTransactionApplicationService(
	base composable.CrudApplicationService,
) itTransaction.TransactionCrudApplicationService {
	return &TransactionApplicationServiceImpl{CrudApplicationService: base}
}

type TransactionApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewInvoiceLineApplicationService(
	base composable.CrudApplicationService,
) itInvoice.InvoiceLineCrudApplicationService {
	return &InvoiceLineApplicationServiceImpl{CrudApplicationService: base}
}

type InvoiceLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}
