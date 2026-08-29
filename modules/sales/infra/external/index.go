// Package external binds Sales' local ports to the services other modules publish. It is the only
// package in Sales that may import another module; everything else depends on interfaces/external,
// so splitting a module into its own process changes this file and nothing else.
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

// InitExternal binds every port Sales consumes and registers what Sales offers back. It runs before
// the engines are created, because a derived service resolves its ports at construction time.
func InitExternal() error {
	// Sales holds a uom_id on every order line, so essential's guard needs an answer from this
	// module before allowing a unit to be edited. Registered eagerly: a probe that arrives after
	// the first edit request did not run.
	RegisterUomUsageProbe()

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
			settings itSettings.EffectiveSettingsAppService,
		) itExt.FulfillmentExtService {
			// An adapter necessarily: Sales sends one commercial intent, while inventory needs a
			// document sequenced through create, confirm, reserve and validate. Binding directly
			// would put inventory's lifecycle into Sales, where it would go stale.
			return &fulfillmentAdapter{
				transfers:      transfers,
				operationTypes: &settingsOperationTypes{settings: settings},
			}
		},
		func(methods itMethod.PaymentMethodAppService) itExt.PaymentMethodExtService {
			// Direct hand-over: the upstream service has exactly the two methods the port declares.
			return methods
		},
	); err != nil {
		return err
	}

	// The channel's payment actions reach the port through a package variable rather than the
	// container, because an action callback is handed only its own engine.
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
		// neither is a single-row update and the etag cannot guard them.
		dynamicengines.SetPricingPorts(tax, settings, dLock, products, fulfillment, basis)
		return nil
	})
}
