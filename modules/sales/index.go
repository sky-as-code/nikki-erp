// Package sales owns the commercial transaction: sales order and lines, pricing and promotion,
// billing, payment allocation, return and refund. It is channel-agnostic — kiosk, POS and
// storefront differ only by data (a sales channel row and its sales points), never by branching in
// this module. It owns no Product, UoM, Warehouse, Stock, Payment Method, VAT Invoice or Accounting
// data; it holds ids into those modules, reads them through ports, and sends business intent (a
// fulfilment request, not a stock movement) rather than instructions.
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
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/sales/dynamicengines"
	eventhandlers "github.com/sky-as-code/nikki-erp/modules/sales/event_handlers"
	"github.com/sky-as-code/nikki-erp/modules/sales/infra/external"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
	itMessage "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/message"
	salescqrs "github.com/sky-as-code/nikki-erp/modules/sales/transport/cqrs"
	eventtransport "github.com/sky-as-code/nikki-erp/modules/sales/transport/event"
	"github.com/sky-as-code/nikki-erp/modules/sales/transport/restful"
)

// ModuleSingleton is the symbol the plugin loader looks up. It is typed DynamicModule rather than
// InCodeModule so that dropping RegisterModels fails the build instead of silently registering no
// schemas.
var ModuleSingleton modules.DynamicModule = &SalesModule{}

// OnAppStarted is found by a runtime type assertion, so this assertion is what turns a rename or
// signature change into a compile error rather than a silently unregistered settings schema.
var _ modules.InCodeModuleAppStarted = &SalesModule{}

type SalesModule struct{}

func (*SalesModule) LabelKey() string {
	return "sales.moduleLabel"
}

func (*SalesModule) Name() string {
	return modconstants.SalesModuleName
}

func (*SalesModule) ModelPrefix() string {
	return modconstants.SalesModuleName
}

// Deps names every module Sales reads through a port. essential and core are injected automatically
// by buildDependencyGraph; essential is named anyway because Sales consumes it directly.
func (*SalesModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
		"settings",
		"inventory",
		"contacts",
		"paymentinvoice",

		// The electronic-invoice job is registered with the scheduler at boot, so the scheduler must
		// have built its engines before Sales starts.
		"jobscheduler",

		// The reprice action resolves accounting's tax port eagerly at Init, so accounting must
		// have registered its service before Sales starts. Without naming it here the loader
		// starts Sales first.
		"accounting",
	}
}

func (*SalesModule) IsInternal() bool {
	return false
}

