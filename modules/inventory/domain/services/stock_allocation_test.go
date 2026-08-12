package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
)

func lockedQuant(id, onHand, reserved string) LockedQuant {
	return LockedQuant{
		Id:       model.Id(id),
		OnHand:   decimal.RequireFromString(onHand),
		Reserved: decimal.RequireFromString(reserved),
	}
}

func TestAllocateFromQuantsTakesFromTheFirstRowItCan(t *testing.T) {
	quants := []LockedQuant{
		lockedQuant("q1", "100", "0"),
		lockedQuant("q2", "100", "0"),
	}

	allocations, short := AllocateFromQuants(decimal.RequireFromString("60"), quants)

	require.Len(t, allocations, 1, "one row covers the demand, so the second is untouched")
	assert.Equal(t, model.Id("q1"), allocations[0].QuantId)
	assert.True(t, allocations[0].Quantity.Equal(decimal.RequireFromString("60")))
	assert.True(t, short.IsZero(), "the demand was fully covered")
}

func TestAllocateFromQuantsSpansRowsInOrder(t *testing.T) {
	// The lock returns rows oldest first, so consuming them in order is FIFO.
	quants := []LockedQuant{
		lockedQuant("q1", "40", "0"),
		lockedQuant("q2", "100", "0"),
	}

	allocations, short := AllocateFromQuants(decimal.RequireFromString("70"), quants)

	require.Len(t, allocations, 2)
	assert.Equal(t, model.Id("q1"), allocations[0].QuantId)
	assert.True(t, allocations[0].Quantity.Equal(decimal.RequireFromString("40")), "the older row is emptied first")
	assert.True(t, allocations[1].Quantity.Equal(decimal.RequireFromString("30")), "the rest comes from the next row")
	assert.True(t, short.IsZero())
}

func TestAllocateFromQuantsAllocatesAgainstAvailableNotOnHand(t *testing.T) {
	// 100 on hand with 90 already reserved leaves 10 to give away. Allocating against on-hand
	// would promise the same stock to two demands, which is the whole failure reservation exists
	// to prevent.
	quants := []LockedQuant{lockedQuant("q1", "100", "90")}

	allocations, short := AllocateFromQuants(decimal.RequireFromString("50"), quants)

	require.Len(t, allocations, 1)
	assert.True(t, allocations[0].Quantity.Equal(decimal.RequireFromString("10")))
	assert.True(t, short.Equal(decimal.RequireFromString("40")), "the rest is a shortfall")
}

func TestAllocateFromQuantsReportsAShortfallRatherThanFailing(t *testing.T) {
	// Partial allocation is a normal outcome: the move becomes partially available and the
	// transfer simply does not become ready.
	quants := []LockedQuant{lockedQuant("q1", "5", "0")}

	allocations, short := AllocateFromQuants(decimal.RequireFromString("100"), quants)

	require.Len(t, allocations, 1)
	assert.True(t, short.Equal(decimal.RequireFromString("95")))
}

func TestAllocateFromQuantsSkipsRowsWithNothingToGive(t *testing.T) {
	// A zero-quantity move line would record a movement that never happened.
	quants := []LockedQuant{
		lockedQuant("q1", "10", "10"),
		lockedQuant("q2", "0", "0"),
		lockedQuant("q3", "7", "0"),
	}

	allocations, short := AllocateFromQuants(decimal.RequireFromString("5"), quants)

	require.Len(t, allocations, 1, "only the row with stock available is used")
	assert.Equal(t, model.Id("q3"), allocations[0].QuantId)
	assert.True(t, short.IsZero())
}

func TestAllocateFromQuantsIgnoresNegativeOnHand(t *testing.T) {
	// A negative balance is recordable — an oversell that has not been corrected yet — but it has
	// nothing to promise, so it must not produce a negative allocation.
	quants := []LockedQuant{lockedQuant("q1", "-5", "0")}

	allocations, short := AllocateFromQuants(decimal.RequireFromString("10"), quants)

	assert.Empty(t, allocations)
	assert.True(t, short.Equal(decimal.RequireFromString("10")))
}

func TestAllocateFromQuantsHandlesNoStockAtAll(t *testing.T) {
	allocations, short := AllocateFromQuants(decimal.RequireFromString("10"), nil)

	assert.Empty(t, allocations)
	assert.True(t, short.Equal(decimal.RequireFromString("10")))
}

func TestAllocateFromQuantsKeepsDecimalPrecision(t *testing.T) {
	// BR §7.3 forbids floating point for stock quantities: in float64, 0.1+0.2 != 0.3.
	quants := []LockedQuant{
		lockedQuant("q1", "0.1", "0"),
		lockedQuant("q2", "0.2", "0"),
	}

	allocations, short := AllocateFromQuants(decimal.RequireFromString("0.3"), quants)

	require.Len(t, allocations, 2)
	assert.Equal(t, "0.3", TotalAllocated(allocations).String())
	assert.True(t, short.IsZero(), "0.1 + 0.2 must cover 0.3 exactly")
}

func TestReleaseQuantityClampsAtZero(t *testing.T) {
	// STOCK-INV-002: reserved must never go negative.
	result, clamped := ReleaseQuantity(decimal.RequireFromString("10"), decimal.RequireFromString("25"))

	assert.True(t, result.IsZero())
	assert.True(t, clamped, "the caller must be told its bookkeeping disagrees with the stored balance")
}

func TestReleaseQuantitySubtractsNormally(t *testing.T) {
	result, clamped := ReleaseQuantity(decimal.RequireFromString("10"), decimal.RequireFromString("4"))

	assert.Equal(t, "6", result.String())
	assert.False(t, clamped)
}
