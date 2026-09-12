package constants

// Resource codes for authorization. They must be byte-identical to the dynamic-model schema names,
// because the dynamic resource engine asserts permissions using the schema name as the resource
// code. A code that drifts from its schema name denies every request with no obvious cause.
const (
	ResourceNotification          = "notification_notification"
	ResourceNotificationRecipient = "notification_recipient"
	ResourceNotificationDelivery  = "notification_delivery"
)

// Custom action codes, beyond the CRUD set the engine registers for every resource.
const (
	// ActionReadInbox covers the endpoints a user calls for their own notifications: the inbox
	// list, the unread count and the stream. They are one action because they answer the same
	// question — "what is waiting for me" — and differ only in transport.
	ActionReadInbox = "read_inbox"

	// ActionMarkRead covers both the single and the bulk mark-read.
	ActionMarkRead = "mark_read"

	// ActionSend is what an internal module needs to raise a notification. Deliberately NOT
	// granted to the user role: sending is service-principal work (BR 29), and no REST route
	// reaches it.
	ActionSend = "send"
)
