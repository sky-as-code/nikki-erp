package constants

// Action codes beyond the engine's built-in CRUD.
//
// Publishing is a separate power from updating: an update edits a draft nothing has calculated
// with, while publishing makes a configuration binding on every later transaction and freezes its
// material fields forever. Override substitutes the determined tax set on a transaction, so it is
// likewise not folded into any CRUD power on the tax master.
const (
	// Configuration lifecycle. Withdraw retires a configuration from new determination while
	// leaving it readable; it is not the system-level archive, which must not substitute for it.
	ActionPublish  = "publish"
	ActionWithdraw = "withdraw"

	// Manual tax override on a calculation. Requires a reason.
	ActionOverride = "override"

	// Tax Simulator. Runs calculation only and creates no transaction.
	ActionSimulate = "simulate"

	// Built-in codes reused by actions that grant no new power of their own.
	ActionRead   = "read"
	ActionCreate = "create"
)
