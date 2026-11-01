package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"
	itPaymentMethod "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
	itTransaction "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/transaction"
)

// The CRUD domain services of the resources whose rules live outside CRUD: an order and its
// transactions are written by the payment flow (order_*.go), an invoice by Issue and
// IssueFromSource. Each still gets its own layer so a rule has a home the day it is needed.

func NewPaymentMethodCrudDomainService(
	base dynamicresource.CrudDomainService,
) itPaymentMethod.PaymentMethodCrudDomainService {
	return &PaymentMethodCrudDomainServiceImpl{CrudDomainService: base}
}

type PaymentMethodCrudDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}

func NewTransactionCrudDomainService(base dynamicresource.CrudDomainService) itTransaction.TransactionCrudDomainService {
	return &TransactionCrudDomainServiceImpl{CrudDomainService: base}
}

type TransactionCrudDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}

func NewOrderCrudDomainService(base dynamicresource.CrudDomainService) itOrder.OrderCrudDomainService {
	return &OrderCrudDomainServiceImpl{CrudDomainService: base}
}

type OrderCrudDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}

func NewInvoiceCrudDomainService(base dynamicresource.CrudDomainService) itInvoice.InvoiceCrudDomainService {
	return &InvoiceCrudDomainServiceImpl{CrudDomainService: base}
}

type InvoiceCrudDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}

func NewInvoiceLineCrudDomainService(base dynamicresource.CrudDomainService) itInvoice.InvoiceLineCrudDomainService {
	return &InvoiceLineCrudDomainServiceImpl{CrudDomainService: base}
}

type InvoiceLineCrudDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}
