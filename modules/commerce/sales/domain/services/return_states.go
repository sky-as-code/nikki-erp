package services

import (
	"slices"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The return state machine. Both rules are pure functions over values rather than records, so
// every branch is testable without a database and the caller decides where the inputs came from.

// returnTransitions maps each return status to those reachable from it.
//
// A return may be cancelled from draft or approved but NOT from processing: by then goods may be
// moving or money leaving, and neither can be un-asked. completed is terminal, but a completed
// return may still carry a failed fiscal adjustment; retrying it does not reopen the return.
var returnTransitions = map[string][]string{
	string(models.SalesReturnStatusDraft): {
		string(models.SalesReturnStatusApproved),
		string(models.SalesReturnStatusCancelled),
	},
	string(models.SalesReturnStatusApproved): {
		string(models.SalesReturnStatusProcessing),
		string(models.SalesReturnStatusCancelled),
	},
	// No route back to approved or to cancelled. Once a side effect is in flight the way out of a
	// mistake is another return or a fresh sale, not an undo.
	string(models.SalesReturnStatusProcessing): {
		string(models.SalesReturnStatusCompleted),
	},
	string(models.SalesReturnStatusCompleted): {},
	string(models.SalesReturnStatusCancelled): {},
}

// CanTransitionReturn reports whether a return may move from one status to another.
func CanTransitionReturn(from, to string) bool {
	return canTransition(returnTransitions, from, to)
}

// stepSettled reports whether one of a return's three side effects is finished with.
//
// not_required counts as settled: a return of services owes no goods, so its inventory step is
// satisfied by not applying. Requiring an actual completion would leave such returns open forever.
func stepSettled(status string) bool {
	return status == string(models.SalesReturnStepCompleted) ||
		status == string(models.SalesReturnStepNotRequired)
}

// DeriveReturnStatus answers what a return's status should be, from the three steps beneath it.
//
// A return is complete once the two CUSTOMER-FACING steps are settled; the fiscal adjustment is
// deliberately not consulted, because the customer has their money and the goods are back, and a
// failed adjustment stays visible on its own column as a retryable job for whoever handles tax.
//
// A failed customer-facing step produces no status of its own: the return stays `processing` and
// carries the failure reason, because a failed refund is a job someone must finish.
func DeriveReturnStatus(current, inventoryStatus, refundStatus string) string {
	if current == string(models.SalesReturnStatusCancelled) {
		return current
	}
	if stepSettled(inventoryStatus) && stepSettled(refundStatus) {
		return string(models.SalesReturnStatusCompleted)
	}
	if current == string(models.SalesReturnStatusApproved) ||
		current == string(models.SalesReturnStatusProcessing) {
		return string(models.SalesReturnStatusProcessing)
	}
	return current
}

// ReturnableLine carries what is needed to decide how much of one order line may still come back.
//
// RequiresFulfillment is snapshotted onto the line at order creation rather than looked up now: a
// later change to the product master must not alter what an existing order allows to be returned.
type ReturnableLine struct {
	Ordered             decimal.Decimal
	Fulfilled           decimal.Decimal
	PreviouslyReturned  decimal.Decimal
	RequiresFulfillment bool
}

// ReturnableQuantity answers how much of a line may still be returned.
//
// For goods that ship, the basis is what was HANDED OVER minus what has already come back. For
// services, digital items and anything not stock-tracked there was never a delivery step, so the
// basis is what was ORDERED minus what has come back; measuring those against delivery would make
// them permanently un-refundable.
//
// PreviouslyReturned counts quantities from ACCEPTED returns, not merely those whose stock has
// physically moved: counting only completed movements would let the same quantity be refunded
// twice while the first return is in flight.
//
// Never negative: an over-returned line (possible only through data repair) yields zero, which
// would otherwise underflow any comparison against it.
func ReturnableQuantity(line ReturnableLine) decimal.Decimal {
	basis := line.Fulfilled
	if !line.RequiresFulfillment {
		basis = line.Ordered
	}

	remaining := basis.Sub(line.PreviouslyReturned)
	if remaining.IsNegative() {
		return decimal.Zero
	}
	return remaining
}

// RequiresInventoryReturn reports whether a return line's goods physically come back. A return
// line freezes the answer, so a product reclassified afterwards cannot change what an in-flight
// return is waiting for.
func RequiresInventoryReturn(line ReturnableLine) bool {
	return line.RequiresFulfillment
}

// DeriveInventoryStepStatus answers what a return's inventory step should start as: `not_required`
// when no line owes goods, which is what lets a services-only return complete on the refund alone.
func DeriveInventoryStepStatus(lines []ReturnableLine) string {
	if slices.ContainsFunc(lines, RequiresInventoryReturn) {
		return string(models.SalesReturnStepPending)
	}
	return string(models.SalesReturnStepNotRequired)
}

// DeriveRefundStepStatus answers what a return's refund step should start as: `not_required` when
// nothing is owed back, as for a return against an order that was never paid. Requiring a
// completion would hold such a return open against a refund that will never be made.
func DeriveRefundStepStatus(refundTotal decimal.Decimal) string {
	if refundTotal.IsZero() || refundTotal.IsNegative() {
		return string(models.SalesReturnStepNotRequired)
	}
	return string(models.SalesReturnStepPending)
}
