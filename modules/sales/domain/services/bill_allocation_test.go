package services

import (
	"testing"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// BR §36's invariant and the allocator that makes it satisfiable. The repository-reading checks are
// exercised live; what is pinned here is the arithmetic, which is where "exactly" is won or lost.

func allocationRow(total string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesBillLineFieldAllocatedTotalAmount: total,
	}
}

// THE property. Three equal bills against 100 at whole-dong scale is 33.33 three times, which is
// 99.99 — and a business whose bills sum to less than its sale has a hole in it. D-04 assigns the
// residual instead of letting it vanish.
func TestAllocationsSumToTheTotalExactly(t *testing.T) {
	cases := []struct {
		name  string
		total string
		bills []string
	}{
		{"three equal bills of 100", "100", []string{"1", "1", "1"}},
		{"two bills of an odd amount", "12345", []string{"1", "2"}},
		{"seven bills", "1000", []string{"1", "1", "1", "1", "1", "1", "1"}},
		{"lopsided", "999", []string{"97", "2", "1"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			inputs := make([]AllocationInput, 0, len(testCase.bills))
			for index, quantity := range testCase.bills {
				inputs = append(inputs, AllocationInput{
					Key:       string(rune('A' + index)),
					Reference: dec(quantity),
					Tiebreak:  int32(index),
				})
			}

			shares := AllocateAcrossBills(dec(testCase.total), inputs, 0)

			sum := decimal.Zero
			for _, share := range shares {
				sum = sum.Add(share)
			}
			if !sum.Equal(dec(testCase.total)) {
				t.Errorf("shares sum to %s, want exactly %s — the residual was lost",
					sum, testCase.total)
			}
		})
	}
}

// Every bill gets an entry, including one whose share rounds to zero. A missing key would make the
// caller write no allocation row for that bill, and the bill would then have a total nothing
// accounts for.
func TestEveryBillGetsAShareEvenIfZero(t *testing.T) {
	inputs := []AllocationInput{
		{Key: "BIG", Reference: dec("1000"), Tiebreak: 1},
		{Key: "TINY", Reference: dec("1"), Tiebreak: 2},
	}

	shares := AllocateAcrossBills(dec("10"), inputs, 0)

	if len(shares) != 2 {
		t.Fatalf("both bills must receive an entry, got %d", len(shares))
	}
	if _, present := shares["TINY"]; !present {
		t.Error("the bill whose share rounds to zero must still be present")
	}
}

// The sum helper must treat an absent amount as zero rather than panicking. Allocations arrive from
// a repository and a partially written row is a real possibility.
func TestSummingToleratesMissingAmounts(t *testing.T) {
	sum := models.SumAllocatedTotal([]dmodel.DynamicFields{
		allocationRow("100"),
		{},
		allocationRow("50"),
	})

	if !sum.Equal(dec("150")) {
		t.Errorf("sum = %s, want 150 — an absent amount must count as zero", sum)
	}
}

func TestSummingNothingIsZero(t *testing.T) {
	if !models.SumAllocatedTotal(nil).IsZero() {
		t.Error("no allocations must sum to zero, not to a nil decimal")
	}
}

// The check reports the direction of the failure, because under- and over-allocation are different
// bugs: one leaves money uncollected, the other bills a customer twice for the same goods.
func TestTheCheckReportsWhichWayItIsWrong(t *testing.T) {
	under := AllocationCheck{
		OrderTotal: dec("100"), AllocatedTotal: dec("90"), Difference: dec("10"), BillCount: 1,
	}
	over := AllocationCheck{
		OrderTotal: dec("100"), AllocatedTotal: dec("110"), Difference: dec("-10"), BillCount: 2,
	}

	if under.Balances() || over.Balances() {
		t.Fatal("neither case balances")
	}
	if !under.Difference.IsPositive() {
		t.Error("under-allocation must read as a positive difference: the bills owe the order")
	}
	if !over.Difference.IsNegative() {
		t.Error("over-allocation must read as a negative difference")
	}
}

// A balanced order balances, and the zero-difference case must not be reported as a failure.
func TestABalancedOrderPasses(t *testing.T) {
	check := AllocationCheck{
		OrderTotal: dec("48000"), AllocatedTotal: dec("48000"),
		Difference: decimal.Zero, BillCount: 3,
	}
	if !check.Balances() {
		t.Error("an exactly allocated order must balance")
	}
}

// Splitting the same amount twice must give the same answer, whatever order the bills were listed
// in. Otherwise re-running a split would move a dong between bills and the invariant would hold
// while the individual bills changed underneath a customer.
func TestAllocationIsStableAcrossInputOrder(t *testing.T) {
	forward := []AllocationInput{
		{Key: "A", Reference: dec("1"), Tiebreak: 1},
		{Key: "B", Reference: dec("1"), Tiebreak: 2},
		{Key: "C", Reference: dec("1"), Tiebreak: 3},
	}
	reversed := []AllocationInput{forward[2], forward[1], forward[0]}

	first := AllocateAcrossBills(dec("100"), forward, 0)
	second := AllocateAcrossBills(dec("100"), reversed, 0)

	for key, want := range first {
		if got := second[key]; !got.Equal(want) {
			t.Errorf("bill %s got %s one way and %s the other; the split must not depend on "+
				"the order the bills were listed in", key, want, got)
		}
	}
}

// A single bill takes the whole order. The degenerate case, worth pinning because it is what the
// initial bill of an unsplit sale is.
func TestOneBillTakesEverything(t *testing.T) {
	shares := AllocateAcrossBills(dec("48000"),
		[]AllocationInput{{Key: "ONLY", Reference: dec("1"), Tiebreak: 1}}, 0)

	if !shares["ONLY"].Equal(dec("48000")) {
		t.Errorf("a single bill must take the whole amount, got %s", shares["ONLY"])
	}
}
