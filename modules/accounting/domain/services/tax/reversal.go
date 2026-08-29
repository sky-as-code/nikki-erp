package tax

import (
	"github.com/shopspring/decimal"
)

// Refund and return handling.
//
// A refund reverses what was actually charged, not what would be charged today: a sale made at 8%
// refunds 8% even if the rate is now 10%. That means reading the sale's frozen snapshot, never
// re-running determination against current configuration.
//
// Tax keeps no reversal state; the caller owns the snapshot and the already-reversed running totals
// and passes both in.

// ReversalComponentInput is one original snapshot component being reversed against.
type ReversalComponentInput struct {
	// OriginalComponentReference is opaque to Tax and never resolved against a sales order.
	OriginalComponentReference string

	// OriginalReversibleBasis is what the original tax was charged on, and the denominator of a
	// proportional reversal.
	OriginalReversibleBasis decimal.Decimal

	// OriginalTaxAmount is what was charged, from the frozen snapshot.
	OriginalTaxAmount decimal.Decimal

	// AlreadyReversedBasis and AlreadyReversedTaxAmount are the caller-owned running totals of prior
	// partial refunds; Tax keeps no state of its own.
	AlreadyReversedBasis     decimal.Decimal
	AlreadyReversedTaxAmount decimal.Decimal

	// RequestedReversalBasis is how much of the original basis this refund covers.
	RequestedReversalBasis decimal.Decimal

	// IsFinalReversal marks the last refund against this component; it absorbs any rounding residual
	// so the totals close exactly.
	IsFinalReversal bool
}

// ReversalComponentResult is what to reverse for one component.
type ReversalComponentResult struct {
	OriginalComponentReference string

	// ReversalTaxAmount is negative: a reversal of a positive charge.
	ReversalTaxAmount decimal.Decimal

	// RemainingTaxAmount is what stays reversible afterwards, so a caller can tell a component is
	// exhausted without recomputing.
	RemainingTaxAmount decimal.Decimal
}

// ReverseFull negates the original snapshot amount as it stands: no redetermination, no
// current-rate lookup, no recalculation from a base, since recomputing would reintroduce the
// dependency on today's configuration that the snapshot exists to remove.
func ReverseFull(components []ReversalComponentInput) []ReversalComponentResult {
	results := make([]ReversalComponentResult, 0, len(components))
	for _, component := range components {
		remaining := component.OriginalTaxAmount.Sub(component.AlreadyReversedTaxAmount)
		results = append(results, ReversalComponentResult{
			OriginalComponentReference: component.OriginalComponentReference,
			ReversalTaxAmount:          remaining.Neg(),
			RemainingTaxAmount:         decimal.Zero,
		})
	}
	return results
}

// ReversePartial reverses part of an original charge in proportion to the basis returned, rounding
// each share by the policy. The final reversal takes whatever remains instead of its proportional
// share: independently rounded partials do not generally sum to the original (three thirds of 10.00
// give 3.33 each, stranding a cent).
//
// Results are clamped to what is left, so overlapping or duplicated refunds cannot reverse more tax
// than was charged. Returned amounts are negative.
func ReversePartial(
	components []ReversalComponentInput, policy RoundingPolicy,
) []ReversalComponentResult {
	results := make([]ReversalComponentResult, 0, len(components))

	for _, component := range components {
		remaining := component.OriginalTaxAmount.Sub(component.AlreadyReversedTaxAmount)

		var reversal decimal.Decimal
		switch {
		case component.IsFinalReversal:
			// The final reversal absorbs the residual by definition.
			reversal = remaining

		case !component.OriginalReversibleBasis.IsPositive():
			// Nothing to prorate against; reversing the remainder would over-refund on malformed
			// input, so reverse nothing and let the caller notice.
			reversal = decimal.Zero

		default:
			proportional := component.OriginalTaxAmount.
				Mul(component.RequestedReversalBasis).
				DivRound(component.OriginalReversibleBasis, workingScale)
			reversal = policy.Round(proportional)
		}

		// Clamp, so no sequence of partial refunds can exceed the original charge.
		if reversal.GreaterThan(remaining) {
			reversal = remaining
		}
		if reversal.IsNegative() {
			reversal = decimal.Zero
		}

		results = append(results, ReversalComponentResult{
			OriginalComponentReference: component.OriginalComponentReference,
			ReversalTaxAmount:          reversal.Neg(),
			RemainingTaxAmount:         remaining.Sub(reversal),
		})
	}
	return results
}
