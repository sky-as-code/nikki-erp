package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The purchase order state machine (BR §11-§24), and the agreement's (§40-§46).
//
// Both are pure: no repository, no context, nothing but strings. That is what makes the rules
// testable without a database, and it is why the table lives here rather than inline in the
// operations that consult it.
//
// One record carries the whole RFQ-to-PO cycle (PUR-R1), so what looks like two documents in the
// requirement is two regions of one table here. Confirming does not create a purchase order; it
// moves the request for quotation into the status that makes it one.

// orderTransitions maps each order status to the statuses reachable from it.
//
// `purchase_order` is NOT terminal — an order can still be cancelled after it is confirmed, because
// a vendor can pull out of a deal that both sides had agreed to (BR 23). What cancel does not do is
// erase it: the record and its audit trail stay, which is the difference between cancelling and
// deleting.
//
// `cancelled` IS terminal. Reviving a cancelled order would produce a document whose history says
// it was called off and whose status says it is live, and there is no reading of that pair a vendor
// could rely on. The correct move is to duplicate it, which starts a new RFQ with a new code.
var orderTransitions = map[string][]string{
	string(models.PurchaseOrderStatusRfq): {
		string(models.PurchaseOrderStatusRfqSent),
		string(models.PurchaseOrderStatusToApprove),
		string(models.PurchaseOrderStatusPurchaseOrder),
		string(models.PurchaseOrderStatusCancelled),
	},
	string(models.PurchaseOrderStatusRfqSent): {
		string(models.PurchaseOrderStatusToApprove),
		string(models.PurchaseOrderStatusPurchaseOrder),
		string(models.PurchaseOrderStatusCancelled),
	},
	// to_approve leads only forward or out. There is no route back to rfq: an order waiting on an
	// approver has already been confirmed by its buyer, and putting it back into draft would let
	// the buyer withdraw a request the approver is looking at.
	string(models.PurchaseOrderStatusToApprove): {
		string(models.PurchaseOrderStatusPurchaseOrder),
		string(models.PurchaseOrderStatusCancelled),
	},
	string(models.PurchaseOrderStatusPurchaseOrder): {
		string(models.PurchaseOrderStatusCancelled),
	},
	string(models.PurchaseOrderStatusCancelled): {},
}

// agreementTransitions maps each agreement status to the statuses reachable from it (BR §46).
//
// closed and cancelled are both terminal, and they mean different things: a closed agreement ran
// its course and the orders drawn against it stand, while a cancelled one was called off.
var agreementTransitions = map[string][]string{
	string(models.AgreementStatusDraft): {
		string(models.AgreementStatusConfirmed),
		string(models.AgreementStatusCancelled),
	},
	string(models.AgreementStatusConfirmed): {
		string(models.AgreementStatusClosed),
		string(models.AgreementStatusCancelled),
	},
	string(models.AgreementStatusClosed):    {},
	string(models.AgreementStatusCancelled): {},
}

// CanTransitionOrder reports whether an order may move from one status to another.
//
// A transition to the current status is allowed, so that an idempotent retry is not an error: a
// caller who confirms twice because the first response was lost gets the same answer both times
// rather than a failure on the second.
func CanTransitionOrder(from, to string) bool {
	return canTransition(orderTransitions, from, to)
}

// CanTransitionAgreement reports whether an agreement may move from one status to another.
func CanTransitionAgreement(from, to string) bool {
	return canTransition(agreementTransitions, from, to)
}

func canTransition(table map[string][]string, from, to string) bool {
	if from == to {
		return true
	}
	for _, allowed := range table[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// IsOrderOpen reports whether an order can still be worked on.
func IsOrderOpen(status string) bool {
	return status != string(models.PurchaseOrderStatusCancelled)
}

// IsOrderCommitted reports whether the order represents a commitment to the vendor.
//
// This is the test behind "can this still be freely edited": before confirmation an order is a
// draft the buyer owns, and after it is a document the vendor is holding a copy of.
func IsOrderCommitted(status string) bool {
	return status == string(models.PurchaseOrderStatusPurchaseOrder)
}

// AssertOrderTransition refuses an illegal order transition, naming what was attempted.
//
// The two mistakes a user is most likely to make on purpose get their own message, because "you
// cannot" is not an answer to either. Acting on a cancelled order needs to point at duplicate, and
// re-confirming a committed one needs to say it is already confirmed rather than implying the
// order is in some unexpected state.
func AssertOrderTransition(from, to string, vErrs *ft.ClientErrors) {
	if CanTransitionOrder(from, to) {
		return
	}
	if from == string(models.PurchaseOrderStatusCancelled) {
		appendOrderViolation(vErrs, "purchase_order.cancelled_is_final",
			"a cancelled purchase order cannot be revived; duplicate it to start a new request for quotation")
		return
	}
	if from == string(models.PurchaseOrderStatusPurchaseOrder) &&
		to == string(models.PurchaseOrderStatusToApprove) {
		appendOrderViolation(vErrs, "purchase_order.already_confirmed",
			"this purchase order is already confirmed")
		return
	}
	appendOrderViolation(vErrs, "purchase_order.invalid_transition",
		"a purchase order cannot go from '"+from+"' to '"+to+"'")
}

// AssertAgreementTransition refuses an illegal agreement transition.
func AssertAgreementTransition(from, to string, vErrs *ft.ClientErrors) {
	if CanTransitionAgreement(from, to) {
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(
		models.AgreementSchemaName,
		"purchase_agreement.invalid_transition",
		"a purchase agreement cannot go from '"+from+"' to '"+to+"'",
	))
}

func appendOrderViolation(vErrs *ft.ClientErrors, key, message string) {
	vErrs.Append(*ft.NewBusinessViolation(models.PurchaseOrderSchemaName, key, message))
}
