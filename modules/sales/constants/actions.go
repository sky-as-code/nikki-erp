package constants

// Action codes beyond the engine's built-in CRUD.
//
// Each lifecycle operation carries its own permission rather than being folded into "update",
// because they are materially different powers. Confirming an order commits the business to a
// sale and redeems the customer's voucher; suspending a sales point stops a whole kiosk selling;
// editing a display name does neither. A role that may do the last should not thereby be able to
// do the first.
//
// Codes are declared as they are needed by the tasks that introduce them. The lifecycle actions
// below belong to SALES-004; the transactional ones arrive with their own tasks.
const (
	// Sales channel.
	ActionRegister = "register"
	ActionSuspend  = "suspend"
	ActionActivate = "activate"
	ActionArchive  = "archive"

	// Sales point. Unarchive is deliberately distinct from activate: activating a suspended point
	// resumes selling, while unarchiving resurrects a point that was decommissioned. Granting the
	// first should not grant the second.
	ActionUnarchive = "unarchive"

	// Sales channel payment methods.
	ActionEnablePaymentMethod  = "enable_payment_method"
	ActionDisablePaymentMethod = "disable_payment_method"

	// Built-in codes reused by actions that grant no new power of their own.
	ActionRead   = "read"
	ActionCreate = "create"
)
