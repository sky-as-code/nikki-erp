package services

import (
	"slices"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The four status state machines, as pure transition tables.
//
// No repository, no context, no engine. That is deliberate and is what makes them the cheapest
// thing in the module to test: every rule about which status may follow which is decidable from two
// strings, so the tests need no database and can cover the whole table.
//
// The four are separate machines because BR §9 insists the four statuses are independent. A single
// combined machine would have to enumerate the product of four dimensions, most of whose
// combinations are legal and uninteresting, and would make "paid but undelivered" a state to be
// declared rather than simply the pair it is.

// orderTransitions maps each order status to the statuses reachable from it (D-14).
//
// completed and cancelled are both terminal. Completed is terminal because a finished sale is
// corrected through a Return rather than by being reopened — reopening would produce a document
// whose receipt says one thing and whose current state says another. Cancelled is terminal for the
// same reason purchase's is: reviving it would give a record whose history says it was called off
// and whose status says it is live.
//
// There is no `failed` status (D-16). A vending sale where money was captured but nothing was
// dispensed is `cancelled` with `payment_status = refunded` and an event recording the dispense
// failure — one less terminal state to reason about, and the money is where it belongs either way.
var orderTransitions = map[string][]string{
	string(models.SalesOrderStatusDraft): {
		string(models.SalesOrderStatusConfirmed),
		string(models.SalesOrderStatusCancelled),
	},
	// confirmed reaches completed directly, without passing through processing, when the order
	// needs no fulfilment at all (D-14): an order of services or fees has nothing to dispatch, and
	// routing it through a state named for goods movement would be a lie about what happened.
	string(models.SalesOrderStatusConfirmed): {
		string(models.SalesOrderStatusProcessing),
		string(models.SalesOrderStatusCompleted),
		string(models.SalesOrderStatusCancelled),
	},
	string(models.SalesOrderStatusProcessing): {
		string(models.SalesOrderStatusCompleted),
		string(models.SalesOrderStatusCancelled),
	},
	string(models.SalesOrderStatusCompleted): {},
	string(models.SalesOrderStatusCancelled): {},
}

// paymentTransitions maps each payment status to those reachable from it (BR §9.2).
//
// Nothing here is terminal, and that is the point: money keeps moving after a sale is over. A paid
// order can be refunded months later, and a partially refunded one can be refunded again.
//
// unpaid is reachable from partially_paid because a payment can be voided before capture, which
// returns the order to owing its full amount. It is NOT reachable from paid or from either refund
// state: once money has actually been captured, the way back is a refund, which is a different
// status and a different record.
var paymentTransitions = map[string][]string{
	string(models.SalesOrderPaymentStatusUnpaid): {
		string(models.SalesOrderPaymentStatusPartiallyPaid),
		string(models.SalesOrderPaymentStatusPaid),
		string(models.SalesOrderPaymentStatusOverpaid),
	},
	string(models.SalesOrderPaymentStatusPartiallyPaid): {
		string(models.SalesOrderPaymentStatusUnpaid),
		string(models.SalesOrderPaymentStatusPaid),
		string(models.SalesOrderPaymentStatusOverpaid),
		string(models.SalesOrderPaymentStatusPartiallyRefunded),
		string(models.SalesOrderPaymentStatusRefunded),
	},
	string(models.SalesOrderPaymentStatusPaid): {
		string(models.SalesOrderPaymentStatusOverpaid),
		string(models.SalesOrderPaymentStatusPartiallyRefunded),
		string(models.SalesOrderPaymentStatusRefunded),
	},
	// overpaid returns to paid when the excess is given back as change or refunded.
	string(models.SalesOrderPaymentStatusOverpaid): {
		string(models.SalesOrderPaymentStatusPaid),
		string(models.SalesOrderPaymentStatusPartiallyRefunded),
		string(models.SalesOrderPaymentStatusRefunded),
	},
	string(models.SalesOrderPaymentStatusPartiallyRefunded): {
		string(models.SalesOrderPaymentStatusRefunded),
		string(models.SalesOrderPaymentStatusPartiallyRefunded),
	},
	string(models.SalesOrderPaymentStatusRefunded): {},
}

// fulfillmentTransitions maps each fulfilment status to those reachable from it (BR §9.3).
//
// not_required is terminal and reachable only from pending: whether an order owes any goods is
// settled by what its lines are, which is fixed at confirmation. An order cannot discover later
// that it never needed fulfilling.
//
// fulfilled is not terminal — returns come afterwards — but returned is, since everything that
// moved has come back and there is nothing left to move either way.
var fulfillmentTransitions = map[string][]string{
	string(models.SalesOrderFulfillmentStatusPending): {
		string(models.SalesOrderFulfillmentStatusNotRequired),
		string(models.SalesOrderFulfillmentStatusPartiallyFulfilled),
		string(models.SalesOrderFulfillmentStatusFulfilled),
	},
	string(models.SalesOrderFulfillmentStatusNotRequired): {},
	string(models.SalesOrderFulfillmentStatusPartiallyFulfilled): {
		string(models.SalesOrderFulfillmentStatusFulfilled),
		string(models.SalesOrderFulfillmentStatusPartiallyReturned),
		string(models.SalesOrderFulfillmentStatusReturned),
	},
	string(models.SalesOrderFulfillmentStatusFulfilled): {
		string(models.SalesOrderFulfillmentStatusPartiallyReturned),
		string(models.SalesOrderFulfillmentStatusReturned),
	},
	string(models.SalesOrderFulfillmentStatusPartiallyReturned): {
		string(models.SalesOrderFulfillmentStatusReturned),
		string(models.SalesOrderFulfillmentStatusPartiallyReturned),
	},
	string(models.SalesOrderFulfillmentStatusReturned): {},
}

// invoiceTransitions maps each invoice status to those reachable from it (BR §9.4).
//
// failed returns to requested because retrying is the whole point of recording the failure: a tax
// authority rejecting a document is usually a transient or correctable condition, and the operator
// fixes it and asks again. D-17 applies the same reasoning to returns.
var invoiceTransitions = map[string][]string{
	string(models.SalesOrderInvoiceStatusNotRequested): {
		string(models.SalesOrderInvoiceStatusRequested),
	},
	string(models.SalesOrderInvoiceStatusRequested): {
		string(models.SalesOrderInvoiceStatusIssued),
		string(models.SalesOrderInvoiceStatusFailed),
		string(models.SalesOrderInvoiceStatusCancelled),
	},
	string(models.SalesOrderInvoiceStatusFailed): {
		string(models.SalesOrderInvoiceStatusRequested),
		string(models.SalesOrderInvoiceStatusCancelled),
	},
	// An issued fiscal document can be cancelled with the authority, but never un-issued.
	string(models.SalesOrderInvoiceStatusIssued): {
		string(models.SalesOrderInvoiceStatusCancelled),
	},
	string(models.SalesOrderInvoiceStatusCancelled): {},
}

// CanTransitionOrderStatus reports whether an order may move from one status to another.
//
// An unknown `from` answers false rather than panicking: the value came from a database row, and a
// status this build does not recognise is a reason to refuse the transition, not to crash the
// request.
func CanTransitionOrderStatus(from, to string) bool {
	return canTransition(orderTransitions, from, to)
}

func CanTransitionPaymentStatus(from, to string) bool {
	return canTransition(paymentTransitions, from, to)
}

func CanTransitionFulfillmentStatus(from, to string) bool {
	return canTransition(fulfillmentTransitions, from, to)
}

func CanTransitionInvoiceStatus(from, to string) bool {
	return canTransition(invoiceTransitions, from, to)
}

// NextOrderStatuses lists what an order may move to, for a caller building a UI or an error message
// that says what WOULD have been allowed.
func NextOrderStatuses(from string) []string {
	return slices.Clone(orderTransitions[from])
}

func canTransition(table map[string][]string, from, to string) bool {
	// A transition to the status already held is always allowed, and is a no-op. Refusing it would
	// make every idempotent retry an error: a caller re-confirming an already-confirmed order has
	// asked for a state that holds, and reporting failure would drive it to keep retrying.
	if from == to {
		return true
	}
	allowed, known := table[from]
	if !known {
		return false
	}
	return slices.Contains(allowed, to)
}

// DerivePaymentStatus answers what an order's payment status should be, from the money against it.
//
// Pure: it takes the numbers rather than reading them, so the rule can be tested exhaustively and
// so the caller controls which transaction the numbers came from.
//
// `paid iff captured == payable` EXACTLY (BR §42) — not "within a rounding tolerance". The amounts
// are decimals at a fixed scale precisely so that an exact comparison is meaningful; a tolerance
// would let a sale be marked paid while a fraction remained owed, and those fractions would
// accumulate across a day of trading.
func DerivePaymentStatus(payable, captured, refunded decimal.Decimal) string {
	if refunded.IsPositive() {
		// Refund state takes precedence over payment state: what matters once money has started
		// coming back is how much has, not that it was once fully paid.
		if refunded.GreaterThanOrEqual(captured) {
			return string(models.SalesOrderPaymentStatusRefunded)
		}
		return string(models.SalesOrderPaymentStatusPartiallyRefunded)
	}
	if !captured.IsPositive() {
		return string(models.SalesOrderPaymentStatusUnpaid)
	}
	if captured.GreaterThan(payable) {
		return string(models.SalesOrderPaymentStatusOverpaid)
	}
	if captured.Equal(payable) {
		return string(models.SalesOrderPaymentStatusPaid)
	}
	return string(models.SalesOrderPaymentStatusPartiallyPaid)
}

// LineQuantities is the fulfilment state of one line, as three numbers.
//
// Taking a slice of these rather than the line records keeps DeriveFulfillmentStatus pure and makes
// its tests readable: the interesting cases are combinations of quantities, not of database rows.
type LineQuantities struct {
	Ordered   decimal.Decimal
	Fulfilled decimal.Decimal
	Returned  decimal.Decimal

	// RequiresFulfillment is false for a line that owes no goods — a service, a fee, a non-stocked
	// item. An order of only such lines is not_required rather than pending forever (D-14).
	RequiresFulfillment bool
}

// DeriveFulfillmentStatus answers what an order's fulfilment status should be, from its lines.
//
// The return states are checked before the fulfilment ones because a return is later news: an order
// that was fulfilled and then partly returned is partially_returned, and reporting it as fulfilled
// would hide that goods came back.
func DeriveFulfillmentStatus(lines []LineQuantities) string {
	var (
		ordered   = decimal.Zero
		fulfilled = decimal.Zero
		returned  = decimal.Zero
		anyNeeded bool
	)
	for _, line := range lines {
		if !line.RequiresFulfillment {
			continue
		}
		anyNeeded = true
		ordered = ordered.Add(line.Ordered)
		fulfilled = fulfilled.Add(line.Fulfilled)
		returned = returned.Add(line.Returned)
	}

	// No line owes anything — every line is a service or a fee, or the order has no lines at all.
	if !anyNeeded {
		return string(models.SalesOrderFulfillmentStatusNotRequired)
	}

	if returned.IsPositive() {
		if returned.GreaterThanOrEqual(fulfilled) {
			return string(models.SalesOrderFulfillmentStatusReturned)
		}
		return string(models.SalesOrderFulfillmentStatusPartiallyReturned)
	}
	if !fulfilled.IsPositive() {
		return string(models.SalesOrderFulfillmentStatusPending)
	}
	if fulfilled.GreaterThanOrEqual(ordered) {
		return string(models.SalesOrderFulfillmentStatusFulfilled)
	}
	return string(models.SalesOrderFulfillmentStatusPartiallyFulfilled)
}

// DeriveOrderStatus answers whether an order has reached completion, given the other two statuses
// (D-15).
//
// Invoice status is deliberately absent from the inputs. BR §9 insists the four are independent,
// and an unissued VAT invoice must not hold a commercial transaction open: the goods are delivered,
// the money is in, and the fiscal document is somebody else's problem to chase.
//
// It answers only "should this become completed", returning the current status otherwise. Deciding
// draft-to-confirmed or anything-to-cancelled is an operator's act, not a derivation from state,
// and a function that guessed at those would cancel orders nobody asked to cancel.
func DeriveOrderStatus(current, paymentStatus, fulfillmentStatus string) string {
	if current != string(models.SalesOrderStatusConfirmed) &&
		current != string(models.SalesOrderStatusProcessing) {
		return current
	}
	if paymentStatus != string(models.SalesOrderPaymentStatusPaid) {
		return current
	}
	switch models.SalesOrderFulfillmentStatus(fulfillmentStatus) {
	case models.SalesOrderFulfillmentStatusFulfilled,
		models.SalesOrderFulfillmentStatusNotRequired:
		return string(models.SalesOrderStatusCompleted)
	}
	return current
}

// voucherRedemptionTransitions maps each redemption status to what it may become.
//
// A reservation is the only non-terminal state, and it settles exactly one of two ways: 'redeemed'
// when the order confirms, 'released' when the draft is cancelled or expires. A redemption can then
// be undone by a return, which is 'reversed'.
//
// Released and reversed are both terminal and both mean "the code has its use back", but they are
// not interchangeable: a release says no sale ever happened, a reversal says one happened and was
// returned. A campaign report counts them differently, and collapsing them would lose that.
//
// Nothing returns to 'reserved'. Re-applying a released code to the same order writes a new
// redemption rather than reviving the old one, because the composite unique on
// (voucher_code_id, sales_order_id) would refuse it — which is the correct refusal: BR 26 allows one
// use of a code per order, and a released reservation has already recorded that this order used it.
var voucherRedemptionTransitions = map[string][]string{
	string(models.VoucherRedemptionStatusReserved): {
		string(models.VoucherRedemptionStatusRedeemed),
		string(models.VoucherRedemptionStatusReleased),
	},
	string(models.VoucherRedemptionStatusRedeemed): {
		string(models.VoucherRedemptionStatusReversed),
	},
	string(models.VoucherRedemptionStatusReleased): {},
	string(models.VoucherRedemptionStatusReversed): {},
}

// CanTransitionVoucherRedemption reports whether a redemption may move between two statuses.
func CanTransitionVoucherRedemption(from, to string) bool {
	return canTransition(voucherRedemptionTransitions, from, to)
}

// NextVoucherRedemptionStatuses lists what a redemption may become.
func NextVoucherRedemptionStatuses(from string) []string {
	return slices.Clone(voucherRedemptionTransitions[from])
}

// quotationTransitions maps each quotation status to what it may become (BR 87.1).
//
// The shape worth noticing: `draft` and `sent` may BOTH expire and BOTH be cancelled, and neither
// `accepted`, `expired` nor `cancelled` goes anywhere. An accepted quotation is spent — it produced
// an order, and re-accepting it would produce a second, which is two deliveries for one agreement.
//
// An expired quotation does not reopen. The customer was given a deadline and it passed; honouring
// it afterwards is a commercial decision that belongs in a NEW quotation carrying its own dates,
// rather than in silently reviving one whose terms nobody re-agreed to. Reviving would also leave the
// stored prices unrepriced, which is exactly the staleness the expiry existed to prevent.
//
// Nothing returns to `draft`. A quotation the customer has seen cannot be un-seen, so un-sending one
// would let an offer be edited while the customer holds the version they were sent.
var quotationTransitions = map[string][]string{
	string(models.SalesQuotationStatusDraft): {
		string(models.SalesQuotationStatusSent),
		string(models.SalesQuotationStatusCancelled),
		string(models.SalesQuotationStatusExpired),

		// Direct draft → accepted is permitted: a back-office operator quoting a customer on the
		// phone may take the acceptance in the same conversation, and forcing a send in between
		// would be recording a step that did not happen.
		string(models.SalesQuotationStatusAccepted),
	},
	string(models.SalesQuotationStatusSent): {
		string(models.SalesQuotationStatusAccepted),
		string(models.SalesQuotationStatusExpired),
		string(models.SalesQuotationStatusCancelled),
	},
	string(models.SalesQuotationStatusAccepted):  {},
	string(models.SalesQuotationStatusExpired):   {},
	string(models.SalesQuotationStatusCancelled): {},
}

// CanTransitionQuotation reports whether a quotation may move between two statuses.
func CanTransitionQuotation(from, to string) bool {
	return canTransition(quotationTransitions, from, to)
}

// NextQuotationStatuses lists what a quotation may become.
func NextQuotationStatuses(from string) []string {
	return slices.Clone(quotationTransitions[from])
}
