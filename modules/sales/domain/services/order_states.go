package services

import (
	"slices"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The four status state machines, as pure transition tables. They stay separate because the four
// statuses are independent: a combined machine would enumerate the product of four dimensions and
// make "paid but undelivered" a declared state rather than simply the pair it is.

// orderTransitions maps each order status to the statuses reachable from it.
//
// completed and cancelled are both terminal: a finished sale is corrected through a Return rather
// than reopened, and reviving a cancelled one would give a record whose history says it was called
// off. There is no failed status — a vending sale that took money but dispensed nothing is cancelled
// with payment_status refunded plus an event recording the failure.
var orderTransitions = map[string][]string{
	string(models.SalesOrderStatusDraft): {
		string(models.SalesOrderStatusConfirmed),
		string(models.SalesOrderStatusCancelled),
	},
	// confirmed reaches completed directly when the order needs no fulfilment at all: an order of
	// services or fees has nothing to dispatch.
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

// paymentTransitions maps each payment status to those reachable from it. Money keeps moving after a
// sale is over, so a paid order can still be refunded. unpaid is reachable from partially_paid
// because a payment can be voided before capture; it is not reachable from paid or either refund
// state, since once money is captured the way back is a refund.
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

// fulfillmentTransitions maps each fulfilment status to those reachable from it. not_required is
// terminal and reachable only from pending, because whether an order owes goods is fixed at
// confirmation. fulfilled is not terminal since returns come afterwards; returned is.
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

// invoiceTransitions maps each invoice status to those reachable from it. failed returns to requested
// because a tax authority rejecting a document is usually transient or correctable, and the operator
// fixes it and asks again.
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

// CanTransitionOrderStatus reports whether an order may move from one status to another. An unknown
// from answers false: the value came from a database row, and a status this build does not recognise
// is a reason to refuse, not to crash.
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

// NextOrderStatuses lists what an order may move to.
func NextOrderStatuses(from string) []string {
	return slices.Clone(orderTransitions[from])
}

func canTransition(table map[string][]string, from, to string) bool {
	// A transition to the status already held is a no-op and always allowed, so an idempotent retry
	// is not an error.
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
// paid means captured == payable exactly, never within a rounding tolerance: the amounts are decimals
// at a fixed scale, and a tolerance would leave fractions owed that accumulate across a trading day.
func DerivePaymentStatus(payable, captured, refunded decimal.Decimal) string {
	if refunded.IsPositive() {
		// Refund state takes precedence over payment state once money starts coming back.
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

// LineQuantities is the fulfilment state of one line, as three numbers, so
// DeriveFulfillmentStatus stays pure.
type LineQuantities struct {
	Ordered   decimal.Decimal
	Fulfilled decimal.Decimal
	Returned  decimal.Decimal

	// RequiresFulfillment is false for a line that owes no goods — a service, a fee, a non-stocked
	// item. An order of only such lines is not_required rather than pending forever.
	RequiresFulfillment bool
}

// DeriveFulfillmentStatus answers what an order's fulfilment status should be, from its lines. Return
// states are checked first because a return is later news: reporting a partly returned order as
// fulfilled would hide that goods came back.
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

// DeriveOrderStatus answers whether an order has reached completion, given the other two statuses.
//
// Invoice status is deliberately not an input: an unissued VAT invoice must not hold a commercial
// transaction open. It answers only "should this become completed" and returns the current status
// otherwise, because confirming or cancelling is an operator's act, not a derivation from state.
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
// A reservation settles as redeemed when the order confirms or released when the draft is cancelled
// or expires; a redemption is undone by a return as reversed. Released and reversed both give the
// code its use back but are not interchangeable — a release says no sale happened, a reversal says
// one happened and was returned, and campaign reports count them differently. Nothing returns to
// reserved: the composite unique on (voucher_code_id, sales_order_id) allows one use per order.
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

func CanTransitionVoucherRedemption(from, to string) bool {
	return canTransition(voucherRedemptionTransitions, from, to)
}

func NextVoucherRedemptionStatuses(from string) []string {
	return slices.Clone(voucherRedemptionTransitions[from])
}

// quotationTransitions maps each quotation status to what it may become.
//
// accepted, expired and cancelled are all terminal: an accepted quotation is spent, and re-accepting
// would produce a second order for one agreement. An expired one does not reopen — honouring it later
// belongs in a new quotation with its own dates and freshly priced lines. Nothing returns to draft,
// which would let an offer be edited while the customer holds the version they were sent.
var quotationTransitions = map[string][]string{
	string(models.SalesQuotationStatusDraft): {
		string(models.SalesQuotationStatusSent),
		string(models.SalesQuotationStatusCancelled),
		string(models.SalesQuotationStatusExpired),

		// Direct draft → accepted is permitted: an operator quoting a customer on the phone may take
		// the acceptance in the same conversation.
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

func CanTransitionQuotation(from, to string) bool {
	return canTransition(quotationTransitions, from, to)
}

func NextQuotationStatuses(from string) []string {
	return slices.Clone(quotationTransitions[from])
}
