package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func TestCanTransitionTransferAllowsTheDocumentedPaths(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
		why      string
	}{
		{models.StockTransferStatusDraft, models.StockTransferStatusConfirmed, true, "confirm"},
		{models.StockTransferStatusDraft, models.StockTransferStatusWaiting, true, "confirm with an unmet dependency"},
		{models.StockTransferStatusConfirmed, models.StockTransferStatusReady, true, "reserving everything"},
		{models.StockTransferStatusReady, models.StockTransferStatusConfirmed, true, "unreserve takes readiness away"},
		{models.StockTransferStatusReady, models.StockTransferStatusDone, true, "validate"},
		{models.StockTransferStatusConfirmed, models.StockTransferStatusDone, true, "validating a partially available transfer"},
		{models.StockTransferStatusDraft, models.StockTransferStatusCancelled, true, "cancelling before confirming"},
		{models.StockTransferStatusReady, models.StockTransferStatusCancelled, true, "cancelling a reserved transfer"},

		{models.StockTransferStatusDone, models.StockTransferStatusCancelled, false, "AC-STOCK-009: a completed transfer cannot be cancelled"},
		{models.StockTransferStatusDone, models.StockTransferStatusReady, false, "done is terminal"},
		{models.StockTransferStatusCancelled, models.StockTransferStatusDraft, false, "cancelled is terminal"},
		{models.StockTransferStatusCancelled, models.StockTransferStatusDone, false, "a cancelled transfer moves nothing"},
		{models.StockTransferStatusDraft, models.StockTransferStatusDone, false, "a draft transfer cannot be validated"},
	}

	for _, tc := range cases {
		assert.Equalf(t, tc.want, CanTransitionTransfer(tc.from, tc.to), "%s -> %s: %s", tc.from, tc.to, tc.why)
	}
}

func TestCanTransitionTransferAllowsStayingPut(t *testing.T) {
	for status := range transferTransitions {
		assert.Truef(t, CanTransitionTransfer(status, status), "%s -> %s should be a no-op", status, status)
	}
}

func TestDeriveTransferStatusSummarisesItsMoves(t *testing.T) {
	assigned := models.StockMoveStatusAssigned
	partial := models.StockMoveStatusPartiallyAvailable
	confirmed := models.StockMoveStatusConfirmed
	waiting := models.StockMoveStatusWaiting
	done := models.StockMoveStatusDone
	cancelled := models.StockMoveStatusCancelled

	cases := []struct {
		name    string
		current string
		moves   []string
		want    string
	}{
		{
			"every move assigned makes the transfer ready",
			models.StockTransferStatusConfirmed, []string{assigned, assigned}, models.StockTransferStatusReady,
		},
		{
			"one partially available move is enough to keep it out of ready",
			models.StockTransferStatusConfirmed, []string{assigned, partial}, models.StockTransferStatusConfirmed,
		},
		{
			"a waiting move makes the whole transfer waiting",
			models.StockTransferStatusConfirmed, []string{assigned, waiting}, models.StockTransferStatusWaiting,
		},
		{
			"nothing reserved yet",
			models.StockTransferStatusConfirmed, []string{confirmed, confirmed}, models.StockTransferStatusConfirmed,
		},
		{
			"all moves done makes the transfer done",
			models.StockTransferStatusReady, []string{done, done}, models.StockTransferStatusDone,
		},
		{
			"done and cancelled moves together still close the transfer as done",
			models.StockTransferStatusReady, []string{done, cancelled}, models.StockTransferStatusDone,
		},
		{
			"all moves cancelled means nothing moved, so the transfer is cancelled",
			models.StockTransferStatusReady, []string{cancelled, cancelled}, models.StockTransferStatusCancelled,
		},
		{
			"closed moves are ignored when judging readiness of the rest",
			models.StockTransferStatusConfirmed, []string{done, assigned}, models.StockTransferStatusReady,
		},
	}

	for _, tc := range cases {
		assert.Equalf(t, tc.want, DeriveTransferStatus(tc.current, tc.moves), tc.name)
	}
}

func TestDeriveTransferStatusRefusesToCallAnEmptyTransferReady(t *testing.T) {
	// "every move is assigned" is vacuously true of no moves at all. Guarding the empty case is
	// what stops a transfer with nothing in it reporting itself ready to ship.
	got := DeriveTransferStatus(models.StockTransferStatusConfirmed, nil)
	assert.Equal(t, models.StockTransferStatusConfirmed, got, "an empty transfer is never ready")
}

func TestDeriveTransferStatusLeavesClosedTransfersAlone(t *testing.T) {
	for _, closed := range []string{models.StockTransferStatusDone, models.StockTransferStatusCancelled} {
		got := DeriveTransferStatus(closed, []string{models.StockMoveStatusAssigned})
		assert.Equalf(t, closed, got, "%s must not be recomputed", closed)
	}
}

func TestAssertTransferTransitionNamesTheReverseTransferRemedy(t *testing.T) {
	vErrs := ft.NewClientErrors()
	AssertTransferTransition(models.StockTransferStatusDone, models.StockTransferStatusCancelled, vErrs)

	require.Equal(t, 1, vErrs.Count(), "cancelling a done transfer must be refused")
	assert.Equal(t, "stock_transfer.done_not_cancellable", (*vErrs)[0].Key,
		"the message must point at the reverse transfer, not just say no")
}

func TestAssertTransferTransitionAcceptsALegalMove(t *testing.T) {
	vErrs := ft.NewClientErrors()
	AssertTransferTransition(models.StockTransferStatusConfirmed, models.StockTransferStatusReady, vErrs)

	assert.Equal(t, 0, vErrs.Count(), "a legal transition must raise nothing")
}
