package services

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The move state machine, free of any repository dependency so it tests without a database. The
// services that change stock consult these functions rather than restating the rules inline.

// moveTransitions maps each state to the states reachable from it. Nothing leads out of `done` or
// `cancelled`: a recorded move is history, corrected by a reverse movement rather than by moving
// the original backwards.
var moveTransitions = map[string][]string{
	models.StockMoveStatusDraft: {
		models.StockMoveStatusWaiting,
		models.StockMoveStatusConfirmed,
		models.StockMoveStatusCancelled,
	},
	models.StockMoveStatusWaiting: {
		models.StockMoveStatusConfirmed,
		models.StockMoveStatusPartiallyAvailable,
		models.StockMoveStatusAssigned,
		models.StockMoveStatusCancelled,
	},
	models.StockMoveStatusConfirmed: {
		models.StockMoveStatusPartiallyAvailable,
		models.StockMoveStatusAssigned,
		models.StockMoveStatusCancelled,
	},
	// Reservation is reversible while the move is open, so both allocation states can fall back to
	// `confirmed` — that is what unreserve does.
	models.StockMoveStatusPartiallyAvailable: {
		models.StockMoveStatusConfirmed,
		models.StockMoveStatusAssigned,
		models.StockMoveStatusDone,
		models.StockMoveStatusCancelled,
	},
	models.StockMoveStatusAssigned: {
		models.StockMoveStatusConfirmed,
		models.StockMoveStatusPartiallyAvailable,
		models.StockMoveStatusDone,
		models.StockMoveStatusCancelled,
	},
	models.StockMoveStatusDone:      {},
	models.StockMoveStatusCancelled: {},
}

// CanTransitionMove reports whether a move may move from one state to another. A transition to the
// current state is allowed, because callers recompute from allocation and write back
// unconditionally.
func CanTransitionMove(from, to string) bool {
	if from == to {
		return true
	}
	for _, allowed := range moveTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// IsMoveOpen reports whether a move can still be worked on. A closed move is one whose outcome is
// already recorded, so neither reservation nor validation may touch it.
func IsMoveOpen(status string) bool {
	return status != models.StockMoveStatusDone && status != models.StockMoveStatusCancelled
}

// DeriveMoveStatus computes a move's state from how much of its demand is reserved. Partial
// allocation is a normal outcome, not a failure. Over-reservation cannot happen through the
// reservation engine, which clamps to the outstanding quantity, and reads as fully assigned if it
// somehow does.
//
// Full allocation must be tested BEFORE zero allocation, so a zero-demand move reads as `assigned`
// rather than `confirmed`. The other order looks equivalent but leaves such a move permanently
// un-allocatable, keeping its transfer out of `ready` forever with nothing a user can reserve.
func DeriveMoveStatus(current string, demand, reserved decimal.Decimal) string {
	if !IsMoveOpen(current) {
		return current
	}
	switch {
	case reserved.GreaterThanOrEqual(demand):
		return models.StockMoveStatusAssigned
	case reserved.IsZero():
		return models.StockMoveStatusConfirmed
	default:
		return models.StockMoveStatusPartiallyAvailable
	}
}
