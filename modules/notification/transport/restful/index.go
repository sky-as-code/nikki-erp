package restful

import (
	stdErr "errors"

	"net/http"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/notification/transport/restful/v1"
)

// Notification exposes a deliberately narrow HTTP surface.
//
// There is NO route that raises a notification: sending is something a module does on a user's
// behalf over the command bus or a direct call, and an HTTP route for it would let any caller write
// into any inbox (BR 29). The resources themselves are read-only over REST for the same reason —
// a notification is written by the module that raised it, never by the person receiving it.
func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewNotificationRest,
		v1.NewRecipientRest,
		v1.NewDeliveryRest,
		v1.NewInboxRest,
		v1.NewStreamRest,
	)
	if err != nil {
		return err
	}

	return deps.Invoke(func(route *echo.Group) error {
		routeV1 := route.Group("/v1/notification")
		return stdErr.Join(
			initResourcesV1(routeV1),
			initInboxV1(routeV1),
			initStreamV1(routeV1),
		)
	})
}

// initResourcesV1 registers the read half of the CRUD table for each resource.
//
// Create, update and delete are withheld on purpose: every write to these tables goes through
// SendNotification or MarkRead, which enforce the rules that a raw CRUD write would bypass.
func initResourcesV1(route *echo.Group) error {
	return deps.Invoke(func(
		notificationRest *v1.NotificationRest,
		recipientRest *v1.RecipientRest,
		deliveryRest *v1.DeliveryRest,
	) error {
		readOnly := []composable.CrudAction{
			composable.CrudActionGetById,
			composable.CrudActionSearch,
			composable.CrudActionExists,
			composable.CrudActionGetSchema,
		}

		return stdErr.Join(
			composable.NewRestEngine(models.NotificationSchemaName, notificationRest).
				AddCrudRoutes(readOnly...).
				RegisterRoutes(route),
			composable.NewRestEngine(models.RecipientSchemaName, recipientRest).
				AddCrudRoutes(readOnly...).
				RegisterRoutes(route),
			composable.NewRestEngine(models.DeliverySchemaName, deliveryRest).
				AddCrudRoutes(readOnly...).
				RegisterRoutes(route),
		)
	})
}

// initInboxV1 registers what a signed-in person calls about their own notifications.
//
// These are hand-registered rather than AddRoute actions because they are not addressed by a
// resource id: "my unread count" names no record. Each therefore carries SmokeAuthz explicitly,
// which a hand-registered route does not inherit.
func initInboxV1(route *echo.Group) error {
	return deps.Invoke(func(inboxRest *v1.InboxRest) {
		route.GET("/notifications", wrap(inboxRest.GetInbox), m.SmokeAuthz())
		route.GET("/notifications/unread-count", wrap(inboxRest.GetUnreadCount), m.SmokeAuthz())
		route.POST("/notifications/mark-read", wrap(inboxRest.MarkRead), m.SmokeAuthz())

		// After the literal paths above: ":id" would otherwise swallow "unread-count".
		route.POST("/notifications/:id/read", wrap(inboxRest.MarkOneRead), m.SmokeAuthz())
	})
}

// initStreamV1 registers the NDJSON stream.
//
// Hand-registered because the resource engine answers one JSON document per request and this
// answers a response that never ends — the documented exception for an unusual HTTP surface.
func initStreamV1(route *echo.Group) error {
	return deps.Invoke(func(streamRest *v1.StreamRest) {
		route.GET("/notifications/stream", wrap(streamRest.Stream), m.SmokeAuthz())
	})
}

// wrap adapts a composable handler to an echo one.
//
// The inbox handlers take the payload map the resource engine would have bound, so that they can be
// written the same way as every other custom action in the codebase. These routes bind nothing
// themselves: the filters are query parameters and the bodies are read by the handler.
func wrap(handler composable.HandlerFn) echo.HandlerFunc {
	return func(echoCtx *echo.Context) error {
		payload := map[string]any{}
		if echoCtx.Request().Method == http.MethodPost {
			// A body is optional on these routes; an unreadable one is left empty and the handler
			// reports the missing field rather than a parse error.
			_ = echoCtx.Bind(&payload)
		}
		return handler(echoCtx, payload)
	}
}
