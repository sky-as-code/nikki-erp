package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The purchase order and agreement state machines. Both are pure string tables, so the rules are
// testable without a database. One record carries the whole RFQ-to-PO cycle: confirming does not
// create a purchase order, it moves the request for quotation into the status that makes it one.

// orderTransitions maps each order status to the statuses reachable from it. `purchase_order` is
// not terminal: a confirmed order can still be cancelled, which keeps the record and its audit
// trail rather than erasing them. `cancelled` is terminal — reviving one would contradict its own
// history, so the correct move is to duplicate it into a new RFQ with a new code.
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
	// to_approve leads only forward or out; a route back to rfq would let the buyer withdraw a
	// request the approver is already looking at.
	string(models.PurchaseOrderStatusToApprove): {
		string(models.PurchaseOrderStatusPurchaseOrder),
		string(models.PurchaseOrderStatusCancelled),
	},
	string(models.PurchaseOrderStatusPurchaseOrder): {
		string(models.PurchaseOrderStatusCancelled),
	},
	string(models.PurchaseOrderStatusCancelled): {},
}

// agreementTransitions maps each agreement status to the statuses reachable from it. closed and
// cancelled are both terminal but differ in meaning: a closed agreement ran its course and the
// orders drawn against it stand, a cancelled one was called off.
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

// CanTransitionOrder reports whether an order may move from one status to another. A move to the
// current status is allowed so that a retried confirm is idempotent rather than an error.
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

// IsOrderCommitted reports whether the order represents a commitment to the vendor. This is the
// test behind "can this still be freely edited".
func IsOrderCommitted(status string) bool {
	return status == string(models.PurchaseOrderStatusPurchaseOrder)
}

// AssertOrderTransition refuses an illegal order transition. The two likely mistakes — acting on a
// cancelled order and re-confirming a committed one — get their own messages.
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
