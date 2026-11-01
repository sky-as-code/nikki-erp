package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"
)

// The container names the REST structs inject. A dig name tag must be a literal, so routes_test.go
// asserts each still equals dynamicresource.EngineDependencyName of its schema.
const (
	PaymentMethodEngineName  = "dynengine_paymentinvoice_payment_method"
	TransactionEngineName    = "dynengine_paymentinvoice_transaction"
	InvoiceLineEngineName    = "dynengine_paymentinvoice_invoice_line"
	PaymentProfileEngineName = "dynengine_paymentinvoice_payment_profile"
	OrderEngineName          = "dynengine_paymentinvoice_order"
	InvoiceEngineName        = "dynengine_paymentinvoice_invoice"
)

type paymentMethodRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_payment_method"`
}

func NewPaymentMethodRest(params paymentMethodRestParams) *PaymentMethodRest {
	rest := &PaymentMethodRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type PaymentMethodRest struct {
	dynamicresource.CrudRestBase
}

type transactionRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_transaction"`
}

func NewTransactionRest(params transactionRestParams) *TransactionRest {
	rest := &TransactionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type TransactionRest struct {
	dynamicresource.CrudRestBase
}

type invoiceLineRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_invoice_line"`
}

func NewInvoiceLineRest(params invoiceLineRestParams) *InvoiceLineRest {
	rest := &InvoiceLineRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type InvoiceLineRest struct {
	dynamicresource.CrudRestBase
}

type paymentProfileRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_payment_profile"`
}

func NewPaymentProfileRest(params paymentProfileRestParams) *PaymentProfileRest {
	rest := &PaymentProfileRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type PaymentProfileRest struct {
	dynamicresource.CrudRestBase
}

type orderRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_order"`
}

func NewOrderRest(params orderRestParams) *OrderRest {
	rest := &OrderRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.orderSvc = params.Engine.ApplicationService().(itOrder.OrderCrudApplicationService)
	return rest
}

type OrderRest struct {
	dynamicresource.CrudRestBase
	orderSvc itOrder.OrderCrudApplicationService
}

type invoiceRestParams struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_invoice"`
}

func NewInvoiceRest(params invoiceRestParams) *InvoiceRest {
	rest := &InvoiceRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.invoiceSvc = params.Engine.ApplicationService().(itInvoice.InvoiceCrudApplicationService)
	return rest
}

type InvoiceRest struct {
	dynamicresource.CrudRestBase
	invoiceSvc itInvoice.InvoiceCrudApplicationService
}
