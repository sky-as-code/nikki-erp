package tax

import (
	"github.com/shopspring/decimal"
)

// Refund and return handling.
//
// The governing idea is that a refund reverses what was actually charged, not what would be charged
// today (BR-TAX-ESS-032). A sale made at 8% while the rate is now 10% must refund 8%, and the only
// way to know that is to read the frozen snapshot the sale left behind rather than re-running
// determination against current configuration.
//
// Tax stores none of this. The caller owns the transaction and its snapshot, and supplies both;
// Tax owns the arithmetic. That is why nothing here touches a repository — and why
// ReversalComponentInput carries how much has already been reversed rather than Tax remembering
// (BR-TAX-ESS-SUP-025).

// ReversalComponentInput is one original snapshot component being reversed against.
type ReversalComponentInput struct {
	// OriginalComponentReference identifies the component in the caller's own snapshot. It is
	// opaque to Tax, which never resolves it against a sales order.
	OriginalComponentReference string

	// OriginalReversibleBasis is the quantity or amount the original tax was charged on, and the
	// denominator of a proportional reversal.
	OriginalReversibleBasis decimal.Decimal

	// OriginalTaxAmount is what was charged, from the frozen snapshot.
	OriginalTaxAmount decimal.Decimal

	// AlreadyReversedBasis and AlreadyReversedTaxAmount are the running totals of prior partial
	// refunds. The caller owns them; Tax deliberately keeps no reversal state of its own.
	AlreadyReversedBasis     decimal.Decimal
	AlreadyReversedTaxAmount decimal.Decimal

	// RequestedReversalBasis is how much of the original basis this refund covers.
	RequestedReversalBasis decimal.Decimal

	// IsFinalReversal marks the last refund against this component, which absorbs any rounding
	// residual so the totals close exactly.
	IsFinalReversal bool
}

// ReversalComponentResult is what to reverse for one component.
type ReversalComponentResult struct {
	OriginalComponentReference string

	// ReversalTaxAmount is negative, expressing a reversal of a positive charge.
	ReversalTaxAmount decimal.Decimal

	// RemainingTaxAmount is what would still be reversible afterwards, so a caller can tell whether
	// a component is exhausted without recomputing it.
	RemainingTaxAmount decimal.Decimal
}

// ReverseFull reverses an entire original charge.
//
// No redetermination, no current-rate lookup, no recalculation from a base — the original amount is
// negated as it stands (BR-TAX-ESS-SUP-024). Recomputing would reintroduce exactly the dependency
// on today's configuration that the snapshot exists to remove.
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

// ReversePartial reverses part of an original charge, in proportion to the basis returned.
//
// The proportional rule is not applied to the final reversal. Several partial refunds each rounding
// independently will not generally sum to the original: three refunds of a third of 10.00 give
// 3.33 each, leaving a cent stranded. So the last refund takes whatever remains instead of its
// proportional share, which is what makes the invariant in BR-TAX-ESS-033 hold exactly rather than
// approximately.
//
// Every result is also clamped to what is left, so that a caller sending overlapping or duplicated
// refunds cannot reverse more tax than was charged (TAX-INV-11).
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
			// Nothing to prorate against. Reversing the whole remainder would over-refund on the
			// strength of a malformed input, so this reverses nothing and leaves the caller to
			// notice.
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
