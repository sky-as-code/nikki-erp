package constants

// Action codes beyond the engine's built-in CRUD.
//
// Publishing is a separate power from updating, and deliberately so: an update edits a draft that
// nothing has calculated with, while publishing makes a configuration binding on every subsequent
// transaction and freezes its material fields forever (BR-TAX-ESS-SUP-002). A role that may
// correct a typo in a draft rate should not thereby be able to put that rate into effect.
//
// Override and simulate are named by the requirement itself (BR-TAX-ESS-053) rather than chosen
// here. Override lets a user substitute the determined tax set on a transaction, which is why it
// is not folded into any CRUD power on the tax master.
const (
	// Configuration lifecycle. Withdraw is distinct from archive: withdrawing retires a
	// configuration from new determination while leaving it readable, whereas archiving is the
	// system-level hide that BR-TAX-ESS-SUP-002 insists must not substitute for lifecycle state.
	ActionPublish  = "publish"
	ActionWithdraw = "withdraw"

	// Manual tax override on a calculation (BR-TAX-ESS-024). Requires a reason.
	ActionOverride = "override"

	// Tax Simulator (BR-TAX-ESS-051). Runs calculation only and creates no transaction.
	ActionSimulate = "simulate"

	// Built-in codes reused by actions that grant no new power of their own.
	ActionRead   = "read"
	ActionCreate = "create"
)
