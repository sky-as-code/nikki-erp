// Package sales owns the commercial transaction: the sales order and its lines, pricing and
// promotion, billing, payment allocation, return and refund — per
// docs/requirements/sales/00-business-requirement-md and the channel/sales-point change request in
// docs/requirements/sales/01-sales-channel.md.
//
// It is deliberately channel-agnostic. A vending kiosk, a future POS and a future storefront all
// create the same sales_order through the same pricing engine; nothing in this module branches on
// which one is calling. What differs between them is data — a sales channel row and its sales
// points — not code.
//
// It owns none of Product, UoM, Warehouse, Stock, Payment Method, VAT Invoice or Accounting. It
// holds ids into those modules and reads them through ports, and it sends business intent rather
// than instructions: a fulfilment request rather than a stock movement, a fiscal document request
// rather than a call to a tax provider.
package sales

import (
	"context"
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/sales/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/sales/infra/external"
	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
	"github.com/sky-as-code/nikki-erp/modules/sales/transport/restful"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader.
//
// It is typed DynamicModule rather than InCodeModule so that dropping RegisterModels fails the
// build. Under the wider interface the method is found by a type assertion instead, and a module
// that has lost it still compiles, still loads, and silently registers no schemas at all.
var ModuleSingleton modules.DynamicModule = &SalesModule{}

// OnAppStarted is found by a type assertion rather than by the interface above, so a rename or a
// changed signature would not fail the build — the module would simply load with its settings
// schema never registered, and nothing would say so. This assertion is what turns that into a
// compile error.
var _ modules.InCodeModuleAppStarted = &SalesModule{}

type SalesModule struct{}

func (*SalesModule) LabelKey() string {
	return "sales.moduleLabel"
}

func (*SalesModule) Name() string {
	return modconstants.SalesModuleName
}

// Deps names every module Sales reads through a port.
//
// dynamicresource hosts the resource engines. settings stores the commercial policy an organization
// configures — the rounding scale, the return window, whether cash change is possible. inventory
// supplies the product variant a line sells
// and receives the fulfilment requests an order raises. essential supplies UoM conversion and
// currency. contacts supplies the party a customer_reference points at. paymentinvoice owns
// payment method master data and the invoice capability a fiscal request is delegated to.
//
// essential and core are injected automatically by buildDependencyGraph; essential is named anyway
// because Sales consumes it directly, and a reader should not have to know the implicit rule to
// see that.
func (*SalesModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
		"settings",
		"inventory",
		"contacts",
		"paymentinvoice",

		// Sales binds accounting's tax port in infra/external, and binds it EAGERLY - the reprice
		// action resolves it at Init rather than at first request. So accounting must have run its
		// own Init and registered the service before this module starts, which is what naming it
		// here guarantees. Without it the loader is free to start Sales first, and does.
		"accounting",
	}
}

func (*SalesModule) IsInternal() bool {
	return false
}

func (*SalesModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements DynamicModule.
//
// The order is load-bearing three times over, and is fixed now rather than when the first resource
// arrives: the external ports bind first, because a derived service resolves its ports when it is
// constructed; the engines are created before the derived services, because a derived service
// wraps the engine's own; and the REST layer is registered last, because it registers the engines'
// routes and so cannot run before they exist.
func (*SalesModule) Init() error {
	if err := external.InitExternal(); err != nil {
		return err
	}
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	if err := dynamicengines.InitDomainServices(); err != nil {
		return err
	}
	if err := app.InitApplicationServices(); err != nil {
		return err
	}
	// The payment mapping gate is one of Sales' own application services, so it can only be
	// resolved once the step above has registered it - not with the external ports, which bind
	// before anything in this module exists.
	if err := deps.Invoke(func(channels itChannel.ChannelPaymentAppService) error {
		dynamicengines.SetChannelPaymentService(channels)
		return nil
	}); err != nil {
		return err
	}
	return restful.InitRestfulHandlers()
}

// OnAppStarted implements InCodeModuleAppStarted.
//
// The settings schema is registered here rather than in Init() because peer module init order is
// nondeterministic: Init() cannot assume the settings module has built its engines yet, while
// OnAppStarted runs after every module has initialized. Registration is idempotent, so it runs
// unconditionally.
// The outbox sweep registers here for a second reason on top of that one: a sweep that writes rows
// must not tick against a half-built container, which is what registering it in Init() would risk.
func (*SalesModule) OnAppStarted() error {
	return deps.Invoke(func(
		settingsSvc itExt.SettingsRegistrationExtService,
		publisher itMessage.IntegrationEventPublisher,
		effective itExt.EffectiveSettingsExtService,
		cronjobs job.CronjobRegistry,
		logger logging.LoggerService,
	) error {
		if err := registerOrgSettings(
			corectx.NewRequestContext(context.Background()), settingsSvc); err != nil {
			return err
		}
		if err := app.NewOutboxJobs(publisher, logger).RegisterJobs(cronjobs); err != nil {
			return err
		}
		return app.NewExpiryJobs(effective, logger).RegisterJobs(cronjobs)
	})
}

// RegisterModels implements DynamicModule.
//
// Registration order is load-bearing: an edge is resolved against the schema registry at
// registration time, so a referenced schema must be registered before the one pointing at it —
// the sales channel before its sales points, the order before its lines.
//
// The schemas are listed here rather than scattered across the packages that own them, so that
// the order is visible in a single place and a missing registration is a gap in a list rather than
// an absence nobody can see.
func (*SalesModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.SalesChannelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPointSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesChannelPaymentRelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderLineComponentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderAdjustmentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderEventSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPricelistSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPricelistItemSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesComboSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesComboComponentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPromotionProgramSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPromotionConditionGroupSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPromotionConditionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPromotionConditionTargetSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPromotionRewardSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPromotionCompatibilitySchemaBuilder()),
		// Vouchers register after the promotion program they point at: an edges_to must resolve
		// against a schema that is already registered.
		dmodel.RegisterSchemaB(models.SalesVoucherCodeSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesVoucherRedemptionSchemaBuilder()),
		// Bills register after the order and its lines, which their edges point at.
		dmodel.RegisterSchemaB(models.SalesBillSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesBillLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesBillRelationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPaymentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesFulfillmentRequestSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesFulfillmentRequestLineSchemaBuilder()),
		// The fiscal request registers after the bill it points at.
		dmodel.RegisterSchemaB(models.SalesFiscalRequestSchemaBuilder()),

		// Manual overrides register after the order they discount.
		dmodel.RegisterSchemaB(models.SalesManualDiscountSchemaBuilder()),

		// Quotations register after the channel they point at.
		dmodel.RegisterSchemaB(models.SalesQuotationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesQuotationLineSchemaBuilder()),

		// The outbox has no edges at all — an integration event must outlive the record it
		// describes, and a cascade would delete exactly the history a consumer is replaying — so
		// its position in this list is free.
		dmodel.RegisterSchemaB(models.SalesIntegrationOutboxSchemaBuilder()),
	)
}
