package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The gate that makes a dispense result safe to deliver twice, and the checks that decide whether a
// report describes the attempt it claims to answer.

func attemptRecord(eventId, payloadHash, status string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesFulfillmentAttemptFieldId:                attemptTestId,
		models.SalesFulfillmentAttemptFieldResultEventId:     eventId,
		models.SalesFulfillmentAttemptFieldResultPayloadHash: payloadHash,
		models.SalesFulfillmentAttemptFieldAttemptStatus:     status,
	}
}

const attemptTestId = "AT1"

// A result arriving for the first time is applied.
func TestAFirstResultIsApplied(t *testing.T) {
	attempt := attemptRecord("", "", string(models.FulfillmentAttemptStatusPending))

	replay, vErrs := assertResultIsNew(attempt, "EV1", "hash-1")
	if vErrs != nil {
		t.Fatalf("a first result must be accepted: %v", *vErrs)
	}
	if replay {
		t.Error("a first result is not a replay")
	}
}

// THE idempotency case. The same event delivered again with the same content is the network doing
// its job: it must be answered from the record, never applied a second time, or the delivered
// quantity would be counted twice.
func TestTheSameEventWithTheSameContentIsAReplay(t *testing.T) {
	attempt := attemptRecord("EV1", "hash-1", string(models.FulfillmentAttemptStatusSucceeded))

	replay, vErrs := assertResultIsNew(attempt, "EV1", "hash-1")
	if vErrs != nil {
		t.Fatalf("a redelivery must not be refused: %v", *vErrs)
	}
	if !replay {
		t.Error("the same event with the same content must be answered as a replay")
	}
}

// THE conflict case. One physical event cannot have two outcomes, and nothing here can adjudicate
// between them, so the second claim is refused rather than guessed at.
func TestTheSameEventWithDifferentContentIsRefused(t *testing.T) {
	attempt := attemptRecord("EV1", "hash-1", string(models.FulfillmentAttemptStatusSucceeded))

	replay, vErrs := assertResultIsNew(attempt, "EV1", "hash-2")
	if replay {
		t.Error("a contradicting payload is not a replay")
	}
	if vErrs == nil || !hasReason(vErrs, ReasonResultEventConflict) {
		t.Errorf("a contradicting payload must be refused as a conflict, got %v", vErrs)
	}
}

// A different event against an attempt that already reported would double-count the goods: the
// first result is the one that happened.
func TestASecondEventOnAReportedAttemptIsRefused(t *testing.T) {
	attempt := attemptRecord("EV1", "hash-1", string(models.FulfillmentAttemptStatusSucceeded))

	_, vErrs := assertResultIsNew(attempt, "EV2", "hash-2")
	if vErrs == nil || !hasReason(vErrs, ReasonResultAlreadyReported) {
		t.Errorf("a second result must be refused, got %v", vErrs)
	}
}

// An attempt that was settled without ever reporting — a cancelled one — cannot then be reported on.
func TestASettledAttemptWithNoEventCannotReport(t *testing.T) {
	attempt := attemptRecord("", "", string(models.FulfillmentAttemptStatusCancelled))

	_, vErrs := assertResultIsNew(attempt, "EV1", "hash-1")
	if vErrs == nil || !hasReason(vErrs, ReasonResultAlreadyReported) {
		t.Errorf("a cancelled attempt is no longer awaiting a result, got %v", vErrs)
	}
}

// The payload hash must not depend on the order items arrive in: two deliveries of one event may
// legitimately order them differently, and treating that as a contradiction would refuse a replay
// that agrees in every respect that matters.
func TestThePayloadHashIgnoresItemOrder(t *testing.T) {
	forward := ApplyAttemptResultParams{
		InventoryResultRef: "INV1",
		Items: []ApplyAttemptResultItem{
			{FulfillmentItemId: "I1", DispensedQty: qtyOf("1"), FailedQty: qtyOf("0")},
			{FulfillmentItemId: "I2", DispensedQty: qtyOf("0"), FailedQty: qtyOf("2")},
		},
	}
	reversed := ApplyAttemptResultParams{
		InventoryResultRef: "INV1",
		Items: []ApplyAttemptResultItem{
			forward.Items[1], forward.Items[0],
		},
	}

	if hashResultPayload(forward) != hashResultPayload(reversed) {
		t.Error("the same result reported in another order is the same result")
	}
}

// A different quantity IS a different claim, and must hash differently or a contradiction would
// slip through as a replay.
func TestThePayloadHashDistinguishesQuantities(t *testing.T) {
	one := ApplyAttemptResultParams{
		InventoryResultRef: "INV1",
		Items: []ApplyAttemptResultItem{
			{FulfillmentItemId: "I1", DispensedQty: qtyOf("1"), FailedQty: qtyOf("1")},
		},
	}
	other := ApplyAttemptResultParams{
		InventoryResultRef: "INV1",
		Items: []ApplyAttemptResultItem{
			{FulfillmentItemId: "I1", DispensedQty: qtyOf("2"), FailedQty: qtyOf("0")},
		},
	}

	if hashResultPayload(one) == hashResultPayload(other) {
		t.Error("different quantities are different claims and must not hash alike")
	}
}

// Matching a report to its attempt.

func storedAttemptItem(itemId, fulfillmentItemId, attempted string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesFulfillmentAttemptItemFieldId:                itemId,
		models.SalesFulfillmentAttemptItemFieldFulfillmentItemId: fulfillmentItemId,
		models.SalesFulfillmentAttemptItemFieldAttemptedQty:      qtyOf(attempted),
	}
}

