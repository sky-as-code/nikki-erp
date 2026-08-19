package constants

// Action codes beyond the engine's built-in CRUD.
//
// Each lifecycle operation carries its own permission rather than being folded into "update",
// because they are materially different powers. Approving an order commits the business to a
// purchase; unlocking one reopens a document that was deliberately closed; editing a note does
// neither. A role that may do the last should not thereby be able to do the first.
const (
	// Purchase order.
	ActionSend                = "send"
	ActionConfirm             = "confirm"
	ActionApprove             = "approve"
	ActionLock                = "lock"
	ActionUnlock              = "unlock"
	ActionAcknowledge         = "acknowledge"
	ActionCancel              = "cancel"
	ActionMerge               = "merge"
	ActionCreateAlternative   = "create_alternative"
	ActionCompareAlternatives = "compare_alternatives"

	// Purchase agreement.
	ActionClose     = "close"
	ActionCreateRfq = "create_rfq"

	// Built-in codes reused by actions that grant no new power of their own: printing produces a
	// document the caller can already read, and duplicating is a create.
	ActionRead   = "read"
	ActionCreate = "create"
)
