package constants

const PaymentInvoiceModuleName = "paymentinvoice"

// Resource codes for authorization. They are byte-identical to the dynamic-model schema names,
// because the dynamic resource engine asserts permissions using the schema name as the resource
// code. A code that drifts from its schema name denies every request with no obvious cause.
const (
	ResourceOrder       = "paymentinvoice_order"
	ResourceTransaction = "paymentinvoice_transaction"
	ResourceInvoice     = "paymentinvoice_invoice"
	ResourceInvoiceLine = "paymentinvoice_invoice_line"

	// ResourcePaymentProfile guards a set of gateway credentials. It is its own resource rather
	// than a field of the payment method: reading which methods a payer may choose is an everyday
	// permission, while reading the merchant credentials those payments settle into is not, and
	// one code cannot express both.
	ResourcePaymentProfile = "paymentinvoice_payment_profile"
)