func TestAWellFormedReportMatches(t *testing.T) {
	stored := []dmodel.DynamicFields{storedAttemptItem("AI1", "FI1", "3")}
	reported := []ApplyAttemptResultItem{
		{FulfillmentItemId: "FI1", DispensedQty: qtyOf("2"), FailedQty: qtyOf("1")},
	}

	matched, vErrs := matchReportedItems(stored, reported)
	if vErrs != nil {
		t.Fatalf("a balanced report of every item must match: %v", *vErrs)
	}
	if len(matched) != 1 {
		t.Errorf("expected one matched item, got %d", len(matched))
	}
}

// THE quantity check. A report that does not add up to what was attempted is describing something
// other than this try.
func TestAnUnbalancedReportIsRefused(t *testing.T) {
	stored := []dmodel.DynamicFields{storedAttemptItem("AI1", "FI1", "3")}
	reported := []ApplyAttemptResultItem{
		{FulfillmentItemId: "FI1", DispensedQty: qtyOf("1"), FailedQty: qtyOf("1")},
	}

	_, vErrs := matchReportedItems(stored, reported)
	if vErrs == nil || !hasReason(vErrs, ReasonResultQuantityMismatch) {
		t.Errorf("2 of 3 accounted for must be refused, got %v", vErrs)
	}
}

// A partial report would leave items pending on a settled attempt.
func TestAReportMissingAnItemIsRefused(t *testing.T) {
	stored := []dmodel.DynamicFields{
		storedAttemptItem("AI1", "FI1", "1"),
		storedAttemptItem("AI2", "FI2", "1"),
	}
	reported := []ApplyAttemptResultItem{
		{FulfillmentItemId: "FI1", DispensedQty: qtyOf("1"), FailedQty: qtyOf("0")},
	}

	_, vErrs := matchReportedItems(stored, reported)
	if vErrs == nil {
		t.Error("a result must report every item the attempt asked for")
	}
}

// Two entries for one item could disagree, and nothing could say which was right.
func TestADuplicatedItemIsRefused(t *testing.T) {
	stored := []dmodel.DynamicFields{storedAttemptItem("AI1", "FI1", "2")}
	reported := []ApplyAttemptResultItem{
		{FulfillmentItemId: "FI1", DispensedQty: qtyOf("2"), FailedQty: qtyOf("0")},
		{FulfillmentItemId: "FI1", DispensedQty: qtyOf("0"), FailedQty: qtyOf("2")},
	}

	_, vErrs := matchReportedItems(stored, reported)
	if vErrs == nil || !hasReason(vErrs, ReasonResultItemUnknown) {
		t.Errorf("one item reported twice must be refused, got %v", vErrs)
	}
}

func TestAnItemFromAnotherAttemptIsRefused(t *testing.T) {
	stored := []dmodel.DynamicFields{storedAttemptItem("AI1", "FI1", "1")}
	reported := []ApplyAttemptResultItem{
		{FulfillmentItemId: "FI9", DispensedQty: qtyOf("1"), FailedQty: qtyOf("0")},
	}

	_, vErrs := matchReportedItems(stored, reported)
	if vErrs == nil || !hasReason(vErrs, ReasonResultItemUnknown) {
		t.Errorf("an item that was never attempted must be refused, got %v", vErrs)
	}
}

// The max_attempts policy.

func fulfillmentWithMaxAttempts(maxAttempts *int32) dmodel.DynamicFields {
	record := dmodel.DynamicFields{
		models.SalesOrderFulfillmentFieldId: "FU1",
	}
	if maxAttempts != nil {
		record[models.SalesOrderFulfillmentFieldMaxAttempts] = *maxAttempts
	}
	return record
}

// THE null case. A method with no attempt limit is never terminal by counter — the customer decides
// when to stop, and expiring their attempts by policy would take that decision away.
func TestANullMaxAttemptsIsNeverTerminal(t *testing.T) {
	if isTerminallyFailed(fulfillmentWithMaxAttempts(nil), 99) {
		t.Error("a method with no attempt limit must never be terminal by counter")
	}
}

func TestAttemptsAreTerminalOnceExhausted(t *testing.T) {
	one := int32(1)
	three := int32(3)

	if isTerminallyFailed(fulfillmentWithMaxAttempts(&one), 0) {
		t.Error("no attempts made yet is not exhausted")
	}
	if !isTerminallyFailed(fulfillmentWithMaxAttempts(&one), 1) {
		t.Error("one attempt of one allowed is exhausted")
	}
	if isTerminallyFailed(fulfillmentWithMaxAttempts(&three), 2) {
		t.Error("two attempts of three allowed leaves one")
	}
	if !isTerminallyFailed(fulfillmentWithMaxAttempts(&three), 3) {
		t.Error("three attempts of three allowed is exhausted")
	}
}

// A refund may only ever be raised for what is still owed. Delivered goods are the customer's, and
// already-refunded quantity was paid back once — both are out of the remaining quantity, which is
// why this reads that rather than summing failures across attempts, where a retried unit would be
// counted every time it failed.
func TestTerminallyFailedQuantityIsWhatIsStillOwed(t *testing.T) {
	item := dmodel.DynamicFields{
		models.SalesOrderFulfillmentItemFieldOrderedQty:   qtyOf("5"),
		models.SalesOrderFulfillmentItemFieldFulfilledQty: qtyOf("2"),
		models.SalesOrderFulfillmentItemFieldRefundedQty:  qtyOf("1"),
	}

	if got := TerminallyFailedQuantity(item); !got.Equal(qtyOf("2")) {
		t.Errorf("5 ordered, 2 delivered, 1 refunded still owes 2, got %s", got)
	}
}

func qtyOf(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}
