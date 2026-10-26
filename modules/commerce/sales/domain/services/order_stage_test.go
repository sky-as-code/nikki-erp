package services

import (
	"testing"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The stage event's shape and its version counter, pinned without a repository. A consumer orders
// and deduplicates a sale's history on these, so being wrong here is a kiosk acting on a stage the
// order left some time ago.

func stageOrderRecord(id, status string, stageVersion int32) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOrderFieldId:           id,
		models.SalesOrderFieldOrderNumber:  "SO-" + id,
		models.SalesOrderFieldStatus:       status,
		models.SalesOrderFieldStageVersion: stageVersion,
		models.SalesOrderFieldCurrencyCode: "VND",
		models.SalesOrderFieldGrandTotal:   "50000",
		models.SalesOrderFieldTaxTotal:     "0",
		basemodel.FieldOrgId:               "ORG1",
	}
}

// The version counts stages entered, so each transition is exactly one higher than the last. A
// consumer comparing two events of one order relies on this to drop the older of the pair.
func TestTheStageVersionCountsOneStageAtATime(t *testing.T) {
	if got := nextStageVersion(stageOrderRecord("OR1", "draft", 1)); got != 2 {
		t.Errorf("next stage version = %d, want 2", got)
	}
	if got := nextStageVersion(stageOrderRecord("OR1", "confirmed", 2)); got != 3 {
		t.Errorf("next stage version = %d, want 3", got)
	}
}

// An order written before the column existed reads zero and enters its next stage as 1 rather than
// as some invented count of a history nobody recorded.
func TestAnOrderWithNoStageVersionStartsAtOne(t *testing.T) {
	legacy := dmodel.DynamicFields{models.SalesOrderFieldId: "OR-OLD"}
	if got := nextStageVersion(legacy); got != 1 {
		t.Errorf("next stage version = %d, want 1", got)
	}
}

// The absence of a previous stage is null, not the empty string: a consumer telling a creation from
// a move should not have to know that "" means one of them.
func TestAnOrderEnteringItsFirstStageHasANullPreviousStage(t *testing.T) {
	if got := stageOrNil(""); got != nil {
		t.Errorf("previous stage = %v, want nil", got)
	}
	if got := stageOrNil("draft"); got != "draft" {
		t.Errorf("previous stage = %v, want draft", got)
	}
}

// The transition table is what TransitionOrderStage validates against, and it must refuse a move
// the lifecycle does not allow rather than writing it and announcing it.
func TestTheLifecycleRefusesAMoveItDoesNotAllow(t *testing.T) {
	for _, testCase := range []struct {
		from, to string
		allowed  bool
	}{
		{"draft", "confirmed", true},
		{"draft", "cancelled", true},
		{"confirmed", "completed", true},
		{"confirmed", "draft", false},
		{"completed", "confirmed", false},
		{"cancelled", "confirmed", false},
	} {
		if got := CanTransitionOrderStatus(testCase.from, testCase.to); got != testCase.allowed {
			t.Errorf("%s -> %s allowed = %v, want %v",
				testCase.from, testCase.to, got, testCase.allowed)
		}
	}
}

// Re-entering the stage already held is allowed, so an idempotent retry is not an error.
func TestAMoveToTheStageAlreadyHeldIsNotAnError(t *testing.T) {
	if !CanTransitionOrderStatus("confirmed", "confirmed") {
		t.Error("an order must be allowed to stay in the stage it is already in")
	}
}

// An archive is a filing decision, not a step in the sale, and the four independent statuses move on
// their own clocks. None of them is a stage, so none may be announced as one: a consumer told a sale
// moved on when only the money did would act far too early.
func TestPaymentAndFulfilmentStatusesAreNotOrderStages(t *testing.T) {
	for _, notAStage := range []string{"unpaid", "paid", "partially_paid", "pending", "fulfilled"} {
		if CanTransitionOrderStatus("confirmed", notAStage) {
			t.Errorf("%q is not an order stage and must not be reachable as one", notAStage)
		}
	}
}

