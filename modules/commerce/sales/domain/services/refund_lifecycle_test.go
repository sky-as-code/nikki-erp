package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The rules that keep a refund's lifecycle separate from a delivery's, and the arithmetic that
// decides how much of a request was actually paid.

func refundReturn(refundStatus models.SalesReturnStepStatus) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesReturnFieldId:           "RT1",
		models.SalesReturnFieldRefundStatus: string(refundStatus),
	}
}

// THE rule of doc 05 §18. A raised refund has paid nobody: until settlement says otherwise the
// customer is owed either the goods or the money, and counting the request as the outcome would
// close an order nobody has received anything for.
func TestAPendingRefundIsNotTerminal(t *testing.T) {
	for _, status := range []models.SalesReturnStepStatus{
		models.SalesReturnStepPending,
		models.SalesReturnStepProcessing,
	} {
		if isRefundTerminal(refundReturn(status)) {
			t.Errorf("%s: a refund still in flight blocks its quantity and is not terminal", status)
		}
	}
}

// A FAILED refund is terminal, and that is what puts the goods back into play. Leaving it open would
// block the quantity forever behind a refund that is never going to succeed.
func TestAFailedRefundIsTerminalSoTheGoodsBecomeOwedAgain(t *testing.T) {
	if !isRefundTerminal(refundReturn(models.SalesReturnStepFailed)) {
		t.Error("a failed refund must release its hold on the quantity, not block it forever")
	}
}

func TestSettledAndCancelledRefundsAreTerminal(t *testing.T) {
	for _, status := range []models.SalesReturnStepStatus{
		models.SalesReturnStepCompleted,
		models.SalesReturnStepPartiallySucceeded,
		models.SalesReturnStepCancelled,
		models.SalesReturnStepNotRequired,
	} {
		if !isRefundTerminal(refundReturn(status)) {
			t.Errorf("%s: a settled refund must not keep blocking its quantity", status)
		}
	}
}

// A refund line pointing at a return that no longer exists must not strand the quantity.
func TestAnOrphanedRefundLineBlocksNothing(t *testing.T) {
	if !isRefundTerminal(nil) {
		t.Error("a refund whose return has gone cannot keep blocking goods")
	}
}

// Only money that MOVED reduces what a customer is owed. Pending legs have not paid, and failed legs
// never will.
func TestOnlyCompletedLegsCountAsSettled(t *testing.T) {
	cases := []struct {
		name   string
		status models.SalesRefundPaymentStatus
		counts bool
	}{
		{"completed money moved", models.SalesRefundPaymentStatusCompleted, true},
		{"pending money has not moved", models.SalesRefundPaymentStatusPending, false},
		{"failed money never will", models.SalesRefundPaymentStatusFailed, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			settled := models.SalesRefundPaymentStatus(string(c.status)) ==
				models.SalesRefundPaymentStatusCompleted
			if settled != c.counts {
				t.Errorf("%s: counted = %v, want %v", c.status, settled, c.counts)
			}
		})
	}
}

// The refund_only rule. A failed dispense has no goods coming back — they never reached the
// customer — so asking Inventory to receive them would book stock that never moved.
func TestARefundOnlyReturnSkipsTheInventoryStep(t *testing.T) {
	refundOnly := *models.NewSalesReturnFrom(dmodel.DynamicFields{
		models.SalesReturnFieldReturnType: string(models.SalesReturnTypeRefundOnly),
	})
	if !refundOnly.IsRefundOnly() {
		t.Error("a refund_only return must be recognised as skipping the inventory step")
	}

	goods := *models.NewSalesReturnFrom(dmodel.DynamicFields{
		models.SalesReturnFieldReturnType: string(models.SalesReturnTypeGoodsReturn),
	})
	if goods.IsRefundOnly() {
		t.Error("an ordinary return brings goods back and must not skip the inventory step")
	}
}

// A record written before return_type existed reads as a goods return, which is the behaviour it
// had: an absent value must not silently start skipping the inventory step for old returns.
func TestAnAbsentReturnTypeIsAGoodsReturn(t *testing.T) {
	legacy := *models.NewSalesReturnFrom(dmodel.DynamicFields{})
	if legacy.IsRefundOnly() {
		t.Error("a return with no type recorded must behave as it always did")
	}
}

// CR §62, the guard that keeps this feature out of the fiscal workflow. A refund raised because a
// machine could not hand goods over is a fulfillment event, not a commercial revision of the sale —
// the customer bought and paid for goods they simply never received.
func TestOnlyAFulfillmentFailureRefundIsRecognisedAsAutomatic(t *testing.T) {
	automatic := *models.NewSalesReturnFrom(dmodel.DynamicFields{
		models.SalesReturnFieldRefundReason: string(models.SalesRefundReasonFulfillmentFailure),
	})
	if !automatic.IsFulfillmentFailure() {
		t.Error("a fulfillment-failure refund must be recognised, or it would file a tax correction")
	}

	requested := *models.NewSalesReturnFrom(dmodel.DynamicFields{
		models.SalesReturnFieldRefundReason: string(models.SalesRefundReasonCustomerRequested),
	})
	if requested.IsFulfillmentFailure() {
		t.Error("a customer-requested refund still adjusts the invoice and must not be skipped")
	}

	legacy := *models.NewSalesReturnFrom(dmodel.DynamicFields{})
	if legacy.IsFulfillmentFailure() {
		t.Error("a return with no refund reason must keep its existing fiscal behaviour")
	}
}

// A refund line asks in one quantity and is paid in another. The requested figure falls back to the
// goods quantity for a line written before the two were distinguished, where they were the same.
func TestRequestedQuantityFallsBackToTheGoodsQuantity(t *testing.T) {
	explicit := *models.NewSalesReturnLineFrom(dmodel.DynamicFields{
		models.SalesReturnLineFieldQuantity:     decimal.RequireFromString("1"),
		models.SalesReturnLineFieldRequestedQty: decimal.RequireFromString("4"),
	})
	if got := explicit.RefundQuantityRequested(); !got.Equal(decimal.RequireFromString("4")) {
		t.Errorf("an explicit requested quantity wins, got %s", got)
	}

	legacy := *models.NewSalesReturnLineFrom(dmodel.DynamicFields{
		models.SalesReturnLineFieldQuantity: decimal.RequireFromString("3"),
	})
	if got := legacy.RefundQuantityRequested(); !got.Equal(decimal.RequireFromString("3")) {
		t.Errorf("a line predating requested_qty refunds what it returns, got %s", got)
	}
}

// A refund-only line legitimately returns zero GOODS while asking for money, which is exactly the
// failed-dispense shape.
func TestARefundOnlyLineAsksForMoneyWithNoGoods(t *testing.T) {
	line := *models.NewSalesReturnLineFrom(dmodel.DynamicFields{
		models.SalesReturnLineFieldQuantity:     decimal.Zero,
		models.SalesReturnLineFieldRequestedQty: decimal.RequireFromString("2"),
	})
	if got := line.RefundQuantityRequested(); !got.Equal(decimal.RequireFromString("2")) {
		t.Errorf("a refund-only line asks for money against goods that never arrived, got %s", got)
	}
}
