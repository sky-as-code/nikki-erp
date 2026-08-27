package services

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The gates and the arithmetic of split and merge. Both operations write through the repository and
// are exercised live; what is pinned here is what they refuse and how they apportion.

func billRecord(id, status, currency, total string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesBillFieldId:           id,
		models.SalesBillFieldBillNumber:   "BILL-" + id,
		models.SalesBillFieldSalesOrderId: "OR1",
		models.SalesBillFieldStatus:       status,
		models.SalesBillFieldCurrencyCode: currency,
		models.SalesBillFieldTotalAmount:  total,
	}
}

func sourceAllocation(lineId, quantity, total string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesBillLineFieldSalesOrderLineId:     lineId,
		models.SalesBillLineFieldQuantity:             quantity,
		models.SalesBillLineFieldAllocatedNetAmount:   total,
		models.SalesBillLineFieldAllocatedTaxAmount:   "0",
		models.SalesBillLineFieldAllocatedTotalAmount: total,
	}
}

// A settled bill has money against it and a cancelled one is already superseded. Splitting either
// would leave a payment pointing at a bill that no longer represents what was paid.
func TestOnlyAnOpenBillCanBeSplit(t *testing.T) {
	parts := []SplitBillPart{{}, {}}

	for _, status := range []string{
		string(models.SalesBillStatusSettled),
		string(models.SalesBillStatusCancelled),
	} {
		t.Run(status, func(t *testing.T) {
			vErrs := assertSplittable(
				billRecord("B1", status, "VND", "100"),
				SplitBillParams{SourceBillId: "B1", Parts: parts})
			if vErrs == nil || !hasReasonKey(vErrs, ReasonBillNotOpen) {
				t.Errorf("a %s bill must not be splittable", status)
			}
		})
	}
}

// Splitting into one is not a split; into none would destroy the bill. Both are caller mistakes.
func TestASplitNeedsAtLeastTwoParts(t *testing.T) {
	for name, parts := range map[string][]SplitBillPart{
		"none": nil,
		"one":  {{}},
	} {
		t.Run(name, func(t *testing.T) {
			vErrs := assertSplittable(
				billRecord("B1", string(models.SalesBillStatusOpen), "VND", "100"),
				SplitBillParams{SourceBillId: "B1", Parts: parts})
			if vErrs == nil || !hasReasonKey(vErrs, ReasonSplitNeedsTwoParts) {
				t.Errorf("splitting into %s parts must be refused", name)
			}
		})
	}
}

// THE split rule. Under-allocating leaves value on no bill at all, so the customer is never asked
// for it; over-allocating bills the same goods twice. Neither is visible from one part alone.
func TestASplitMustAccountForExactlyWhatTheBillHeld(t *testing.T) {
	existing := []dmodel.DynamicFields{sourceAllocation("OL1", "10", "1000")}

	cases := []struct {
		name     string
		quantity []string
		refused  bool
	}{
		{"exact", []string{"6", "4"}, false},
		{"under", []string{"6", "3"}, true},
		{"over", []string{"6", "5"}, true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			parts := make([]SplitBillPart, 0, len(testCase.quantity))
			for _, quantity := range testCase.quantity {
				parts = append(parts, SplitBillPart{
					Allocations: map[string]decimal.Decimal{"OL1": dec(quantity)},
				})
			}

			vErrs := assertAllocationsCoverSource(existing, SplitBillParams{Parts: parts})
			if testCase.refused && vErrs == nil {
				t.Errorf("%s allocation must be refused", testCase.name)
			}
			if !testCase.refused && vErrs != nil {
				t.Errorf("%s allocation must be accepted, got %v", testCase.name, vErrs.ToError())
			}
		})
	}
}

