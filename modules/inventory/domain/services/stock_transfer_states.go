package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The transfer state machine. See BR §6.1.
//
// Like the move machine this is pure: no repository, no context, nothing but strings. The transfer
// state is a summary of its moves' states, so DeriveTransferStatus is the function that matters —
// the transition table exists to reject the states that summary should never produce.

// transferTransitions maps each state to the states reachable from it.
//
// `done` is terminal, and deliberately so: a completed transfer is the record of a movement that
// physically happened, and there is no sequence of edits that can make it not have happened. The
// correction is a reverse transfer (STOCK-INV-005, AC-STOCK-009).
var transferTransitions = map[string][]string{
	models.StockTransferStatusDraft: {
		models.StockTransferStatusWaiting,
		models.StockTransferStatusConfirmed,
		models.StockTransferStatusReady,
		models.StockTransferStatusCancelled,
	},
	models.StockTransferStatusWaiting: {
		models.StockTransferStatusConfirmed,
		models.StockTransferStatusReady,
		models.StockTransferStatusCancelled,
	},
	// Both directions between confirmed and ready are legal: reserving makes a transfer ready, and
	// unreserving takes the readiness away again.
	models.StockTransferStatusConfirmed: {
		models.StockTransferStatusWaiting,
		models.StockTransferStatusReady,
		models.StockTransferStatusDone,
		models.StockTransferStatusCancelled,
	},
	models.StockTransferStatusReady: {
		models.StockTransferStatusWaiting,
		models.StockTransferStatusConfirmed,
		models.StockTransferStatusDone,
		models.StockTransferStatusCancelled,
	},
	models.StockTransferStatusDone:      {},
	models.StockTransferStatusCancelled: {},
}

// CanTransitionTransfer reports whether a transfer may move from one state to another. As with
// moves, a transition to the current state is allowed so that an idempotent recompute is not an
// error.
func CanTransitionTransfer(from, to string) bool {
	if from == to {
		return true
	}
	for _, allowed := range transferTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// IsTransferOpen reports whether a transfer can still be worked on.
func IsTransferOpen(status string) bool {
	return status != models.StockTransferStatusDone && status != models.StockTransferStatusCancelled
}

// DeriveTransferStatus summarises a transfer's moves into the transfer's own state.
//
// The rules, in the order they are applied (BR §6.1):
//   - A transfer with no move at all is not ready. This is guarded explicitly rather than left to
//     "every move is assigned", which is vacuously true of an empty list and would report an empty
//     transfer as ready to ship.
//   - Once every move is closed, the transfer is done — unless every one of them was cancelled, in
//     which case nothing moved and the transfer is cancelled too.
//   - Ready requires every open move to be fully assigned. Under the `all_at_once` shipping policy
//     that is the whole story; under `partial` it is the same test, because a partly-allocated
//     transfer can still be validated for what it has and the readiness flag only reports full
//     allocation.
//
// It never returns `draft`: draft is the state a transfer is created in and left by confirm, not
// something an allocation can put it back into.
func DeriveTransferStatus(current string, moveStatuses []string) string {
	if !IsTransferOpen(current) {
		return current
	}
	if len(moveStatuses) == 0 {
		return current
	}

	closed, cancelled, assigned, waiting := countMoveStatuses(moveStatuses)

	if closed == len(moveStatuses) {
		if cancelled == len(moveStatuses) {
			return models.StockTransferStatusCancelled
		}
		return models.StockTransferStatusDone
	}
	open := len(moveStatuses) - closed
	if assigned == open {
		return models.StockTransferStatusReady
	}
	if waiting > 0 {
		return models.StockTransferStatusWaiting
	}
	return models.StockTransferStatusConfirmed
}

func countMoveStatuses(moveStatuses []string) (closed, cancelled, assigned, waiting int) {
	for _, status := range moveStatuses {
		switch status {
		case models.StockMoveStatusCancelled:
			closed++
			cancelled++
		case models.StockMoveStatusDone:
			closed++
		case models.StockMoveStatusAssigned:
			assigned++
		case models.StockMoveStatusWaiting:
			waiting++
		}
	}
	return closed, cancelled, assigned, waiting
}

// AssertTransferTransition refuses an illegal transfer transition, naming what was attempted.
//
// Cancelling a Done transfer gets its own message: it is the mistake a user is most likely to make
// on purpose, and "you cannot" is not an answer — telling them a reverse transfer is the way to
// undo it is (AC-STOCK-009).
func AssertTransferTransition(from, to string, vErrs *ft.ClientErrors) {
	if CanTransitionTransfer(from, to) {
		return
	}
	if from == models.StockTransferStatusDone && to == models.StockTransferStatusCancelled {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockTransferSchemaName,
			"stock_transfer.done_not_cancellable",
			"a completed transfer cannot be cancelled; record a reverse transfer to undo its movements",
		))
		return
	}
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockTransferSchemaName,
		"stock_transfer.invalid_transition",
		"a transfer cannot go from '"+from+"' to '"+to+"'",
	))
}
