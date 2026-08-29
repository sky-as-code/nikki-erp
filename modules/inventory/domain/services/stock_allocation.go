package services

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
)

// Allocation is one decision to draw a quantity from one balance.
type Allocation struct {
	QuantId    model.Id
	Quantity   decimal.Decimal
	LotRef     string
	PackageRef string
	OwnerRef   string
}

// AllocateFromQuants decides how much to take from each locked balance to cover a wanted quantity.
// Pure: it takes the rows the lock returned and returns the split, leaving the writes to the caller.
//
// Rows are consumed in the order given, which LockQuantsForUpdate makes oldest-first, so removal is
// FIFO. A row with nothing available is skipped rather than allocated zero, since a zero-quantity
// move line records a movement that did not happen.
//
// A shortfall is a normal outcome, not an error: the move becomes partially available and the
// transfer does not become ready.
func AllocateFromQuants(wanted decimal.Decimal, quants []LockedQuant) ([]Allocation, decimal.Decimal) {
	remaining := wanted
	allocations := make([]Allocation, 0, len(quants))

	for _, quant := range quants {
		if remaining.LessThanOrEqual(decimal.Zero) {
			break
		}
		available := quant.Available()
		if available.LessThanOrEqual(decimal.Zero) {
			continue
		}

		take := available
		if take.GreaterThan(remaining) {
			take = remaining
		}
		allocations = append(allocations, Allocation{
			QuantId:    quant.Id,
			Quantity:   take,
			LotRef:     quant.LotRef,
			PackageRef: quant.PackageRef,
			OwnerRef:   quant.OwnerRef,
		})
		remaining = remaining.Sub(take)
	}

	if remaining.LessThan(decimal.Zero) {
		remaining = decimal.Zero
	}
	return allocations, remaining
}

// TotalAllocated sums a split, for comparing it against the demand it was meant to cover.
func TotalAllocated(allocations []Allocation) decimal.Decimal {
	total := decimal.Zero
	for _, allocation := range allocations {
		total = total.Add(allocation.Quantity)
	}
	return total
}

// ReleaseQuantity is the new reserved figure after giving some of a reservation back. It clamps at
// zero and reports whether it had to: a negative result means the caller's bookkeeping disagrees
// with the stored balance. The stored figure must never go negative, so the clamp holds the
// invariant and the flag lets the caller refuse rather than write a number it cannot explain.
func ReleaseQuantity(reserved, release decimal.Decimal) (decimal.Decimal, bool) {
	result := reserved.Sub(release)
	if result.LessThan(decimal.Zero) {
		return decimal.Zero, true
	}
	return result, false
}
