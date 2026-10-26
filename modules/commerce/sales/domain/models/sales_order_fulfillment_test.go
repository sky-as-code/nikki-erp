package models

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// The invariants that make a fulfillment's quantities mean anything, and the state predicates every
// execution path branches on.

func fulfillmentWith(status FulfillmentStatus) SalesOrderFulfillment {
	return *NewSalesOrderFulfillmentFrom(dmodel.DynamicFields{
		SalesOrderFulfillmentFieldFulfillmentStatus: string(status),
		SalesOrderFulfillmentFieldFulfillmentType:   string(FulfillmentTypeKioskDispense),
	})
}

func itemWith(ordered, fulfilled, refunded string) SalesOrderFulfillmentItem {
	return *NewSalesOrderFulfillmentItemFrom(dmodel.DynamicFields{
		SalesOrderFulfillmentItemFieldOrderedQty:   decimal.RequireFromString(ordered),
		SalesOrderFulfillmentItemFieldFulfilledQty: decimal.RequireFromString(fulfilled),
		SalesOrderFulfillmentItemFieldRefundedQty:  decimal.RequireFromString(refunded),
	})
}

// THE invariant. Delivering and refunding the same unit would credit a customer twice for something
// they bought once — as goods and as money — so the two settled outcomes together can never exceed
// what was ordered.
func TestFulfilledAndRefundedTogetherCannotExceedOrdered(t *testing.T) {
	if _, ok := itemWith("3", "2", "2").AssertQuantitiesValid(); ok {
		t.Error("delivering 2 and refunding 2 of 3 credits the customer for 4 and must be refused")
	}
	if _, ok := itemWith("3", "2", "1").AssertQuantitiesValid(); !ok {
		t.Error("delivering 2 and refunding 1 of 3 settles the item exactly and must be allowed")
	}
}

func TestNoQuantityMayBeNegative(t *testing.T) {
	for _, quantities := range [][3]string{
		{"-1", "0", "0"},
		{"3", "-1", "0"},
		{"3", "0", "-1"},
	} {
		item := itemWith(quantities[0], quantities[1], quantities[2])
		if field, ok := item.AssertQuantitiesValid(); ok {
			t.Errorf("%v must be refused as negative", quantities)
		} else if field == "" {
			t.Errorf("%v was refused without naming the offending field", quantities)
		}
	}
}

// Remaining is what is still owed, and it never goes below zero even if the stored quantities
// disagree: a negative debt is not a credit the next caller should be able to spend.
func TestRemainingQuantityNeverGoesNegative(t *testing.T) {
	if got := itemWith("3", "2", "1").RemainingQuantity(); !got.IsZero() {
		t.Errorf("2 delivered and 1 refunded of 3 owes nothing, got %s", got)
	}
	if got := itemWith("3", "5", "0").RemainingQuantity(); !got.IsZero() {
		t.Errorf("an over-delivered item owes nothing rather than a negative, got %s", got)
	}
	if got := itemWith("3", "1", "0").RemainingQuantity(); !got.Equal(decimal.RequireFromString("2")) {
		t.Errorf("1 delivered of 3 still owes 2, got %s", got)
	}
}

// Settlement turns on remaining quantity ALONE. An item whose shortfall was refunded is settled just
// as fully as one that was delivered: the customer is owed nothing either way.
func TestAnItemIsSettledByRefundJustAsByDelivery(t *testing.T) {
	if !itemWith("2", "2", "0").IsSettled() {
		t.Error("a fully delivered item is settled")
	}
	if !itemWith("2", "0", "2").IsSettled() {
		t.Error("a fully refunded item is settled just as a delivered one is")
	}
	if itemWith("2", "1", "0").IsSettled() {
		t.Error("an item still owing goods is not settled")
	}
}

