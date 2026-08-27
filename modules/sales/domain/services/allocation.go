package services

import (
	"sort"

	"github.com/shopspring/decimal"
)

// Proportional allocation with an exact residual rule (D-04).
//
// The same algorithm serves three requirements: combo component allocation (BR §18), fixed-voucher
// allocation across lines (BR §61), and bill line allocation (BR §36). One implementation rather
// than three, because the hard part is identical in all of them and getting it subtly different in
// each is how a total stops adding up.
//
// The hard part is the residual. Allocating 100 across three equal shares at two decimal places
// gives 33.33 three times, which is 99.99 — a hundredth has to go somewhere, and where it goes must
// be deterministic or the same order priced twice would produce different numbers.

// AllocationInput is one recipient of a share.
type AllocationInput struct {
	// Key identifies the recipient to the caller: a component id, a line id. Passed through
	// untouched so the caller can match results back without relying on slice order.
	Key string

	// Reference is what this recipient's share is proportional to — its standalone price, its net
	// amount, whatever basis the requirement names. Zero is legitimate: a free component takes no
	// share of the bundle price.
	Reference decimal.Decimal

	// Tiebreak orders recipients whose Reference is equal, so the residual lands somewhere
	// reproducible. Callers pass a line number or a sequence; D-04 breaks ties by lowest.
	Tiebreak int32
}

// AllocationResult is one recipient's share.
type AllocationResult struct {
	Key    string
	Amount decimal.Decimal
}

// Allocate distributes total across the inputs proportionally to their references, rounded to the
// given scale, guaranteeing that the shares sum to total EXACTLY.
//
// The exactness is the contract, and it is why the residual exists at all: rounding each share
// independently almost never sums back to the original, so the difference is computed and assigned
// rather than left to disappear. `Σ allocated == total` is asserted by the caller's tests and is
// what makes a combo's components reconcile against its price, and a split bill against its order.
//
// D-04 assigns the entire residual to the recipient with the LARGEST reference, breaking ties by
// lowest Tiebreak and then by Key. Largest-reference rather than "the last one" — which is what
// BR §18 says without defining "last" — because it is independent of insertion order and puts the
// sub-unit noise where it is proportionally least visible.
//
// Edge cases, all of which are real:
//   - No inputs: nothing to allocate, empty result.
//   - Every reference zero (a bundle of free items): the total is given entirely to the first
//     recipient by tiebreak order, since a proportional split of zero references is undefined and
//     dropping the total would break the sum invariant.
//   - A negative total (a discount expressed as a negative): handled identically. The proportions
//     are unaffected by sign, and the residual rule still lands the remainder deterministically.
func Allocate(
	total decimal.Decimal, inputs []AllocationInput, scale int32,
) []AllocationResult {
	if len(inputs) == 0 {
		return nil
	}

	results := make([]AllocationResult, len(inputs))
	for index, input := range inputs {
		results[index] = AllocationResult{Key: input.Key, Amount: decimal.Zero}
	}

	referenceTotal := decimal.Zero
	for _, input := range inputs {
		referenceTotal = referenceTotal.Add(input.Reference)
	}

	residualIndex := residualRecipient(inputs)

	// Every reference is zero, so proportions are undefined. The whole total goes to the recipient
	// the residual rule picks, which keeps the sum exact rather than silently losing the amount.
	if referenceTotal.IsZero() {
		results[residualIndex].Amount = total.Round(scale)
		return results
	}

	allocated := decimal.Zero
	for index, input := range inputs {
		share := total.Mul(input.Reference).Div(referenceTotal).Round(scale)
		results[index].Amount = share
		allocated = allocated.Add(share)
	}

	// Whatever rounding lost or gained, assign it in one place. It is added rather than
	// redistributed: spreading it would re-introduce the same rounding problem one level down.
	if residual := total.Sub(allocated); !residual.IsZero() {
		results[residualIndex].Amount = results[residualIndex].Amount.Add(residual)
	}
	return results
}

// residualRecipient picks who absorbs the rounding difference: largest reference, then lowest
// tiebreak, then lowest key (D-04).
//
// All three levels are needed. Reference alone is ambiguous when two components cost the same;
// tiebreak alone would ignore the requirement's basis; and the key is the last resort that makes
// the answer total even when a caller passes duplicate tiebreaks.
func residualRecipient(inputs []AllocationInput) int {
	best := 0
	for index := 1; index < len(inputs); index++ {
		candidate, incumbent := inputs[index], inputs[best]

		if comparison := candidate.Reference.Cmp(incumbent.Reference); comparison != 0 {
			if comparison > 0 {
				best = index
			}
			continue
		}
		if candidate.Tiebreak != incumbent.Tiebreak {
			if candidate.Tiebreak < incumbent.Tiebreak {
				best = index
			}
			continue
		}
		if candidate.Key < incumbent.Key {
			best = index
		}
	}
	return best
}

// AllocationSum totals a result set, for the caller asserting the invariant.
func AllocationSum(results []AllocationResult) decimal.Decimal {
	total := decimal.Zero
	for _, result := range results {
		total = total.Add(result.Amount)
	}
	return total
}

// SortAllocationsByKey orders results deterministically, for a caller that needs to compare two
// allocations or write them in a stable order.
func SortAllocationsByKey(results []AllocationResult) {
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})
}
