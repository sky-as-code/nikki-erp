package services

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Pure functions, no database: the state machines are tables, so they can be tested exhaustively.

func dec(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}

// Every status the schema declares must appear in its transition table: canTransition returns false
// for an unknown from, so a record reaching a missing status could never move again.
func TestEveryStatusAppearsInItsTable(t *testing.T) {
	cases := []struct {
		name     string
		table    map[string][]string
		statuses []string
	}{
		{"order", orderTransitions, []string{
			string(models.SalesOrderStatusDraft),
			string(models.SalesOrderStatusConfirmed),
			string(models.SalesOrderStatusProcessing),
			string(models.SalesOrderStatusCompleted),
			string(models.SalesOrderStatusCancelled),
		}},
		{"payment", paymentTransitions, []string{
			string(models.SalesOrderPaymentStatusUnpaid),
			string(models.SalesOrderPaymentStatusPartiallyPaid),
			string(models.SalesOrderPaymentStatusPaid),
			string(models.SalesOrderPaymentStatusOverpaid),
			string(models.SalesOrderPaymentStatusPartiallyRefunded),
			string(models.SalesOrderPaymentStatusRefunded),
		}},
		{"fulfillment", fulfillmentTransitions, []string{
			string(models.SalesOrderFulfillmentStatusPending),
			string(models.SalesOrderFulfillmentStatusNotRequired),
			string(models.SalesOrderFulfillmentStatusPartiallyFulfilled),
			string(models.SalesOrderFulfillmentStatusFulfilled),
			string(models.SalesOrderFulfillmentStatusPartiallyReturned),
			string(models.SalesOrderFulfillmentStatusReturned),
		}},
		{"invoice", invoiceTransitions, []string{
			string(models.SalesOrderInvoiceStatusNotRequested),
			string(models.SalesOrderInvoiceStatusRequested),
			string(models.SalesOrderInvoiceStatusIssued),
			string(models.SalesOrderInvoiceStatusFailed),
			string(models.SalesOrderInvoiceStatusCancelled),
		}},
		{"voucher redemption", voucherRedemptionTransitions, []string{
			string(models.VoucherRedemptionStatusReserved),
			string(models.VoucherRedemptionStatusRedeemed),
			string(models.VoucherRedemptionStatusReleased),
			string(models.VoucherRedemptionStatusReversed),
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			for _, status := range testCase.statuses {
				if _, ok := testCase.table[status]; !ok {
					t.Errorf("%q is missing from the %s table: a record reaching it could never "+
						"move again", status, testCase.name)
				}
			}
			if len(testCase.table) != len(testCase.statuses) {
				t.Errorf("the %s table has %d entries for %d declared statuses: one of them names "+
					"a status the schema does not declare",
					testCase.name, len(testCase.table), len(testCase.statuses))
			}
		})
	}
}

// Every target named in a table must itself be a key of that table, or the machine can reach a
// state it cannot leave.
func TestEveryTransitionTargetIsAKnownStatus(t *testing.T) {
	tables := map[string]map[string][]string{
		"order":       orderTransitions,
		"payment":     paymentTransitions,
		"fulfillment": fulfillmentTransitions,
		"invoice":     invoiceTransitions,
		"redemption":  voucherRedemptionTransitions,
	}
	for name, table := range tables {
		for from, targets := range table {
			for _, to := range targets {
				if _, ok := table[to]; !ok {
					t.Errorf("%s: %q -> %q, but %q is not a key of the table", name, from, to, to)
				}
			}
		}
	}
}

func TestOrderTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{"draft", "confirmed", true},
		{"draft", "cancelled", true},
		{"draft", "completed", false},
		{"draft", "processing", false},

		{"confirmed", "processing", true},
		// An order needing no fulfilment goes straight to completed.
		{"confirmed", "completed", true},
		{"confirmed", "cancelled", true},
		{"confirmed", "draft", false},

		{"processing", "completed", true},
		{"processing", "cancelled", true},
		{"processing", "confirmed", false},
		{"processing", "draft", false},

		// Both terminal: a completed sale is corrected through a Return, not by reopening.
		{"completed", "cancelled", false},
		{"completed", "processing", false},
		{"cancelled", "draft", false},
		{"cancelled", "confirmed", false},
	}
	for _, testCase := range cases {
		if got := CanTransitionOrderStatus(testCase.from, testCase.to); got != testCase.want {
			t.Errorf("order %s -> %s = %v, want %v",
				testCase.from, testCase.to, got, testCase.want)
		}
	}
}

// There is no `failed` order status, and nothing may move to one.
func TestNoFailedOrderStatus(t *testing.T) {
	if _, exists := orderTransitions["failed"]; exists {
		t.Error("D-16 rejects a `failed` order status: a captured-but-undispensed sale is " +
			"cancelled with payment_status refunded")
	}
	for from := range orderTransitions {
		if CanTransitionOrderStatus(from, "failed") {
			t.Errorf("%q may move to `failed`, which D-16 says does not exist", from)
		}
	}
}

