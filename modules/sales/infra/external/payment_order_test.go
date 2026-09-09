package external

import (
	"testing"

	itOrder "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/order"
)

// The status mapping is the part of this adapter that carries a judgement, so it is the part worth
// pinning: reconciliation acts on settled/failed/neither, and getting "neither" wrong either strands
// a payment as pending forever or fails one the customer is still paying.

func TestOrderStatusMapping(t *testing.T) {
	cases := []struct {
		status  string
		settled bool
		failed  bool
		why     string
	}{
		{itOrder.OrderStatusPending, false, false,
			"the customer has not paid yet and may still do so"},
		{itOrder.OrderStatusProcessing, false, false,
			"the gateway accepted the order but the money has not landed"},
		{itOrder.OrderStatusPaymentSuccess, true, false,
			"the money is in"},
		{itOrder.OrderStatusPaymentFailed, false, true,
			"the collection is over without the money arriving"},
		{itOrder.OrderStatusCanceled, false, true,
			"nobody is going to pay this"},
		{itOrder.OrderStatusExpired, false, true,
			"the window closed unpaid"},

		// A refunded order was paid. Reconciliation asking whether the payment ever completed must
		// not be told no simply because the money was later given back — Sales tracks the giving
		// back on its own refund legs.
		{itOrder.OrderStatusRefundSuccess, true, false,
			"it was paid, then refunded"},
		{itOrder.OrderStatusRefundFailed, true, false,
			"it was paid; the refund failing does not un-pay it"},
	}

	for _, testCase := range cases {
		t.Run(testCase.status, func(t *testing.T) {
			if got := isSettledOrderStatus(testCase.status); got != testCase.settled {
				t.Errorf("settled = %v, want %v (%s)", got, testCase.settled, testCase.why)
			}
			if got := isFailedOrderStatus(testCase.status); got != testCase.failed {
				t.Errorf("failed = %v, want %v (%s)", got, testCase.failed, testCase.why)
			}
		})
	}
}

// Settled and failed are mutually exclusive: a caller that saw both would have no defined behaviour,
// and the sweep would decide by whichever branch it happened to test first.
func TestOrderStatusIsNeverBothSettledAndFailed(t *testing.T) {
	statuses := []string{
		itOrder.OrderStatusPending, itOrder.OrderStatusProcessing,
		itOrder.OrderStatusPaymentSuccess, itOrder.OrderStatusPaymentFailed,
		itOrder.OrderStatusCanceled, itOrder.OrderStatusRefundSuccess,
		itOrder.OrderStatusRefundFailed, itOrder.OrderStatusExpired,
	}
	for _, status := range statuses {
		if isSettledOrderStatus(status) && isFailedOrderStatus(status) {
			t.Errorf("status %q is both settled and failed", status)
		}
	}
}

// An unknown status is neither, so a status this build has not heard of leaves the payment pending
// for a human rather than being guessed into a terminal state.
func TestUnknownOrderStatusIsNeitherSettledNorFailed(t *testing.T) {
	if isSettledOrderStatus("something_new") || isFailedOrderStatus("something_new") {
		t.Error("an unrecognised status must not be forced into a terminal answer")
	}
}
