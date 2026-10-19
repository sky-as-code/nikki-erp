// Package dynamicengines registers the composable onions of the Payment & Invoice resources and
// publishes their typed layers into the dependency container. Each onion is a container
// constructor: Init only registers, and BuildAllEngines forces the construction once every
// service an onion injects is registered.
package dynamicengines

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// InitDynamicEngines registers every resource this module serves.
func InitDynamicEngines() error {
	return stdErr.Join(
		registerPaymentMethodEngine(),
		registerPaymentProfileEngine(),
		registerTransactionEngine(),
		registerOrderEngine(),
		registerInvoiceEngine(),
		registerInvoiceLineEngine(),
	)
}

// SchemaNames lists the resources served by a composable onion, for route registration and the
// boot-time check that each was built.
func SchemaNames() []string {
	return []string{
		models.PaymentMethodSchemaName,
		models.PaymentProfileSchemaName,
		models.TransactionSchemaName,
		models.OrderSchemaName,
		models.InvoiceSchemaName,
		models.InvoiceLineSchemaName,
	}
}

// EngineSchemaNames is SchemaNames under the name the route tests are written against.
func EngineSchemaNames() []string {
	return SchemaNames()
}

// buildOnion builds one onion and installs it into the resource hub, through which the payment
// flow, the sweeps and the webhooks reach the repositories.
func buildOnion(
	impl *composable.DynamicResourceEngineOnionImpl, param composable.BuildParam,
) composable.DynamicResourceEngineOnion {
	onion := composable.MustBuild(impl, param)
	services.InstallResource(impl.SchemaName, onion.Repository(), onion.DomainService())
	return onion
}