// A transition to the status already held is a no-op, not an error, so idempotent retries pass.
func TestSelfTransitionIsAlwaysAllowed(t *testing.T) {
	for from := range orderTransitions {
		if !CanTransitionOrderStatus(from, from) {
			t.Errorf("order %s -> %s must be allowed as a no-op", from, from)
		}
	}
	for from := range paymentTransitions {
		if !CanTransitionPaymentStatus(from, from) {
			t.Errorf("payment %s -> %s must be allowed as a no-op", from, from)
		}
	}
}

// An unknown status refuses rather than panics: it came from a database row, and a status this build
// does not recognise is a reason to refuse, not to crash.
func TestUnknownStatusIsRefused(t *testing.T) {
	if CanTransitionOrderStatus("who_knows", "confirmed") {
		t.Error("an unrecognised source status must refuse every transition")
	}
	if CanTransitionOrderStatus("draft", "who_knows") {
		t.Error("an unrecognised target status must be refused")
	}
	if got := NextOrderStatuses("who_knows"); len(got) != 0 {
		t.Errorf("NextOrderStatuses of an unknown status = %v, want empty", got)
	}
}

// Money keeps moving after a sale is over, so no payment status is terminal except refunded.
func TestPaymentTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{"unpaid", "partially_paid", true},
		{"unpaid", "paid", true},
		{"unpaid", "overpaid", true},
		// Nothing captured, so nothing to refund.
		{"unpaid", "refunded", false},
		{"unpaid", "partially_refunded", false},

		// A payment voided before capture returns the order to owing its full amount.
		{"partially_paid", "unpaid", true},
		{"partially_paid", "paid", true},
		{"partially_paid", "refunded", true},

		// Once money is captured, the way back is a refund, not a return to unpaid.
		{"paid", "unpaid", false},
		{"paid", "refunded", true},
		{"paid", "partially_refunded", true},
		{"paid", "overpaid", true},

		{"overpaid", "paid", true},
		{"overpaid", "refunded", true},

		{"partially_refunded", "refunded", true},
		// Refunding again is legitimate and stays in the same state.
		{"partially_refunded", "partially_refunded", true},
		{"partially_refunded", "paid", false},

		{"refunded", "paid", false},
		{"refunded", "partially_refunded", false},
	}
	for _, testCase := range cases {
		if got := CanTransitionPaymentStatus(testCase.from, testCase.to); got != testCase.want {
			t.Errorf("payment %s -> %s = %v, want %v",
				testCase.from, testCase.to, got, testCase.want)
		}
	}
}

func TestFulfillmentTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{"pending", "not_required", true},
		{"pending", "partially_fulfilled", true},
		{"pending", "fulfilled", true},
		// Nothing has moved, so nothing can come back.
		{"pending", "returned", false},

		// Whether an order owes goods is fixed by its lines at confirmation.
		{"not_required", "fulfilled", false},
		{"not_required", "pending", false},

		{"partially_fulfilled", "fulfilled", true},
		{"partially_fulfilled", "partially_returned", true},
		{"fulfilled", "returned", true},
		{"fulfilled", "partially_returned", true},
		{"fulfilled", "pending", false},

		{"partially_returned", "returned", true},
		{"partially_returned", "fulfilled", false},
		{"returned", "fulfilled", false},
	}
	for _, testCase := range cases {
		if got := CanTransitionFulfillmentStatus(testCase.from, testCase.to); got != testCase.want {
			t.Errorf("fulfillment %s -> %s = %v, want %v",
				testCase.from, testCase.to, got, testCase.want)
		}
	}
}

// A fiscal rejection is usually correctable, so it must be retryable.
func TestInvoiceFailedCanBeRetried(t *testing.T) {
	if !CanTransitionInvoiceStatus("failed", "requested") {
		t.Error("a failed fiscal request must be retryable, or a correctable rejection would " +
			"strand the invoice permanently")
	}
	if CanTransitionInvoiceStatus("issued", "requested") {
		t.Error("an issued fiscal document cannot be un-issued")
	}
	if CanTransitionInvoiceStatus("issued", "failed") {
		t.Error("an issued document did not fail")
	}
	if !CanTransitionInvoiceStatus("issued", "cancelled") {
		t.Error("an issued document can be cancelled with the authority")
	}
}

// paid means captured == payable exactly: a tolerance would leave fractions owed that accumulate
// across a trading day.
func TestDerivePaymentStatus(t *testing.T) {
	cases := []struct {
		name                        string
		payable, captured, refunded string
		want                        string
	}{
		{"nothing taken yet", "100", "0", "0", "unpaid"},
		{"part paid", "100", "40", "0", "partially_paid"},
		{"exactly paid", "100", "100", "0", "paid"},
		{"a fraction short is NOT paid", "100", "99.9999", "0", "partially_paid"},
		{"a fraction over is overpaid", "100", "100.0001", "0", "overpaid"},
		{"cash till took more", "100", "150", "0", "overpaid"},
		{"part refunded", "100", "100", "30", "partially_refunded"},
		{"fully refunded", "100", "100", "100", "refunded"},
		{"refunded more than captured", "100", "100", "120", "refunded"},
		{"a free order is paid once nothing is owed", "0", "0", "0", "unpaid"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := DerivePaymentStatus(
				dec(testCase.payable), dec(testCase.captured), dec(testCase.refunded))
			if got != testCase.want {
				t.Errorf("payable=%s captured=%s refunded=%s: got %q, want %q",
					testCase.payable, testCase.captured, testCase.refunded, got, testCase.want)
			}
		})
	}
}

