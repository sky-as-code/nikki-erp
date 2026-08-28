// Package external binds Sales' local ports to the services other modules publish.
//
// This is the ONLY package in Sales that may import another module. Everything else depends on
// the interfaces in interfaces/external, so splitting a module into its own process changes this
// file and nothing else — the bindings become REST or CQRS clients and every caller is unaffected.
package external

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	itAccCurrency "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/currency"
	itTax "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/tax"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/pubsub"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
	itMethod "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
	"github.com/sky-as-code/nikki-erp/modules/sales/dynamicengines"
	salesMessage "github.com/sky-as-code/nikki-erp/modules/sales/infra/external/message"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// InitExternal binds every port Sales consumes, and registers what Sales offers back.
//
// The remaining ports — UoM and currency onto essential, the party read onto contacts — arrive with
// the tasks that first need them, rather than being guessed now.
//
// This runs before the engines are created, because a derived service resolves its ports at
// construction time.
func InitExternal() error {
	// Sales holds a uom_id on every order line, so Essential's BR-UOM-ESS-020 guard needs an answer
	// from this module before it may allow a unit to be edited. Registered here rather than lazily,
	// because a probe that arrives after the first edit request is a probe that did not run.
	RegisterUomUsageProbe()

	if err := deps.Register(
		// Registration and reading are two ports rather than one, so that the code declaring what
		// may be configured cannot also read another owner's values, and vice versa.
		//
		// The registry hangs off the TENANT service although Sales registers an org-level schema.
		// That is not a mismatch: RegisterSchema takes the level as a command field, and it is
		// carried on the tenant contract because declaring what may be configured is a tenant-wide
		// act whatever level the values then live at. iam binds it the same way.
		func(settings itSettings.TenantSettingsAppService) itExt.SettingsRegistrationExtService {
			return settings
		},
		func(settings itSettings.EffectiveSettingsAppService) itExt.EffectiveSettingsExtService {
			return settings
		},
		func(basis itProduct.ProductPricingBasisService) itExt.ProductPricingBasisExtService {
			// A direct hand-over: Inventory's port is already narrowed to the pricing inputs, so
			// there is nothing here to narrow further. Bound separately from the sellability port
			// above so that neither grants what the other does.
			return basis
		},
		func(orgCurrency itAccCurrency.OrgCurrencyService) itExt.OrgCurrencyExtService {
			// Accounting rather than Essential: Essential owns the currency catalogue, Accounting
			// owns which one this organization's books are kept in. A direct hand-over — the port
			// declares the single method the upstream service has.
			return orgCurrency
		},
		func(tax itTax.TaxCalculationAppService) itExt.TaxCalculationExtService {
			// Accounting owns every tax decision: which taxes apply, at what rate, whether the price
			// includes them, and how the result rounds. Sales supplies the commercial base and stores
			// the snapshot it gets back. A direct hand-over, like the payment-method port below.
			return tax
		},
		func(publisher pubsub.Publisher) itMessage.IntegrationEventPublisher {
			// An adapter rather than a hand-over: the broker takes bytes on a topic, and deciding
			// the encoding and the topic is exactly what this layer is for. Nothing above knows
			// integration events are JSON.
			return salesMessage.NewPublisher(publisher)
		},
		func(variants itProduct.ProductVariantDomainService) itExt.ProductVariantExtService {
			// An ADAPTER, not a hand-over: Inventory publishes a general reader and Sales needs one
			// narrow judgement. Passing the reader through would let a caller read a variant's
			// price, and a price read from Inventory is one somebody eventually sells at - which is
			// what Sales' own pricelists exist to prevent (BR 16).
			return &productVariantAdapter{variants: variants}
		},
		func(
			transfers itStock.StockTransferMovementService,
			settings itSettings.EffectiveSettingsAppService,
		) itExt.FulfillmentExtService {
			// An ADAPTER, and necessarily so: Sales sends one commercial intent, while Inventory
			// needs a document sequenced through create, confirm, reserve and validate. Binding
			// directly would put Inventory's lifecycle into Sales, where it would go stale the
			// first time Inventory changed it (SALES-049).
			return &fulfillmentAdapter{
				transfers:      transfers,
				operationTypes: &settingsOperationTypes{settings: settings},
			}
		},
		func(methods itMethod.PaymentMethodAppService) itExt.PaymentMethodExtService {
			// The upstream service has exactly the two methods the port declares, so this is a
			// direct hand-over rather than an adapter. It becomes a client when this application is
			// split into separate processes, and no caller changes.
			return methods
		},
	); err != nil {
		return err
	}

	// The channel's payment actions reach the port through a package variable rather than the
	// container, because an action callback is handed only its own engine. Resolving it once here is
	// what connects the two — the same shape paymentinvoice uses for its order service.
	return deps.Invoke(func(
		methods itExt.PaymentMethodExtService,
		tax itExt.TaxCalculationExtService,
		settings itExt.EffectiveSettingsExtService,
		dLock lock.DistributedLock,
		products itExt.ProductVariantExtService,
		fulfillment itExt.FulfillmentExtService,
		basis itExt.ProductPricingBasisExtService,
	) error {
		dynamicengines.SetPaymentMethodPort(methods)
		// Reprice needs tax and settings; confirm and cancel additionally need the lock, because
		// neither is a single-row update and the etag cannot guard them (D-30). The lock comes from
		// core's own DI, so no port of Sales' own is involved.
		dynamicengines.SetPricingPorts(tax, settings, dLock, products, fulfillment, basis)
		return nil
	})
}