// The retired per-action events must not come back alongside the generic one, or every confirmation
// would be announced twice.
func TestTheStageEventIsTheOnlyOrderLifecycleEvent(t *testing.T) {
	orderEvents := 0
	for _, eventType := range models.SalesEventTypes() {
		switch eventType {
		case models.EventSalesOrderStageChanged:
			orderEvents++
		case "SalesOrderConfirmed", "SalesOrderCancelled":
			t.Errorf("%q was replaced by SalesOrderStageChanged (DEC-003)", eventType)
		}
	}
	if orderEvents != 1 {
		t.Errorf("declared order-lifecycle events = %d, want exactly 1", orderEvents)
	}
}

// The confirmed event names the bill. A consumer reading it may rely on that bill already existing
// in committed state, which is only true because the two are written in one transaction.
func TestTheConfirmedStageEventCarriesTheInitialBill(t *testing.T) {
	payload := stageEventPayloadFor(OrderStageChangedParams{
		Order:         stageOrderRecord("OR1", "draft", 1),
		PreviousStage: "draft",
		CurrentStage:  "confirmed",
		StageVersion:  2,
		InitialBillId: "BILL-1",
		OccurredAt:    time.Now().UTC(),
	})

	if payload["initial_bill_id"] != "BILL-1" {
		t.Errorf("initial_bill_id = %v, want BILL-1", payload["initial_bill_id"])
	}
	if payload["previous_stage"] != "draft" || payload["current_stage"] != "confirmed" {
		t.Errorf("stages = %v -> %v, want draft -> confirmed",
			payload["previous_stage"], payload["current_stage"])
	}
	if payload["stage_version"] != int32(2) {
		t.Errorf("stage_version = %v, want 2", payload["stage_version"])
	}
}

// A transition that raised no bill must not carry an empty one: a consumer checking for the key
// would otherwise find it and read "" as a bill id.
func TestAStageEventWithNoBillOmitsTheField(t *testing.T) {
	payload := stageEventPayloadFor(OrderStageChangedParams{
		Order:         stageOrderRecord("OR1", "confirmed", 2),
		PreviousStage: "confirmed",
		CurrentStage:  "cancelled",
		StageVersion:  3,
	})

	if _, present := payload["initial_bill_id"]; present {
		t.Error("a transition that raised no bill must not carry an initial_bill_id at all")
	}
}

// The totals travel with the event so a consumer never reads back into Sales, and acts on what was
// true at the transition rather than on today's state.
func TestTheStageEventCarriesTheTotalsItAnnounces(t *testing.T) {
	payload := stageEventPayloadFor(OrderStageChangedParams{
		Order:        stageOrderRecord("OR1", "draft", 1),
		CurrentStage: "draft",
		StageVersion: 1,
	})

	if payload["order_number"] != "SO-OR1" {
		t.Errorf("order_number = %v, want SO-OR1", payload["order_number"])
	}
	if payload["currency_code"] != "VND" {
		t.Errorf("currency_code = %v, want VND", payload["currency_code"])
	}
	if payload["grand_total"] == nil {
		t.Error("the grand total must travel with the event")
	}
}

// The occurred_at is the business moment, not the publication one: a consumer ordering by when an
// event was sent would reorder events the business produced in sequence.
func TestTheStageEventStampsWhenTheTransitionHappened(t *testing.T) {
	at := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	payload := stageEventPayloadFor(OrderStageChangedParams{
		Order:        stageOrderRecord("OR1", "draft", 1),
		CurrentStage: "confirmed",
		StageVersion: 2,
		OccurredAt:   at,
	})

	if payload["occurred_at"] != at.Unix() {
		t.Errorf("occurred_at = %v, want %d", payload["occurred_at"], at.Unix())
	}
}