// waiting_customer_action is NOT terminal. The goods are still owed and the customer may still act;
// treating it as an ending would close orders that nobody has finished.
func TestOnlyCompletedAndCancelledAreTerminal(t *testing.T) {
	terminal := map[FulfillmentStatus]bool{
		FulfillmentStatusCompleted: true,
		FulfillmentStatusCancelled: true,
	}
	for _, status := range []FulfillmentStatus{
		FulfillmentStatusPendingReservation,
		FulfillmentStatusReserved,
		FulfillmentStatusReady,
		FulfillmentStatusInProgress,
		FulfillmentStatusPartiallyFulfilled,
		FulfillmentStatusWaitingCustomerAction,
		FulfillmentStatusExpired,
		FulfillmentStatusCompleted,
		FulfillmentStatusCancelled,
	} {
		if got := fulfillmentWith(status).IsTerminal(); got != terminal[status] {
			t.Errorf("%s: terminal = %v, want %v", status, got, terminal[status])
		}
	}
}

// THE guard against dispensing twice. A fulfillment with an attempt outstanding must not accept a
// second one: two machines acting on one reservation hand over goods that were paid for once.
func TestAnInProgressFulfillmentIsNotAttemptable(t *testing.T) {
	if fulfillmentWith(FulfillmentStatusInProgress).IsAttemptable() {
		t.Error("a fulfillment with an attempt outstanding must not accept another")
	}
	if !fulfillmentWith(FulfillmentStatusReady).IsAttemptable() {
		t.Error("a paid, reserved fulfillment must be attemptable")
	}
	if fulfillmentWith(FulfillmentStatusCancelled).IsAttemptable() {
		t.Error("a cancelled fulfillment must not be attemptable")
	}
}

// An expired fulfillment lost its stock but not its entitlement, so it must still be reservable —
// that is what lets a customer come back and choose another kiosk.
func TestAnExpiredFulfillmentIsStillReservable(t *testing.T) {
	if !fulfillmentWith(FulfillmentStatusExpired).IsReservable() {
		t.Error("expiry releases the stock, not the entitlement; reserving again must be allowed")
	}
	if fulfillmentWith(FulfillmentStatusCancelled).IsReservable() {
		t.Error("a cancelled fulfillment has no entitlement left to reserve for")
	}
}

// A fulfillment with no TTL never lapses. A sale dispensed seconds after payment would otherwise be
// expired by a sweep for having no deadline at all.
func TestAFulfillmentWithoutADeadlineNeverLapses(t *testing.T) {
	fulfillment := fulfillmentWith(FulfillmentStatusReserved)
	if fulfillment.HasLapsed(time.Now().UTC().Add(24 * time.Hour)) {
		t.Error("a fulfillment with no reservation deadline must never lapse")
	}
}

func TestAFulfillmentLapsesOnlyAfterItsDeadline(t *testing.T) {
	deadline := model.ModelDateTime(time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC))
	fulfillment := *NewSalesOrderFulfillmentFrom(dmodel.DynamicFields{
		SalesOrderFulfillmentFieldFulfillmentStatus:    string(FulfillmentStatusReserved),
		SalesOrderFulfillmentFieldReservationExpiresAt: deadline,
	})

	if fulfillment.HasLapsed(deadline.GoTime().Add(-time.Minute)) {
		t.Error("a hold must not lapse before its deadline")
	}
	if !fulfillment.HasLapsed(deadline.GoTime().Add(time.Minute)) {
		t.Error("a hold must lapse once its deadline has passed")
	}
}

// Only kiosk dispense has a workflow. Every execution path asks rather than assumes, so a method
// carrying one of the four reserved types is refused outright instead of half-running against rules
// written for a vending machine.
func TestOnlyKioskDispenseExecutes(t *testing.T) {
	if !fulfillmentWith(FulfillmentStatusReady).ExecutesKioskDispense() {
		t.Error("a kiosk fulfillment executes")
	}

	shipping := *NewSalesOrderFulfillmentFrom(dmodel.DynamicFields{
		SalesOrderFulfillmentFieldFulfillmentType: string(FulfillmentTypeCarrierShipping),
	})
	if shipping.ExecutesKioskDispense() {
		t.Error("a reserved fulfillment type must not run the kiosk workflow")
	}
}
