package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// The engine declarations of the module's resources.
//
// A resource that names no DefineActions is served by the built-in CRUD alone, which is enough to
// read what the module has recorded. The three that do add actions add them for different reasons:
// the order and the invoice for operations that are not CRUD at all (taking money, giving it back,
// closing a draft), the payment profile to encrypt its credentials either side of the CRUD it
// already has.

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

// The Payment Profile engine.
//
// It carries the merchant credentials a payment settles into, so its listing deliberately shows
// only what identifies a profile: a search that selects nothing specific returns these three
// fields and no credentials, which keeps the secrets out of every screen that merely lists what is
// configured. Reading one profile still returns them, as does a search that names "config"
// explicitly — both of which the read permission is what guards. See payment_profile_actions.go
// for the encryption either side of the built-in CRUD.
func paymentProfileEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.PaymentProfileSchemaName,
		DefaultFields: []string{
			models.PaymentProfileFieldName,
			models.PaymentProfileFieldMethod,
			models.PaymentProfileFieldOrgId,
		},
		DefineActions: definePaymentProfileActions,
	}
}
