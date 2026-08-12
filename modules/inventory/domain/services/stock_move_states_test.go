package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func TestCanTransitionMoveAllowsTheDocumentedPaths(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
		why      string
	}{
		{models.StockMoveStatusDraft, models.StockMoveStatusConfirmed, true, "confirming a draft move"},
		{models.StockMoveStatusDraft, models.StockMoveStatusWaiting, true, "a move whose predecessor is not done waits"},
		{models.StockMoveStatusConfirmed, models.StockMoveStatusAssigned, true, "full reservation"},
		{models.StockMoveStatusConfirmed, models.StockMoveStatusPartiallyAvailable, true, "partial reservation"},
		{models.StockMoveStatusAssigned, models.StockMoveStatusConfirmed, true, "unreserve releases the allocation"},
		{models.StockMoveStatusPartiallyAvailable, models.StockMoveStatusAssigned, true, "the rest of the stock arrives"},
		{models.StockMoveStatusAssigned, models.StockMoveStatusDone, true, "validate records the movement"},
		{models.StockMoveStatusPartiallyAvailable, models.StockMoveStatusDone, true, "validating what is available"},
		{models.StockMoveStatusConfirmed, models.StockMoveStatusCancelled, true, "cancelling an open move"},

		{models.StockMoveStatusDone, models.StockMoveStatusCancelled, false, "a recorded movement is not undone by cancelling"},
		{models.StockMoveStatusDone, models.StockMoveStatusAssigned, false, "done is terminal"},
		{models.StockMoveStatusCancelled, models.StockMoveStatusConfirmed, false, "cancelled is terminal"},
		{models.StockMoveStatusDraft, models.StockMoveStatusDone, false, "a move cannot be validated before it is confirmed"},
		{models.StockMoveStatusDraft, models.StockMoveStatusAssigned, false, "a draft move holds no reservation"},
	}

	for _, tc := range cases {
		assert.Equalf(t, tc.want, CanTransitionMove(tc.from, tc.to), "%s -> %s: %s", tc.from, tc.to, tc.why)
	}
}

func TestCanTransitionMoveAllowsStayingPut(t *testing.T) {
	// The reservation engine recomputes a state and writes it back unconditionally, so a no-op
	// recompute must not read as an illegal transition.
	for status := range moveTransitions {
		assert.Truef(t, CanTransitionMove(status, status), "%s -> %s should be a no-op", status, status)
	}
}

func TestDeriveMoveStatusReadsTheAllocation(t *testing.T) {
	demand := decimal.RequireFromString("100")

	cases := []struct {
		reserved string
		want     string
		why      string
	}{
		{"0", models.StockMoveStatusConfirmed, "nothing reserved"},
		{"40", models.StockMoveStatusPartiallyAvailable, "some reserved"},
		{"100", models.StockMoveStatusAssigned, "fully reserved"},
		{"120", models.StockMoveStatusAssigned, "over-reserved still reads as assigned"},
		{"0.000001", models.StockMoveStatusPartiallyAvailable, "the smallest representable allocation is still partial"},
	}

	for _, tc := range cases {
		got := DeriveMoveStatus(models.StockMoveStatusConfirmed, demand, decimal.RequireFromString(tc.reserved))
		assert.Equalf(t, tc.want, got, "reserved=%s: %s", tc.reserved, tc.why)
	}
}

func TestDeriveMoveStatusLeavesClosedMovesAlone(t *testing.T) {
	// A done move's allocation is spent, not outstanding. Recomputing it from quantities would
	// walk a recorded movement back to `confirmed` and make it eligible for reservation again.
	demand := decimal.RequireFromString("100")

	for _, closed := range []string{models.StockMoveStatusDone, models.StockMoveStatusCancelled} {
		got := DeriveMoveStatus(closed, demand, decimal.Zero)
		assert.Equalf(t, closed, got, "%s must not be recomputed", closed)
	}
}

func TestDeriveMoveStatusHandlesZeroDemand(t *testing.T) {
	// A zero-demand move has nothing to reserve, so it is trivially fully allocated. Reading it as
	// `confirmed` would leave it permanently un-allocatable, and one such move in a transfer would
	// keep the whole transfer out of `ready` with nothing a user could reserve to fix it.
	got := DeriveMoveStatus(models.StockMoveStatusConfirmed, decimal.Zero, decimal.Zero)
	assert.Equal(t, models.StockMoveStatusAssigned, got,
		"a zero-demand move is fully allocated by definition")
}

func TestIsMoveOpen(t *testing.T) {
	assert.True(t, IsMoveOpen(models.StockMoveStatusConfirmed))
	assert.True(t, IsMoveOpen(models.StockMoveStatusPartiallyAvailable))
	assert.False(t, IsMoveOpen(models.StockMoveStatusDone))
	assert.False(t, IsMoveOpen(models.StockMoveStatusCancelled))
}
