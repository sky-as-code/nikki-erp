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
	}
}

func orderEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.OrderSchemaName,
		DefineActions: defineOrderActions,
	}
}

func transactionEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.TransactionSchemaName,
	}
}

func invoiceEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.InvoiceSchemaName,
		DefineActions: defineInvoiceActions,
	}
}

func invoiceLineEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.InvoiceLineSchemaName,
	}
}
