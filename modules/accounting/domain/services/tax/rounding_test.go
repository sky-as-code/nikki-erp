package tax

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

func policy(method models.RoundingMethod, increment string) RoundingPolicy {
	return RoundingPolicy{
		Scope:     models.RoundingScopeDocument,
		Method:    method,
		Increment: dec(increment),
	}
}

func TestRoundHalfUp(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "0.01")

	assert.True(t, rounder.Round(dec("1.005")).Equal(dec("1.01")))
	assert.True(t, rounder.Round(dec("1.004")).Equal(dec("1.00")))
	// Half away from zero in both directions, so a reversal mirrors its sale exactly.
	assert.True(t, rounder.Round(dec("-1.005")).Equal(dec("-1.01")))
}

func TestRoundHalfEven(t *testing.T) {
	rounder := policy(models.RoundingHalfEven, "0.01")

	assert.True(t, rounder.Round(dec("1.005")).Equal(dec("1.00")), "ties go to the even neighbour")
	assert.True(t, rounder.Round(dec("1.015")).Equal(dec("1.02")))
}

func TestRoundUpAndDown(t *testing.T) {
	up := policy(models.RoundingUp, "0.01")
	down := policy(models.RoundingDown, "0.01")

	assert.True(t, up.Round(dec("1.001")).Equal(dec("1.01")))
	assert.True(t, down.Round(dec("1.009")).Equal(dec("1.00")))
}

// A rounding quantum is not always a power of ten: VND rounds to whole units and cash settlement to
// the nearest five, which is why the increment is authoritative rather than a decimal-place count.
func TestRoundToNonDecimalIncrement(t *testing.T) {
	toWholeDong := policy(models.RoundingHalfUp, "1")
	assert.True(t, toWholeDong.Round(dec("10499.5")).Equal(dec("10500")))

	toNearestFive := policy(models.RoundingHalfUp, "5")
	assert.True(t, toNearestFive.Round(dec("12")).Equal(dec("10")))
	assert.True(t, toNearestFive.Round(dec("13")).Equal(dec("15")))
}

// A non-positive increment must not divide by zero. Configuration validation rejects such a policy,
// so this only guards against a row written directly into the database.
func TestRoundWithoutIncrementIsIdentity(t *testing.T) {
	assert.True(t, policy(models.RoundingHalfUp, "0").Round(dec("1.2345")).Equal(dec("1.2345")))
}

// The components of a group must sum to exactly the rounded group total. Three lines of 33.335 sum
// to 100.005, rounding to 100.01, while rounding each first gives 100.02, so the allocation claws
// back one increment. The document total is what the invoice and the VAT return show.
func TestDocumentAllocationSumsToRoundedTotal(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "0.01")
	inputs := []AllocationInput{
		{LineReference: "L1", ComponentSequence: 1, GroupKey: "VAT10", Unrounded: dec("33.335")},
		{LineReference: "L2", ComponentSequence: 1, GroupKey: "VAT10", Unrounded: dec("33.335")},
		{LineReference: "L3", ComponentSequence: 1, GroupKey: "VAT10", Unrounded: dec("33.335")},
	}

	results := AllocateDocumentRounding(inputs, rounder)

	total := decimal.Zero
	for _, result := range results {
		total = total.Add(result.Rounded)
	}
	assert.True(t, total.Equal(dec("100.01")), "allocated total = %s, want 100.01", total)

	adjustments := decimal.Zero
	for _, result := range results {
		adjustments = adjustments.Add(result.Adjustment)
	}
	assert.True(t, adjustments.Equal(dec("-0.01")), "adjustments = %s", adjustments)
}

// Groups are rounded independently, because a VAT return reports per tax rather than per document.
func TestDocumentAllocationKeepsGroupsSeparate(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "0.01")
	inputs := []AllocationInput{
		{LineReference: "L1", ComponentSequence: 1, GroupKey: "VAT10", Unrounded: dec("10.005")},
		{LineReference: "L1", ComponentSequence: 2, GroupKey: "EXCISE", Unrounded: dec("20.004")},
		{LineReference: "L2", ComponentSequence: 1, GroupKey: "VAT10", Unrounded: dec("10.005")},
	}

	results := AllocateDocumentRounding(inputs, rounder)

	byGroup := map[string]decimal.Decimal{}
	for _, result := range results {
		byGroup[result.GroupKey] = byGroup[result.GroupKey].Add(result.Rounded)
	}
	// Two VAT10 components of 10.005 sum to 20.01, which is already exact at this increment.
	assert.True(t, byGroup["VAT10"].Equal(dec("20.01")), "VAT10 = %s", byGroup["VAT10"])
	assert.True(t, byGroup["EXCISE"].Equal(dec("20")), "EXCISE = %s", byGroup["EXCISE"])
}

// The allocation must be deterministic, so a later refund can reproduce which component absorbed
// the adjustment.
func TestDocumentAllocationIsDeterministic(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "0.01")
	inputs := []AllocationInput{
		{LineReference: "L3", ComponentSequence: 1, GroupKey: "VAT10", Unrounded: dec("5.004")},
		{LineReference: "L1", ComponentSequence: 2, GroupKey: "VAT10", Unrounded: dec("5.004")},
		{LineReference: "L1", ComponentSequence: 1, GroupKey: "VAT10", Unrounded: dec("5.004")},
	}

	first := AllocateDocumentRounding(inputs, rounder)
	for attempt := 0; attempt < 5; attempt++ {
		again := AllocateDocumentRounding(inputs, rounder)
		for index := range first {
			assert.True(t, first[index].Rounded.Equal(again[index].Rounded),
				"component %d moved between runs", index)
			assert.True(t, first[index].Adjustment.Equal(again[index].Adjustment),
				"adjustment %d moved between runs", index)
		}
	}
}

// When every component rounds cleanly nothing is allocated; an adjustment here would show up in a
// snapshot as an unexplained correction.
func TestDocumentAllocationLeavesExactAmountsAlone(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "0.01")
	inputs := []AllocationInput{
		{LineReference: "L1", ComponentSequence: 1, GroupKey: "VAT10", Unrounded: dec("10.00")},
		{LineReference: "L2", ComponentSequence: 1, GroupKey: "VAT10", Unrounded: dec("20.00")},
	}

	for _, result := range AllocateDocumentRounding(inputs, rounder) {
		assert.True(t, result.Adjustment.IsZero(), "unexpected adjustment on %s", result.LineReference)
	}
}

// The component that lost the most to rounding gets the increment back: L2's remainder is largest,
// so it absorbs the delta rather than an arbitrary first row.
func TestDocumentAllocationFavoursLargestRemainder(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "1")
	inputs := []AllocationInput{
		{LineReference: "L1", ComponentSequence: 1, GroupKey: "VAT", Unrounded: dec("10.1")},
		{LineReference: "L2", ComponentSequence: 1, GroupKey: "VAT", Unrounded: dec("10.4")},
	}

	// Unrounded total 20.5 rounds to 21; provisional rounding gives 10 + 10 = 20, so one unit is
	// owed to whichever component was cut hardest.
	results := AllocateDocumentRounding(inputs, rounder)

	assert.True(t, results[1].Adjustment.Equal(dec("1")), "L2 should absorb the increment")
	assert.True(t, results[0].Adjustment.IsZero())
}