// A part naming a line the bill does not hold is refused, not silently ignored. Ignoring it would
// let a caller believe goods were billed that never were.
func TestASplitCannotAllocateALineTheBillDoesNotHold(t *testing.T) {
	vErrs := assertAllocationsCoverSource(
		[]dmodel.DynamicFields{sourceAllocation("OL1", "10", "1000")},
		SplitBillParams{Parts: []SplitBillPart{
			{Allocations: map[string]decimal.Decimal{"OL1": dec("10")}},
			{Allocations: map[string]decimal.Decimal{"OL_OTHER": dec("1")}},
		}})

	if vErrs == nil || !hasReasonKey(vErrs, ReasonAllocationIncomplete) {
		t.Error("allocating a line the bill does not hold must be refused")
	}
}

// Splitting an amount that does not divide evenly must still sum back exactly. 1000 across three
// equal parts is 333.33 three times, which is 999.99 — and a business whose split loses a dong on
// every uneven bill has a slow leak.
func TestASplitSumsBackToTheSourceExactly(t *testing.T) {
	for _, total := range []string{"1000", "999", "100", "7"} {
		t.Run(total, func(t *testing.T) {
			inputs := []AllocationInput{
				{Key: "B1", Reference: dec("1"), Tiebreak: 0},
				{Key: "B2", Reference: dec("1"), Tiebreak: 1},
				{Key: "B3", Reference: dec("1"), Tiebreak: 2},
			}

			shares := AllocateAcrossBills(dec(total), inputs, 0)

			sum := decimal.Zero
			for _, share := range shares {
				sum = sum.Add(share)
			}
			if !sum.Equal(dec(total)) {
				t.Errorf("three-way split of %s sums to %s; the residual was lost", total, sum)
			}
		})
	}
}

// A merge needs at least two bills — merging one is a no-op the caller did not mean.
func TestAMergeNeedsAtLeastTwoBills(t *testing.T) {
	_, vErrs, err := MergeBills(nil, MergeBillParams{SourceBillIds: []string{"B1"}}, stubLock{})
	if err != nil {
		t.Fatalf("the count gate must not need a repository: %v", err)
	}
	if vErrs == nil || !hasReasonKey(vErrs, ReasonMergeNeedsTwoBills) {
		t.Error("merging a single bill must be refused")
	}
}

// Both operations refuse to run without the lock rather than proceeding unguarded. Neither is a
// single-row update, so nothing else protects them.
func TestSplitAndMergeRefuseWithoutTheLock(t *testing.T) {
	if _, _, err := SplitBill(nil, SplitBillParams{}, nil, SalesPolicy{}); err == nil {
		t.Error("split must refuse to run without the distributed lock")
	}
	if _, _, err := MergeBills(nil, MergeBillParams{
		SourceBillIds: []string{"B1", "B2"},
	}, nil); err == nil {
		t.Error("merge must refuse to run without the distributed lock")
	}
}

// A split's child bills are numbered from the parent, so a customer holding one can be matched back
// without a lookup.
func TestChildBillNumbersDeriveFromTheParent(t *testing.T) {
	source := billRecord("B1", string(models.SalesBillStatusOpen), "VND", "100")

	if got := billNumberOf(source, 0); got != "BILL-B1-1" {
		t.Errorf("first child = %q, want BILL-B1-1", got)
	}
	if got := billNumberOf(source, 1); got != "BILL-B1-2" {
		t.Errorf("second child = %q, want BILL-B1-2", got)
	}
}

// stubLock satisfies the lock interface without Redis, for the gates that run before any lock is
// taken. It never grants, so a test that reached the acquire step would stop there rather than
// proceeding against a repository that is not present.
type stubLock struct{}

var _ lock.DistributedLock = stubLock{}

func (stubLock) Acquire(context.Context, string, time.Duration) (bool, error) {
	return false, nil
}

func (stubLock) AcquireWithRetry(
	context.Context, string, time.Duration, int, time.Duration,
) (bool, error) {
	return false, nil
}

func (stubLock) Release(context.Context, string) error { return nil }
