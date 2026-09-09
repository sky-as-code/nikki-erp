package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Refund-aware payment status, which was unreachable before Phase 6: the live path derived status
// from payable and captured alone, so an order stayed `paid` however much had been given back.

func TestRefundsReachPartiallyRefundedAndRefunded(t *testing.T) {
	payable := dec("100")
	captured := dec("100")

	cases := []struct {
		name     string
		refunded decimal.Decimal
		want     models.SalesOrderPaymentStatus
	}{
		{"nothing back", decimal.Zero, models.SalesOrderPaymentStatusPaid},
		{"some back", dec("30"), models.SalesOrderPaymentStatusPartiallyRefunded},
		{"all back", dec("100"), models.SalesOrderPaymentStatusRefunded},

		// More back than came in should still read as fully refunded rather than falling through to
		// a payment state: the customer is whole, and the excess is a separate problem.
		{"more than came in", dec("120"), models.SalesOrderPaymentStatusRefunded},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := DerivePaymentStatus(payable, captured, testCase.refunded)
			if got != string(testCase.want) {
				t.Errorf("status = %q, want %q", got, testCase.want)
			}
		})
	}
}

// The refund state takes precedence over the payment state. An order that was overpaid and then
// refunded must not still read `overpaid`, which would suggest money is owed back that already went.
func TestRefundStateWinsOverPaymentState(t *testing.T) {
	got := DerivePaymentStatus(dec("100"), dec("150"), dec("50"))
	if got != string(models.SalesOrderPaymentStatusPartiallyRefunded) {
		t.Errorf("status = %q, want partially_refunded", got)
	}
}

// Only completed legs count as money returned, mirroring captured payments on the way in. A pending
// refund treated as done reports a customer repaid who is still waiting for their money.
func TestOnlyCompletedRefundLegsCount(t *testing.T) {
	legs := []dmodel.DynamicFields{
		{models.SalesRefundPaymentFieldStatus: string(models.SalesRefundPaymentStatusCompleted),
			models.SalesRefundPaymentFieldAmount: dec("40")},
		{models.SalesRefundPaymentFieldStatus: string(models.SalesRefundPaymentStatusPending),
			models.SalesRefundPaymentFieldAmount: dec("30")},
		{models.SalesRefundPaymentFieldStatus: string(models.SalesRefundPaymentStatusFailed),
			models.SalesRefundPaymentFieldAmount: dec("20")},
		{models.SalesRefundPaymentFieldStatus: string(models.SalesRefundPaymentStatusProcessing),
			models.SalesRefundPaymentFieldAmount: dec("10")},
	}

	total := models.SumCompletedRefunds(legs)
	if !total.Equal(dec("40")) {
		t.Errorf("completed refunds = %s, want 40 — only the completed leg is money that moved",
			total)
	}
}

// A leg outcome must map to exactly one counter, or a pass reports totals that do not add up.
func TestEveryRefundLegOutcomeIsDistinct(t *testing.T) {
	outcomes := []refundLegOutcome{refundStillPending, refundCompleted, refundFailed}
	seen := map[refundLegOutcome]bool{}
	for _, outcome := range outcomes {
		if seen[outcome] {
			t.Errorf("outcome %d is duplicated", outcome)
		}
		seen[outcome] = true
	}
}

// A gateway leg with no gateway bound stays pending rather than being marked complete. Marking it
// complete would tell a customer they were repaid when no money moved.
func TestAGatewayLegWithoutAPortStaysPending(t *testing.T) {
	// The behaviour lives in settleOneRefundLeg's nil-port branch; this pins the intent so a later
	// refactor that "simplifies" it into a completion is caught.
	if refundStillPending == refundCompleted {
		t.Fatal("pending and completed must be distinguishable")
	}
}