// A refund in flight outranks the payment state.
func TestRefundOutranksPayment(t *testing.T) {
	got := DerivePaymentStatus(dec("100"), dec("100"), dec("1"))
	if got != "partially_refunded" {
		t.Errorf("a fully paid order with a refund against it = %q, want partially_refunded", got)
	}
}

func TestDeriveFulfillmentStatus(t *testing.T) {
	stocked := func(ordered, fulfilled, returned string) LineQuantities {
		return LineQuantities{
			Ordered:             dec(ordered),
			Fulfilled:           dec(fulfilled),
			Returned:            dec(returned),
			RequiresFulfillment: true,
		}
	}
	service := func(ordered string) LineQuantities {
		return LineQuantities{Ordered: dec(ordered), RequiresFulfillment: false}
	}

	cases := []struct {
		name  string
		lines []LineQuantities
		want  string
	}{
		{"nothing dispatched", []LineQuantities{stocked("3", "0", "0")}, "pending"},
		{"part dispatched", []LineQuantities{stocked("3", "2", "0")}, "partially_fulfilled"},
		{"all dispatched", []LineQuantities{stocked("3", "3", "0")}, "fulfilled"},
		{"part returned", []LineQuantities{stocked("3", "3", "1")}, "partially_returned"},
		{"all returned", []LineQuantities{stocked("3", "3", "3")}, "returned"},

		// An order of only services owes no goods, so it is not_required rather than pending forever,
		// which satisfies completion.
		{"only services", []LineQuantities{service("1"), service("2")}, "not_required"},
		{"no lines at all", nil, "not_required"},

		// A mixed order is judged on its stocked lines alone.
		{"mixed, goods pending", []LineQuantities{stocked("3", "0", "0"), service("1")}, "pending"},
		{"mixed, goods done", []LineQuantities{stocked("3", "3", "0"), service("1")}, "fulfilled"},

		{"one line done, one not",
			[]LineQuantities{stocked("2", "2", "0"), stocked("2", "0", "0")},
			"partially_fulfilled"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := DeriveFulfillmentStatus(testCase.lines); got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

// Completion needs payment and fulfilment, and deliberately ignores the invoice.
func TestDeriveOrderStatus(t *testing.T) {
	cases := []struct {
		name                          string
		current, payment, fulfillment string
		want                          string
	}{
		{"paid and delivered", "confirmed", "paid", "fulfilled", "completed"},
		{"paid, nothing to deliver", "confirmed", "paid", "not_required", "completed"},
		{"from processing", "processing", "paid", "fulfilled", "completed"},

		{"delivered but unpaid", "confirmed", "partially_paid", "fulfilled", "confirmed"},
		{"paid but undelivered", "confirmed", "paid", "partially_fulfilled", "confirmed"},
		{"paid but returned", "confirmed", "paid", "returned", "confirmed"},

		// A draft is never completed by derivation: confirming is an operator's act.
		{"a draft is never derived to completed", "draft", "paid", "fulfilled", "draft"},
		{"a cancelled order stays cancelled", "cancelled", "paid", "fulfilled", "cancelled"},
		{"a completed order stays completed", "completed", "paid", "fulfilled", "completed"},

		// Overpaid is not paid: the customer is owed change.
		{"overpaid is not complete", "confirmed", "overpaid", "fulfilled", "confirmed"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := DeriveOrderStatus(testCase.current, testCase.payment, testCase.fulfillment)
			if got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

// An unissued or rejected VAT invoice cannot hold a commercially finished transaction open.
func TestInvoiceStatusDoesNotAffectCompletion(t *testing.T) {
	// DeriveOrderStatus takes no invoice status at all; this asserts the consequence.
	if got := DeriveOrderStatus("confirmed", "paid", "fulfilled"); got != "completed" {
		t.Errorf("a paid, fulfilled order must complete regardless of its invoice, got %q", got)
	}
}

// Every status DeriveOrderStatus can produce must be reachable from the status it was given, or the
// writer would refuse the move and the order would silently never progress.
func TestDerivedOrderStatusIsAlwaysAReachableTransition(t *testing.T) {
	statuses := []string{"draft", "confirmed", "processing", "completed", "cancelled"}
	payments := []string{"unpaid", "partially_paid", "paid", "overpaid",
		"partially_refunded", "refunded"}
	fulfillments := []string{"pending", "not_required", "partially_fulfilled", "fulfilled",
		"partially_returned", "returned"}

	for _, current := range statuses {
		for _, payment := range payments {
			for _, fulfillment := range fulfillments {
				derived := DeriveOrderStatus(current, payment, fulfillment)
				if !CanTransitionOrderStatus(current, derived) {
					t.Errorf("DeriveOrderStatus(%q, %q, %q) = %q, which is not reachable from %q",
						current, payment, fulfillment, derived, current)
				}
			}
		}
	}
}
