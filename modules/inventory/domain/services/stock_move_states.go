package services

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The move state machine. See BR §6.2.
//
// It is deliberately free of any repository dependency: every rule here is a statement about two
// strings and two numbers, so the whole table can be tested without a database. The services that
// change stock consult these functions rather than restating the rules inline, which is what keeps
// "which transitions are legal" answerable in one place.

// moveTransitions maps each state to the states reachable from it.
//
// Note what is absent: nothing leads out of `done`, and `cancelled` leads nowhere either. A move
// that has been recorded is history, and history is corrected by a reverse movement rather than by
// moving the original backwards (STOCK-INV-005).
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
	// Reservation is reversible while the move is still open, so both allocation states can fall
	// back to `confirmed` — that is what unreserve does.
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

// CanTransitionMove reports whether a move may move from one state to another.
//
// A transition to the state it is already in is allowed, because the callers recompute a state
// from allocation and write it back unconditionally; treating "no change" as illegal would make
// every idempotent recompute an error.
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

// DeriveMoveStatus computes a move's state from how much of its demand is reserved.
//
// This is the rule in BR §4.2.3.8, and it is a derivation rather than an assignment: the state is
// a reading of the allocation, so a caller that reserves stock does not also decide what to call
// the result. Partial allocation is a normal outcome, not a failure.
//
// Reserving more than demanded cannot happen through the reservation engine, which clamps to the
// outstanding quantity; if it somehow does, the move still reads as fully assigned rather than
// inventing a state for it.
//
// Full allocation is tested before zero allocation, so that a zero-demand move reads as `assigned`
// rather than `confirmed`. The other order looks equivalent and is not: it would leave such a move
// permanently un-allocatable, and one of them in a transfer would keep the transfer out of `ready`
// forever, with nothing a user could reserve to fix it.
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
