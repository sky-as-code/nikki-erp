package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func TestCountVarianceIsCountedMinusSystem(t *testing.T) {
	// Positive means the shelf held more than the books said, so the adjustment gains stock.
	assert.Equal(t, "3", CountVariance(dec(t, "103"), dec(t, "100")).String())
	// Negative means it held less, and the adjustment loses stock. The sign is what tells the
	// caller which way the movement goes, so it must survive.
	assert.Equal(t, "-3", CountVariance(dec(t, "97"), dec(t, "100")).String())
	assert.True(t, CountVariance(dec(t, "100"), dec(t, "100")).IsZero())
}

func TestSnapshotIsStaleOnlyWhenTheBalanceMoved(t *testing.T) {
	assert.False(t, IsCountSnapshotStale(dec(t, "100"), dec(t, "100")))
	assert.True(t, IsCountSnapshotStale(dec(t, "100"), dec(t, "90")))
}

func TestSnapshotStalenessCatchesTheWorkedExample(t *testing.T) {
	// System 100, counter finds 97, a delivery of 10 lands before apply. Applying the -3 variance
	// against 90 would write 87 when the shelf actually holds 107.
	snapshot, afterDelivery := dec(t, "100"), dec(t, "110")

	assert.True(t, IsCountSnapshotStale(snapshot, afterDelivery))
}

func TestSnapshotComparisonIgnoresTrailingZeroes(t *testing.T) {
	// Decimal equality, not string equality: the same quantity stored with a different scale is
	// the same balance, and refusing it would make apply fail at random depending on how the
	// number happened to be written.
	assert.False(t, IsCountSnapshotStale(dec(t, "100.000000"), dec(t, "100")))
}

func TestEnteringACountAcceptsZero(t *testing.T) {
	// "The shelf is empty" is a legitimate count result, and is exactly why count_quantity_set is
	// a separate flag from the value.
	assert.Equal(t, 0, AssertCountEnterable(dec(t, "0")).Count())
}

func TestEnteringACountRefusesANegative(t *testing.T) {
	assert.Equal(t, 1, AssertCountEnterable(dec(t, "-1")).Count())
}

func TestApplyNeedsThePendingFlagNotTheValue(t *testing.T) {
	// The flag is the authority. Testing counted_quantity != 0 as a proxy would refuse every count
	// that found an empty shelf — the case a counter most needs recorded.
	assert.Equal(t, 1, AssertCountApplicable(false).Count())
	assert.Equal(t, 0, AssertCountApplicable(true).Count())
}

func TestCountEntryFieldsSnapshotTheBalanceAndSetTheFlag(t *testing.T) {
	fields := CountEntryFields(dec(t, "97"), dec(t, "100"), "missing", "two boxes short")

	assert.Equal(t, "97", fields[models.StockQuantFieldCountedQuantity])
	assert.Equal(t, "100", fields[models.StockQuantFieldCountSnapshotQty])
	assert.Equal(t, true, fields[models.StockQuantFieldCountQuantitySet])
	assert.Equal(t, "missing", fields[models.StockQuantFieldCountReasonCode])
}

func TestCountEntryFieldsNeverTouchTheBalance(t *testing.T) {
	// Entering a count must not change on-hand, held by the field never appearing in the update.
	fields := CountEntryFields(dec(t, "97"), dec(t, "100"), "", "")

	assert.NotContains(t, fields, models.StockQuantFieldOnHandQuantity)
	assert.NotContains(t, fields, models.StockQuantFieldReservedQuantity)
}

func TestResetClearsTheCountAndLeavesTheBalanceAlone(t *testing.T) {
	fields := CountResetFields()

	assert.Nil(t, fields[models.StockQuantFieldCountedQuantity])
	assert.Equal(t, false, fields[models.StockQuantFieldCountQuantitySet])
	assert.NotContains(t, fields, models.StockQuantFieldOnHandQuantity)
}

func TestResetLeavesTheSchedulingFieldsAlone(t *testing.T) {
	// A balance whose count was reset is still due to be counted, so clearing next_count_date here
	// would quietly drop it off the worklist.
	fields := CountResetFields()

	assert.NotContains(t, fields, models.StockQuantFieldNextCountDate)
	assert.NotContains(t, fields, models.StockQuantFieldLastCountDate)
}

func TestApplyStampsTheHistoryAndClearsTheCount(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)

	fields := CountAppliedFields(now, nil)

	assert.Equal(t, now, fields[models.StockQuantFieldLastCountDate])
	assert.Equal(t, false, fields[models.StockQuantFieldCountQuantitySet])
	assert.Nil(t, fields[models.StockQuantFieldCountedQuantity])
}

func TestApplyRollsTheNextCountDateWhenOneIsGiven(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	next := now.AddDate(0, 3, 0)

	withNext := CountAppliedFields(now, &next)
	withoutNext := CountAppliedFields(now, nil)

	assert.Equal(t, next, withNext[models.StockQuantFieldNextCountDate])
	// Absent rather than nil: writing a null would clear a schedule the caller never asked to
	// change.
	assert.NotContains(t, withoutNext, models.StockQuantFieldNextCountDate)
}
