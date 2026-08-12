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
//
// It is deliberately pure: it takes the rows the lock returned and returns the split, touching
// nothing. That keeps the interesting arithmetic — partial coverage, exhausted rows, rounding —
// testable without a database, and leaves the caller responsible for the writes.
//
// The rows are consumed in the order given, which LockQuantsForUpdate makes oldest-first, so the
// default removal strategy is FIFO. A row with nothing available is skipped rather than allocated
// zero, because a zero-quantity move line records a movement that did not happen.
//
// Returns the allocations and the quantity it could not cover. A shortfall is a normal outcome, not
// an error: the move becomes partially available and the transfer does not become ready.
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

// ReleaseQuantity is the new reserved figure after giving some of a reservation back.
//
// It clamps at zero and reports whether it had to. A reservation can only be released by whoever
// made it, so a negative result means the caller's bookkeeping disagrees with the stored balance —
// STOCK-INV-002 says the stored figure must never go negative, so the clamp holds the invariant and
// the flag lets the caller refuse the operation rather than write a number it cannot explain.
func ReleaseQuantity(reserved, release decimal.Decimal) (decimal.Decimal, bool) {
	result := reserved.Sub(release)
	if result.LessThan(decimal.Zero) {
		return decimal.Zero, true
	}
	return result, false
}
