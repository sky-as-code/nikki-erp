package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// The engine declarations of the module's four resources.
//
// None of them defines an action yet: create_payment, refund and remove_pos_orders arrive with the
// order services, and issue with the invoice ones. Until then every resource is served by the
// built-in CRUD alone, which is enough to read what the module has recorded.

// The Payment Method engine.
//
// The method list is data rather than an enum, so this engine is how a deployment adds a gateway
// account, withdraws one, or adjusts the amounts a gateway will accept — none of which should
// require a release. What stays in code is the adapter named by adapter_code.
func paymentMethodEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.PaymentMethodSchemaName,
		DefaultFields: []string{
			models.PaymentMethodFieldCode,
			models.PaymentMethodFieldName,
			models.PaymentMethodFieldAdapterCode,
			models.PaymentMethodFieldCurrencyId,
			models.PaymentMethodFieldIsActive,
		},
	}
}

func orderEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.OrderSchemaName,
		DefaultFields: []string{
			models.OrderFieldOrderId,
			models.OrderFieldStatus,
			models.OrderFieldAmount,
			models.OrderFieldCurrencyId,
			models.OrderFieldPaymentMethodId,
			models.OrderFieldSource,
		},
		DefineActions: defineOrderActions,
	}
}

func transactionEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.TransactionSchemaName,
		DefaultFields: []string{
			models.TransactionFieldOrderBusinessId,
			models.TransactionFieldTransactionType,
			models.TransactionFieldStatus,
			models.TransactionFieldAmount,
			models.TransactionFieldCurrencyId,
			models.TransactionFieldPaymentMethodId,
			models.TransactionFieldRefTransactionId,
		},
	}
}

func invoiceEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.InvoiceSchemaName,
		DefaultFields: []string{
			models.InvoiceFieldNumber,
			models.InvoiceFieldStatus,
			models.InvoiceFieldPartnerName,
			models.InvoiceFieldTotalAmount,
			models.InvoiceFieldCurrencyId,
			models.InvoiceFieldIssuedAt,
		},
		DefineActions: defineInvoiceActions,
	}
}

func invoiceLineEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.InvoiceLineSchemaName,
		DefaultFields: []string{
			models.InvoiceLineFieldInvoiceId,
			models.InvoiceLineFieldDescription,
			models.InvoiceLineFieldQuantity,
			models.InvoiceLineFieldUnitPrice,
			models.InvoiceLineFieldTaxRatePercent,
			models.InvoiceLineFieldAmount,
		},
	}
}
