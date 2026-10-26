package constants

// Action codes beyond the engine's built-in CRUD. Each lifecycle operation carries its own
// permission rather than being folded into "update", because they grant materially different
// powers.
const (
	// Sales channel.
	ActionRegister = "register"
	ActionSuspend  = "suspend"
	ActionActivate = "activate"
	ActionArchive  = "archive"

	// Sales point. Unarchive is deliberately distinct from activate: activate resumes a suspended
	// point, unarchive resurrects a decommissioned one. Granting one must not grant the other.
	ActionUnarchive = "unarchive"

	// Sales channel payment methods.
	ActionEnablePaymentMethod  = "enable_payment_method"
	ActionDisablePaymentMethod = "disable_payment_method"

	// Built-in codes reused by actions that grant no new power of their own.
	ActionRead   = "read"
	ActionCreate = "create"
)
