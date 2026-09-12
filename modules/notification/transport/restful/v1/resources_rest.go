package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The read-only CRUD surfaces of the three resources.
//
// They exist for administration and support — reading a notification by id, searching deliveries
// to see why one was skipped — not for the inbox, which is served by InboxRest. Nothing here
// writes: every write goes through SendNotification or MarkRead, which enforce rules a raw CRUD
// write would bypass.

type notificationRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_notification_notification"`
}

func NewNotificationRest(params notificationRestParams) *NotificationRest {
	rest := &NotificationRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type NotificationRest struct {
	composable.CrudRestBase
}

type recipientRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_notification_recipient"`
}

func NewRecipientRest(params recipientRestParams) *RecipientRest {
	rest := &RecipientRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type RecipientRest struct {
	composable.CrudRestBase
}

type deliveryRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_notification_delivery"`
}

func NewDeliveryRest(params deliveryRestParams) *DeliveryRest {
	rest := &DeliveryRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

type DeliveryRest struct {
	composable.CrudRestBase
}
