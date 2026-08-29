package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The transfer state machine: pure, no repository or context. A transfer's state is a summary of
// its moves' states, so DeriveTransferStatus is the function that matters; the transition table
// exists to reject states that summary should never produce.

// transferTransitions maps each state to the states reachable from it. `done` is terminal: a
// completed transfer records a movement that physically happened, and the correction is a reverse
// transfer.
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

// CanTransitionTransfer reports whether a transfer may move from one state to another. A transition
// to the current state is allowed, so an idempotent recompute is not an error.
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

// DeriveTransferStatus summarises a transfer's moves into the transfer's own state, in this order:
//   - A transfer with no moves is not ready. Guarded explicitly, because "every move is assigned"
//     is vacuously true of an empty list and would report an empty transfer as ready to ship.
//   - Once every move is closed the transfer is done, unless every one was cancelled, in which case
//     nothing moved and the transfer is cancelled.
//   - Ready requires every open move to be fully assigned, under either shipping policy: a
//     partly-allocated transfer can still be validated for what it has.
//
// It never returns `draft`, which no allocation can put a transfer back into.
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
// Cancelling a done transfer gets its own message pointing at the reverse transfer, which is the
// actual remedy.
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