func (*SalesModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements DynamicModule. The order is load-bearing: external ports bind first because a
// derived service resolves its ports at construction; engines precede derived services because a
// derived service wraps the engine's own; REST registers last because it registers engine routes.
func (*SalesModule) Init() error {
	if err := external.InitExternal(); err != nil {
		return err
	}
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	if err := app.InitApplicationServices(); err != nil {
		return err
	}
	// After the application services: the channel and bill onions inject ChannelPaymentAppService,
	// which the step above registers. Forcing the build any earlier fails with a missing type.
	if err := dynamicengines.BuildAllEngines(); err != nil {
		return err
	}
	// Event handlers before subscribers: a subscriber resolves its handler registry at construction,
	// so wiring them the other way round leaves it with nothing to dispatch to.
	if err := eventhandlers.InitHandlers(); err != nil {
		return err
	}
	if err := eventtransport.InitEventSubscribers(); err != nil {
		return err
	}
	// Before OnAppStarted registers the job that dispatches to it: the scheduler refuses to register
	// a job whose command name is not a known request type.
	if err := salescqrs.InitCqrsHandlers(); err != nil {
		return err
	}
	if err := restful.InitRestfulHandlers(); err != nil {
		return err
	}
	// After the routes, because registering them is what forces dig to build the onions. A resource
	// declared but never built would answer every resource-hub lookup with "not installed", and the
	// first caller to notice would be a cron sweep hours later rather than this boot.
	return assertEveryResourceInstalled()
}

// assertEveryResourceInstalled fails the boot when an onion was declared but never built.
func assertEveryResourceInstalled() error {
	installed := map[string]bool{}
	for _, name := range services.InstalledResourceNames() {
		installed[name] = true
	}
	var missing []error
	for _, name := range dynamicengines.SchemaNames() {
		if !installed[name] {
			missing = append(missing,
				stdErr.New("the '"+name+"' sales resource was declared but never built"))
		}
	}
	return stdErr.Join(missing...)
}

// OnAppStarted implements InCodeModuleAppStarted. The settings schema registers here, not in Init,
// because peer module init order is nondeterministic and the settings module may not have built its
// engines yet; registration is idempotent. The outbox sweep registers here too so it never ticks
// against a half-built container.
func (*SalesModule) OnAppStarted() error {
	return deps.Invoke(func(
		settingsSvc itExt.SettingsRegistrationExtService,
		publisher itMessage.IntegrationEventPublisher,
		effective itExt.EffectiveSettingsExtService,
		orders itExt.PaymentOrderExtService,
		invoicing itInvoicing.InvoicingExtService,
		scheduler itExt.SchedulerExtService,
		reservations itExt.FulfillmentReservationExtService,
		cronjobs job.CronjobRegistry,
		logger logging.LoggerService,
	) error {
		if err := registerOrgSettings(
			corectx.NewRequestContext(context.Background()), settingsSvc); err != nil {
			return err
		}
		outbox := app.NewOutboxJobs(publisher, logger)
		if err := outbox.RegisterJobs(cronjobs); err != nil {
			return err
		}
		// Lets a write path ask for its event to go out at once rather than on the next tick. The
		// sweep registered above is still what guarantees delivery; without this installation the
		// module works exactly as before, only slower.
		services.SetOutboxDrain(outbox.DrainNow)
		// The backstop for a settlement announcement that was lost: the event bus acknowledges
		// before it dispatches, so without this a paid bill could stay open forever.
		if err := app.NewPaymentReconJobs(orders, invoicing, logger).RegisterJobs(cronjobs); err != nil {
			return err
		}
		if err := app.NewExpiryJobs(effective, reservations, logger).RegisterJobs(cronjobs); err != nil {
			return err
		}
		// Registered with the scheduler rather than the in-process cron, unlike the sweeps above:
		// issuing produces a legal document through a third party, so a run that failed has to be
		// visibly failed and retried on a policy someone can see, which a cron loop does not offer.
		return registerEinvoiceJob(
			corectx.NewRequestContext(context.Background()), scheduler, logger)
	})
}

// RegisterModels implements DynamicModule. Order is load-bearing: edges resolve against the schema
// registry at registration time, so a referenced schema must be registered before the one pointing
// at it. All schemas are listed here so that order is visible in one place.
func (*SalesModule) RegisterModels() error {
	return stdErr.Join(
		// The fulfillment method registers before the channel and point, whose default_*_id
		// columns and allowed-method mapping point at it.
		dmodel.RegisterSchemaB(models.SalesFulfillmentMethodSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesChannelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPointSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesChannelPaymentRelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesChannelFulfillmentMethodSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderLineAllocationSchemaBuilder()),
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
		// Vouchers register after the promotion program they point at.
		dmodel.RegisterSchemaB(models.SalesVoucherCodeSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesVoucherRedemptionSchemaBuilder()),
		// Bills register after the order and its lines, which their edges point at.
		dmodel.RegisterSchemaB(models.SalesBillSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesBillLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesBillRelationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPaymentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesFulfillmentRequestSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesFulfillmentRequestLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderFulfillmentSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesOrderFulfillmentItemSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesFulfillmentAttemptSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesFulfillmentAttemptItemSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesFulfillmentTargetChangeSchemaBuilder()),
		// The fiscal request registers after the bill it points at.
		dmodel.RegisterSchemaB(models.SalesFiscalRequestSchemaBuilder()),

		// Billing instructions register after the order they bill, and their attempts after them.
		dmodel.RegisterSchemaB(models.SalesBillingInstructionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesBillingIssuanceAttemptSchemaBuilder()),

		// Manual overrides register after the order they discount.
		dmodel.RegisterSchemaB(models.SalesManualDiscountSchemaBuilder()),

		// Quotations register after the channel they point at.
		dmodel.RegisterSchemaB(models.SalesQuotationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesQuotationLineSchemaBuilder()),

		// Returns register after the order and order line they reference; refund legs after the
		// payment they give back.
		dmodel.RegisterSchemaB(models.SalesReturnSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesReturnLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesRefundPaymentSchemaBuilder()),

		// The outbox has no edges: an integration event must outlive the record it describes, so a
		// cascade would delete history a consumer is replaying. Its position here is free.
		dmodel.RegisterSchemaB(models.SalesIntegrationOutboxSchemaBuilder()),
	)
}
