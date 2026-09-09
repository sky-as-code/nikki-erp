package models

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The derivation rules that stop a reporter asserting an outcome its own numbers contradict, and the
// balance invariant that decides whether a report describes the attempt it claims to answer.

func qty(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}

func attemptItem(attempted, dispensed, failed string) SalesFulfillmentAttemptItem {
	return *NewSalesFulfillmentAttemptItemFrom(dmodel.DynamicFields{
		SalesFulfillmentAttemptItemFieldAttemptedQty: qty(attempted),
		SalesFulfillmentAttemptItemFieldDispensedQty: qty(dispensed),
		SalesFulfillmentAttemptItemFieldFailedQty:    qty(failed),
	})
}

// An item's outcome is read off its quantities, never taken on trust. A reporter able to claim
// success alongside a shortfall would settle a sale that still owes goods.
func TestItemResultIsDerivedFromQuantities(t *testing.T) {
	cases := []struct {
		dispensed, failed string
		want              FulfillmentAttemptItemResult
	}{
		{"2", "0", FulfillmentAttemptItemResultSuccess},
		{"0", "2", FulfillmentAttemptItemResultFailure},
		{"1", "1", FulfillmentAttemptItemResultPartial},
		{"0", "0", FulfillmentAttemptItemResultPending},
	}
	for _, c := range cases {
		if got := DeriveItemResult(qty(c.dispensed), qty(c.failed)); got != c.want {
			t.Errorf("dispensed %s failed %s: got %s, want %s", c.dispensed, c.failed, got, c.want)
		}
	}
}

// THE balance invariant. A report where the parts do not add up to what was attempted describes
// quantity that vanished or appeared from nowhere, and nothing can say which of the three numbers
// lied — so it is refused rather than reconciled.
func TestAnAttemptItemMustBalance(t *testing.T) {
	if !attemptItem("3", "2", "1").AssertQuantitiesBalance() {
		t.Error("2 dispensed and 1 failed of 3 attempted balances and must be accepted")
	}
	if attemptItem("3", "2", "0").AssertQuantitiesBalance() {
		t.Error("2 dispensed and 0 failed of 3 attempted loses a unit and must be refused")
	}
	if attemptItem("3", "3", "1").AssertQuantitiesBalance() {
		t.Error("3 dispensed and 1 failed of 3 attempted invents a unit and must be refused")
	}
	if attemptItem("3", "-1", "4").AssertQuantitiesBalance() {
		t.Error("a negative quantity must be refused even when the arithmetic happens to balance")
	}
}

// A mixed outcome is partially_succeeded, not failed. Calling it a failure would invite a retry for
// quantity the customer is already holding.
func TestAttemptStatusIsDerivedFromItsItems(t *testing.T) {
	cases := []struct {
		name    string
		results []FulfillmentAttemptItemResult
		want    FulfillmentAttemptStatus
	}{
		{
			"everything came out",
			[]FulfillmentAttemptItemResult{FulfillmentAttemptItemResultSuccess, FulfillmentAttemptItemResultSuccess},
			FulfillmentAttemptStatusSucceeded,
		},
		{
			"nothing came out",
			[]FulfillmentAttemptItemResult{FulfillmentAttemptItemResultFailure, FulfillmentAttemptItemResultFailure},
			FulfillmentAttemptStatusFailed,
		},
		{
			"one item delivered, one did not",
			[]FulfillmentAttemptItemResult{FulfillmentAttemptItemResultSuccess, FulfillmentAttemptItemResultFailure},
			FulfillmentAttemptStatusPartiallySucceeded,
		},
		{
			"one item partly delivered",
			[]FulfillmentAttemptItemResult{FulfillmentAttemptItemResultPartial},
			FulfillmentAttemptStatusPartiallySucceeded,
		},
		{
			"nothing reported at all",
			nil,
			FulfillmentAttemptStatusPending,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DeriveAttemptStatus(c.results); got != c.want {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}

// An unanswered item leaves the whole attempt pending. Reporting it as succeeded would close a try
// that is still owed an answer, and release the single-flight slot while a machine may still act.
func TestAnUnansweredItemKeepsTheAttemptPending(t *testing.T) {
	results := []FulfillmentAttemptItemResult{
		FulfillmentAttemptItemResultSuccess,
		FulfillmentAttemptItemResultPending,
	}
	if got := DeriveAttemptStatus(results); got != FulfillmentAttemptStatusPending {
		t.Errorf("an attempt with an unanswered item is not finished, got %s", got)
	}
}

// THE single-flight guard. Two executors acting on one reservation hand over goods that were paid
// for once, so an outstanding attempt is what blocks the next.
func TestOnlyAPendingAttemptIsOutstanding(t *testing.T) {
	outstanding := map[FulfillmentAttemptStatus]bool{FulfillmentAttemptStatusPending: true}
	for _, status := range []FulfillmentAttemptStatus{
		FulfillmentAttemptStatusPending,
		FulfillmentAttemptStatusSucceeded,
		FulfillmentAttemptStatusPartiallySucceeded,
		FulfillmentAttemptStatusFailed,
		FulfillmentAttemptStatusCancelled,
	} {
		attempt := *NewSalesFulfillmentAttemptFrom(dmodel.DynamicFields{
			SalesFulfillmentAttemptFieldAttemptStatus: string(status),
		})
		if got := attempt.IsOutstanding(); got != outstanding[status] {
			t.Errorf("%s: outstanding = %v, want %v", status, got, outstanding[status])
		}
	}
}

// HasResult is what the idempotency gate asks first: an attempt carrying an event id has already
// reported, and an arriving event is either its redelivery or a second claim about one event.
func TestHasResultTurnsOnTheStoredEventId(t *testing.T) {
	fresh := *NewSalesFulfillmentAttemptFrom(dmodel.DynamicFields{})
	if fresh.HasResult() {
		t.Error("an attempt with no stored event id has not reported")
	}

	reported := *NewSalesFulfillmentAttemptFrom(dmodel.DynamicFields{
		SalesFulfillmentAttemptFieldResultEventId: "EV1",
	})
	if !reported.HasResult() {
		t.Error("an attempt carrying an event id has reported")
	}
}
