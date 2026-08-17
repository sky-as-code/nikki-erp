package restful

import (
	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	corelog "github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/dynamicengines"
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
	pathWebhookMomo               = "/webhooks/momo"
	pathWebhookMpos               = "/webhooks/mpos"
	pathWebhookVietQrToken        = "/webhooks/vietqr/token_generate"
	pathWebhookVietQrTransaction  = "/webhooks/vietqr/transaction_sync"
)

func InitRestfulHandlers() error {
	return initPaymentInvoiceV1()
}

func initPaymentInvoiceV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		cfg config.ConfigService,
		orders *services.OrderDomainService,
		registry *itGateway.Registry,
		logger corelog.LoggerService,
	) error {
		routeV1 := route.Group("/v1/paymentinvoice")

		registerEngineRoutes(routeV1)
		registerWebhookRoutes(routeV1, cfg, orders, registry, logger)

		return nil
	})
}

// registerWebhookRoutes exposes the gateway callbacks.
//
// They are the one part of this module that cannot be an engine action: they are called by MoMo,
// by NextPay and by the bank, none of which can present a user's authorization. Each therefore
// authenticates its own caller — by signature, by decryption, or by a bearer we issued — and is
// registered PublicUnauthorized so the framework does not look for a session that will never be
// there.
func registerWebhookRoutes(
	routeV1 *echo.Group,
	cfg config.ConfigService,
	orders *services.OrderDomainService,
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

	webhook := v1.NewWebhookRest(orders, registry, inbound, logger)

	routeV1.POST(pathWebhookMomo, webhook.MomoIpn, m.PublicUnauthorized)
	routeV1.POST(pathWebhookMpos, webhook.MposWebhook, m.PublicUnauthorized)
	routeV1.POST(pathWebhookVietQrToken, webhook.VietQrTokenGenerate, m.PublicUnauthorized)
	routeV1.POST(pathWebhookVietQrTransaction, webhook.VietQrTransactionSync, m.PublicUnauthorized)
}

// registerEngineRoutes exposes every Payment & Invoice resource engine over HTTP.
// A missing engine is skipped, so that a build which drops one still starts.
func registerEngineRoutes(routeV1 *echo.Group) {
	for _, schemaName := range dynamicengines.EngineSchemaNames() {
		engine, exists := dynamicresource.Registry().GetEngine(schemaName)
		if !exists {
			continue
		}
		engine.RestApi().RegisterRoutes(routeV1, m.SmokeAuthz())
	}
}
