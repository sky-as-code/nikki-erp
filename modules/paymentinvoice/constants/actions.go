package constants

// Action codes for the resource-specific engine actions this module defines, over and above the
// engine's built-in CRUD. Each is also the permission code checked for that action, so the two are
// declared together and never spelled separately.
//
// The names are snake_case because DefineAction rejects a RestPath containing a hyphen.
const (
	// ActionCreatePayment starts a payment: it records the order and asks the gateway for the
	// payment instrument (a QR code, a pay URL, or a prompt pushed to a card terminal).
	ActionCreatePayment = "create_payment"

	// ActionRefund returns money for an order already paid.
	ActionRefund = "refund"

	// ActionRemovePosOrders clears the orders queued on one card terminal. It replaces the
	// unauthenticated DELETE /mpos/pos-orders/:posId of the service this module supersedes.
	ActionRemovePosOrders = "remove_pos_orders"

	// ActionIssue closes an invoice draft: it recomputes the totals, assigns the number and
	// stamps the issue date.
	ActionIssue = "issue"
)
