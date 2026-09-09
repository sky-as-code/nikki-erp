package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// paymentWithStatus builds the smallest payment the state questions need.
func paymentWithStatus(status string) *models.SalesPayment {
	return models.NewSalesPaymentFrom(dmodel.DynamicFields{
		models.SalesPaymentFieldStatus: status,
	})
}

// Applying a gateway verdict, tested where it can be tested without a database: the mapping from a
// verdict to a payment state, and the decision of whether a gateway's answer is a verdict at all.

// Every outcome lands somewhere terminal. A verdict that left a payment pending would have it swept
// again forever, asking the gateway about an order that has already answered.
func TestEveryOutcomeMapsToATerminalStatus(t *testing.T) {
	cases := []struct {
		outcome ConfirmPaymentOutcome
		want    string
		why     string
	}{
		{ConfirmPaymentPaid, string(models.SalesPaymentStatusCaptured),
			"the money is in"},
		{ConfirmPaymentFailed, string(models.SalesPaymentStatusFailed),
			"the collection is over without the money"},
		{ConfirmPaymentExpired, string(models.SalesPaymentStatusFailed),
			"an expiry is a failure with a reason; Sales has no separate state for it"},
		{ConfirmPaymentCanceled, string(models.SalesPaymentStatusCancelled),
			"called off rather than declined"},
	}

	for _, testCase := range cases {
		t.Run(string(testCase.outcome), func(t *testing.T) {
			got := paymentStatusFor(testCase.outcome)
			if got != testCase.want {
				t.Errorf("status = %q, want %q (%s)", got, testCase.want, testCase.why)
			}

			payment := paymentWithStatus(got)
			if !payment.IsTerminal() {
				t.Errorf("status %q is not terminal, so the sweep would revisit it forever", got)
			}
		})
	}
}

// Cancelled and failed are kept apart. Both free the method slot, but a declined card is the
// customer's to retry while a cancelled order is the till's, and a receipt should not confuse them.
func TestCancelledIsDistinctFromFailed(t *testing.T) {
	if paymentStatusFor(ConfirmPaymentCanceled) == paymentStatusFor(ConfirmPaymentFailed) {
		t.Error("a cancelled collection must not be recorded as a declined one")
	}
}

// Only a decided order is acted on. Reconciling a payment the customer is still making would fail a
// collection that is about to succeed.
func TestReconcileOnlyActsOnADecidedOrder(t *testing.T) {
	cases := []struct {
		name    string
		status  itExt.GatewayOrderStatus
		want    ConfirmPaymentOutcome
		decided bool
	}{
		{"settled", itExt.GatewayOrderStatus{Settled: true}, ConfirmPaymentPaid, true},
		{"failed", itExt.GatewayOrderStatus{Failed: true}, ConfirmPaymentFailed, true},
		{"still being paid", itExt.GatewayOrderStatus{}, "", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			outcome, decided := reconcileOutcomeOf(testCase.status)
			if decided != testCase.decided {
				t.Fatalf("decided = %v, want %v", decided, testCase.decided)
			}
			if decided && outcome != testCase.want {
				t.Errorf("outcome = %q, want %q", outcome, testCase.want)
			}
		})
	}
}

// A payment that has reached a verdict is terminal, and ConfirmPayment refuses to move it. This is
// what makes the announcement and the sweep safe to both deliver the same verdict — and what stops
// a late failure un-paying a customer whose goods have already gone.
func TestTerminalPaymentsAreRecognised(t *testing.T) {
	terminal := []models.SalesPaymentStatus{
		models.SalesPaymentStatusCaptured,
		models.SalesPaymentStatusFailed,
		models.SalesPaymentStatusCancelled,
	}
	for _, status := range terminal {
		value := string(status)
		if !paymentWithStatus(value).IsTerminal() {
			t.Errorf("%q must be terminal, or a replayed verdict would rewrite it", value)
		}
	}

	inFlight := []models.SalesPaymentStatus{
		models.SalesPaymentStatusPending,
		models.SalesPaymentStatusAuthorized,
	}
	for _, status := range inFlight {
		value := string(status)
		if paymentWithStatus(value).IsTerminal() {
			t.Errorf("%q must not be terminal, or its verdict could never be applied", value)
		}
	}
}
