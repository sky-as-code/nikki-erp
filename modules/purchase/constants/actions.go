package constants

// Action codes beyond the built-in CRUD. Each lifecycle operation carries its own
// permission rather than being folded into "update", because approving an order or unlocking a
// closed document is a materially different power from editing a note.
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
	ActionDuplicate           = "duplicate"

	// ActionReprice is its own power: repricing rewrites what the company will pay across a whole
	// order, but it commits nothing, so it is neither "update" nor "confirm".
	ActionReprice = "reprice"

	// Purchase agreement.
	ActionClose     = "close"
	ActionCreateRfq = "create_rfq"

	// Built-in codes reused by actions granting no new power: printing produces a readable document,
	// duplicating is a create.
	ActionRead   = "read"
	ActionCreate = "create"
)
