package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	corelog "github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/app"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway/vietqr"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
	v1 "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/transport/restful/v1"
)

// Webhook route paths.
//
// These are registered with NextPay and with the bank, so a rename here is an outage that shows up
// as callbacks arriving at a 404 — and a payment callback that 404s is a customer who paid for
// goods the system never released. They are named rather than inlined so the coupling is visible.
const (
	pathWebhookMomo              = "/webhooks/momo"
	pathWebhookMpos              = "/webhooks/mpos"
	pathWebhookVietQrToken       = "/webhooks/vietqr/token_generate"
	pathWebhookVietQrTransaction = "/webhooks/vietqr/transaction_sync"
)

const routeV1Prefix = "/v1/paymentinvoice"

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewPaymentMethodRest,
		v1.NewPaymentProfileRest,
		v1.NewTransactionRest,
		v1.NewOrderRest,
		v1.NewInvoiceRest,
		v1.NewInvoiceLineRest,
	)
	return stdErr.Join(err, initResourcesV1(), initWebhooksV1())
}

// legacyCrudActions is the built-in verb set the legacy engine served: every CRUD action except
// bulk create and import, which these resources never exposed.
func legacyCrudActions() []composable.CrudAction {
	return []composable.CrudAction{
		composable.CrudActionCreate, composable.CrudActionUpdate, composable.CrudActionDelete,
		composable.CrudActionSetArchived, composable.CrudActionGetById, composable.CrudActionSearch,
		composable.CrudActionExists, composable.CrudActionGetSchema, composable.CrudActionComputeField,
	}
}

func initResourcesV1() error {
	return deps.Invoke(func(route *echo.Group) error {
		routeV1 := route.Group(routeV1Prefix)
		return stdErr.Join(
			initPaymentMethodV1(routeV1),
			initPaymentProfileV1(routeV1),
			initTransactionV1(routeV1),
			initOrderV1(routeV1),
			initInvoiceV1(routeV1),
			initInvoiceLineV1(routeV1),
		)
	})
}

func initPaymentMethodV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.PaymentMethodRest) error {
		return crudRoutes(models.PaymentMethodSchemaName, rest).RegisterRoutes(route)
	})
}

// The payment profile's create and update bind "config", which is not a schema field; the
// overrides live on PaymentProfileRest, so the route table is the ordinary one.
func initPaymentProfileV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.PaymentProfileRest) error {
		return crudRoutes(models.PaymentProfileSchemaName, rest).RegisterRoutes(route)
	})
}

func initTransactionV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.TransactionRest) error {
		return crudRoutes(models.TransactionSchemaName, rest).RegisterRoutes(route)
	})
}

func initOrderV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.OrderRest) error {
		return orderRoutes(rest).RegisterRoutes(route)
	})
}

func initInvoiceV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.InvoiceRest) error {
		return invoiceRoutes(rest).RegisterRoutes(route)
	})
}

func initInvoiceLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.InvoiceLineRest) error {
		return crudRoutes(models.InvoiceLineSchemaName, rest).RegisterRoutes(route)
	})
}

func crudRoutes(schemaName string, rest composable.CrudRestHandlers) composable.RestEngine {
	return composable.NewRestEngine(schemaName, rest).AddCrudRoutes(legacyCrudActions()...)
}

// orderRoutes keeps the action paths the vending machines and the ordering system call.
func orderRoutes(rest *v1.OrderRest) composable.RestEngine {
	return crudRoutes(models.OrderSchemaName, rest).
		AddRoute(composable.RouteDefinition{
			Path: constants.ActionCreatePayment, ActionType: composable.ActionTypeGeneric, HandlerFn: rest.CreatePayment,
		}).
		AddRoute(composable.RouteDefinition{
			Path: constants.ActionRefund, ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Refund,
		}).
		AddRoute(composable.RouteDefinition{
			Path:       constants.ActionRemovePosOrders + "/:pos_id",
			ActionType: composable.ActionTypeGeneric, HandlerFn: rest.RemovePosOrders,
		})
}

func invoiceRoutes(rest *v1.InvoiceRest) composable.RestEngine {
	return crudRoutes(models.InvoiceSchemaName, rest).
		AddRoute(composable.RouteDefinition{
			Path: ":id/" + constants.ActionIssue, ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Issue,
		})
}

func initWebhooksV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		cfg config.ConfigService,
		orders *services.OrderDomainService,
		notifier *app.ResultNotifier,
		registry *itGateway.Registry,
		logger corelog.LoggerService,
	) error {
		registerWebhookRoutes(route.Group(routeV1Prefix), cfg, orders, notifier, registry, logger)
		return nil
	})
}

// registerWebhookRoutes exposes the gateway callbacks.
//
// They cannot be resource actions: they are called by MoMo, by NextPay and by the bank, none of
// which can present a user's authorization. Each therefore authenticates its own caller — by
// signature, by decryption, or by a bearer we issued — and is registered PublicUnauthorized so the
// framework does not look for a session that will never be there.
func registerWebhookRoutes(
	routeV1 *echo.Group,
	cfg config.ConfigService,
	orders *services.OrderDomainService,
	notifier *app.ResultNotifier,
	registry *itGateway.Registry,
	logger corelog.LoggerService,
) {
	// The inbound credentials come from configuration and are distinct from the pair the adapter
	// presents when it calls VietQR. An unset pair yields an authenticator that refuses every
	// request, which is why this is warned about at boot rather than discovered at the first
	// callback.
	inbound := vietqr.NewInboundAuth(
		cfg.GetStr(constants.VietQrInboundUsername),
		cfg.GetStr(constants.VietQrInboundPassword),
		cfg.GetStr(constants.VietQrInboundJwtSecret),
	)
	if !inbound.IsConfigured() {
		logger.Warnf("paymentinvoice: VIETQR.INBOUND_* is unset; the VietQR callbacks will refuse every request")
	}

	webhook := v1.NewWebhookRest(orders, notifier, registry, inbound, logger)

	routeV1.POST(pathWebhookMomo, webhook.MomoIpn, m.PublicUnauthorized)
	routeV1.POST(pathWebhookMpos, webhook.MposWebhook, m.PublicUnauthorized)
	routeV1.POST(pathWebhookVietQrToken, webhook.VietQrTokenGenerate, m.PublicUnauthorized)
	routeV1.POST(pathWebhookVietQrTransaction, webhook.VietQrTransactionSync, m.PublicUnauthorized)
}
