// Package external binds Sales' local ports to the services other modules publish. It is the only
// package in Sales that may import another module; everything else depends on interfaces/external,
// so splitting a module into its own process changes this file and nothing else.
package external

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itAccCurrency "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/currency"
	itTax "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
	itParty "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/party"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
	itJob "github.com/sky-as-code/nikki-erp/modules/jobscheduler/interfaces/job"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"
	itMethod "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
	salesInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/infra/external/invoicing"
	salesMessage "github.com/sky-as-code/nikki-erp/modules/sales/infra/external/message"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// InitExternal binds every port Sales consumes and registers what Sales offers back. It runs before
// the engines are created, because a derived service resolves its ports at construction time.
func InitExternal() error {
	if err := deps.Register(
		// Registration and reading are two ports so the code declaring what may be configured
		// cannot also read another owner's values. The registry hangs off the TENANT service even
		// though Sales registers an org-level schema: RegisterSchema takes the level as a command
		// field, and declaring what may be configured is tenant-wide whatever level the values
		// live at.
		func(settings itSettings.TenantSettingsAppService) itExt.SettingsRegistrationExtService {
			return settings
		},
		func(settings itSettings.EffectiveSettingsAppService) itExt.EffectiveSettingsExtService {
			return settings
		},
		func(basis itProduct.ProductPricingBasisService) itExt.ProductPricingBasisExtService {
			// Direct hand-over: inventory's port is already narrowed to the pricing inputs. Bound
			// separately from the sellability port so neither grants what the other does.
			return basis
		},
		func(orgCurrency itAccCurrency.OrgCurrencyService) itExt.OrgCurrencyExtService {
			// Accounting rather than essential: essential owns the currency catalogue, accounting
			// owns which one this organization's books are kept in.
			return orgCurrency
		},
		func(tax itTax.TaxCalculationAppService) itExt.TaxCalculationExtService {
			// Accounting owns every tax decision — which taxes apply, at what rate, tax-inclusive or
			// not, and rounding. Sales supplies the commercial base and stores the snapshot.
			return tax
		},
		func(reservations itStock.WarehouseReservationService) itExt.WarehouseReservationExtService {
			// A hand-over: the port is Inventory's own contract, aliased on this side.
			return reservations
		},
		func(publisher pubsub.Publisher) itMessage.IntegrationEventPublisher {
			// An adapter, not a hand-over: the broker takes bytes on a topic, and nothing above this
			// layer knows integration events are JSON.
			return salesMessage.NewPublisher(publisher)
		},
		func(variants itProduct.ProductVariantDomainService) itExt.ProductVariantExtService {
			// An adapter, not a hand-over: inventory publishes a general reader and Sales needs one
			// narrow judgement. Passing the reader through would let a caller read a variant's
			// price, which somebody eventually sells at — what Sales' own pricelists prevent.
			return &productVariantAdapter{variants: variants}
		},
		func(
			transfers itStock.StockTransferMovementService,
			availability itStock.StockProductSummaryReader,
			settings itSettings.EffectiveSettingsAppService,
		) itExt.FulfillmentExtService {
			// An adapter necessarily: Sales sends one commercial intent, while inventory needs a
			// document sequenced through create, confirm, reserve and validate. Binding directly
			// would put inventory's lifecycle into Sales, where it would go stale.
			return newFulfillmentAdapter(transfers, availability, settings)
		},
		func(
			transfers itStock.StockTransferMovementService,
			availability itStock.StockProductSummaryReader,
			settings itSettings.EffectiveSettingsAppService,
		) itExt.FulfillmentReservationExtService {
			// The same adapter behind a second port. Two ports rather than one because the callers
			// differ: the confirm and cancel flows send commercial intents, while the fulfillment
			// services hold and move stock for a named target, and a module needing only one of
			// those should not have to depend on both.
			return newFulfillmentAdapter(transfers, availability, settings)
		},
		func(methods itMethod.PaymentMethodAppService) itExt.PaymentMethodExtService {
			// Direct hand-over: the upstream service has exactly the two methods the port declares.
			return methods
		},
		func(parties itParty.PartyAppService) itExt.PartyExtService {
			// An adapter, not a hand-over: the upstream answer is an OpResult carrying its own
			// violations, and Sales wants the violations alone so a refusal joins the ones its own
			// gates raise. The APPLICATION service, because assigning a party happens on a request
			// whose user's entitlements are the ones that should decide.
			return &partyAdapter{parties: parties}
		},
		func(jobs itJob.JobDomainService) itExt.SchedulerExtService {
			// An adapter, not a hand-over: the upstream command embeds a dynamic model, so passing it
			// through would put the scheduler's own field names and enum values into Sales' domain.
			// The DOMAIN service rather than the application one — registration happens at boot,
			// where there is no user whose entitlements could be asserted.
			return &schedulerAdapter{jobs: jobs}
		},
		func(invoices itInvoice.InvoiceDomainService) itInvoicing.InvoicingExtService {
			// An adapter necessarily: the port names nothing on the far side, so that pointing Sales
			// at a real e-invoice provider later is a change to one file. It also carries the two
			// conversions this boundary needs — the tax rate from a fraction to a percentage, and
			// the check that the issued document totals what the sale actually came to.
			return salesInvoicing.NewAdapter(invoices)
		},
		func(orders itOrder.OrderDomainService) itExt.PaymentOrderExtService {
			// An adapter, not a hand-over: the conventions this integration runs on — the source tag,
			// the metadata a settlement is matched back on, and passing no return_url because the
			// verdict arrives in-process — belong in one place, and paymentinvoice's order states are
			// mapped here so its state machine does not get duplicated inside Sales.
			//
			// The DOMAIN service, per the port's own note: authorization was established by the
			// request that started this, so there is no application-service counterpart to call.
			return &paymentOrderAdapter{orders: orders}
		},
	); err != nil {
		return err
	}
	// Every port Sales consumes is now injected into the onion that needs it, so nothing is pushed
	// into a package variable here any more: the composable engine hands a resource its
	// dependencies at construction, which is what the legacy action callbacks could not do.
	return nil
}
