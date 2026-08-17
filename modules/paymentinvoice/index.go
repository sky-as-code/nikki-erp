// Package paymentinvoice takes payments through the payment gateways the business has contracts
// with, and issues the invoices that account for them.
//
// It supersedes a standalone NestJS service that spoke to the same three gateways. Two things
// carried over from it unchanged on purpose, because they are contracts with parties outside this
// codebase: the wire shape each gateway expects, and the payload the ordering system is notified
// with once a payment settles.
package paymentinvoice

import (
	stdErr "errors"
	"time"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	httpclientclient "github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	modconstants "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/app"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/transport/restful"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader.
//
// It is typed DynamicModule rather than InCodeModule so that dropping RegisterModels fails the
// build. Under the wider interface the method is found by a type assertion instead, and a module
// that has lost it still compiles, still loads, and silently registers no schemas at all.
var ModuleSingleton modules.DynamicModule = &PaymentInvoiceModule{}

type PaymentInvoiceModule struct{}

func (*PaymentInvoiceModule) LabelKey() string {
	return "paymentinvoice.moduleLabel"
}

func (*PaymentInvoiceModule) Name() string {
	return modconstants.PaymentInvoiceModuleName
}

func (*PaymentInvoiceModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
	}
}

func (*PaymentInvoiceModule) IsInternal() bool {
	return false
}

func (*PaymentInvoiceModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements DynamicModule.
//
// The steps are ordered: the gateway registry must exist before the order service that selects
// from it, that service before the engines whose actions delegate to it, and the engines before
// the REST layer that registers their routes.
func (*PaymentInvoiceModule) Init() error {
	if err := initOrderService(); err != nil {
		return err
	}
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	return restful.InitRestfulHandlers()
}

// initOrderService builds the gateway registry from configuration and installs the order service
// the engine actions delegate to.
//
// The registry is built once and shared: each adapter holds a connection pool and, for VietQR, a
// cached bearer token, so one instance per request would re-authenticate on every payment.
func initOrderService() error {
	return deps.Invoke(func(
		cfg config.ConfigService,
		httpClient *httpclientclient.HttpClient,
		logger logging.LoggerService,
	) error {
		registry, err := gateway.BuildRegistry(cfg, httpClient, logger)
		if err != nil {
			return err
		}

		dynamicengines.SetOrderService(services.NewOrderDomainService(registry))
		dynamicengines.SetInvoiceService(services.NewInvoiceDomainService())
		return nil
	})
}

// RegisterModels implements DynamicModule.
//
// Schemas are registered referenced-before-referencing, because an edge is resolved against the
// schema registry at registration time: the payment method is pointed at by both the order and the
// transaction, the transaction points at the order, and the invoice line at the invoice.
//
// The edges onto essential_currency resolve because Essential is named in Deps() and every
// module's RegisterModels runs in dependency order, before any module's Init().
func (*PaymentInvoiceModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.PaymentMethodSchemaBuilder()),
		dmodel.RegisterSchemaB(models.OrderSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TransactionSchemaBuilder()),
		dmodel.RegisterSchemaB(models.InvoiceSchemaBuilder()),
		dmodel.RegisterSchemaB(models.InvoiceLineSchemaBuilder()),
	)
}

// OnAppStarted implements InCodeModuleAppStarted.
//
// The three sweeps are registered here rather than in Init because they must not start until the
// application is serving: the watchdog writes order state, and running it against a half-built
// container would fail on the first tick.
//
// Their tuning is read once, at registration, so a sweep does not re-read configuration on every
// tick. A deployment that changes an interval restarts to pick it up, which is the same contract
// as every other value in this file.
func (*PaymentInvoiceModule) OnAppStarted() error {
	return deps.Invoke(func(
		cfg config.ConfigService,
		orders *services.OrderDomainService,
		cronRegistry job.CronjobRegistry,
		logger logging.LoggerService,
	) error {
		syncClient := app.NewResultSyncClient(
			time.Duration(cfg.GetInt(modconstants.SyncTimeoutSecs, defaultSyncTimeoutSecs))*time.Second,
			cfg.GetInt(modconstants.SyncMaxRetries, defaultSyncMaxRetries),
		)

		manager := app.NewJobsManager(orders, syncClient, app.JobsConfig{
			ExpireAfter: time.Duration(
				cfg.GetInt(modconstants.OrderExpireAfterMins, defaultExpireAfterMins)) * time.Minute,
			CleanAfter: time.Duration(
				cfg.GetInt(modconstants.OrderCleanAfterHours, defaultCleanAfterHours)) * time.Hour,
		}, logger)

		return manager.RegisterJobs(cronRegistry)
	})
}

// Fallbacks for the sweep tuning, used when a deployment's configuration omits a key.
//
// They match config.default.yaml. A zero would be actively harmful — an expiry of zero minutes
// expires every order the moment it is created — so the sweeps must never inherit an unset value.
const (
	defaultExpireAfterMins = 15
	defaultCleanAfterHours = 24
	defaultSyncTimeoutSecs = 5
	defaultSyncMaxRetries  = 3
)
