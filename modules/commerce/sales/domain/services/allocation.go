package services

import (
	"sort"

	"github.com/shopspring/decimal"
)

// Proportional allocation with an exact residual rule, shared by combo, fixed-voucher and bill
// line allocation. Allocating 100 across three equal shares at two places gives 99.99, so the
// missing hundredth must land somewhere deterministic.

type AllocationInput struct {
	// Key is passed through untouched so results match back without relying on slice order.
	Key string

	// Zero is legitimate: a free component takes no share of the bundle price.
	Reference decimal.Decimal

	// Tiebreak orders recipients whose Reference is equal. Lowest wins.
	Tiebreak int32
}

type AllocationResult struct {
	Key    string
	Amount decimal.Decimal
}

// Allocate distributes total across the inputs proportionally to their references, rounded to the
// given scale, guaranteeing the shares sum to total EXACTLY.
//
// The whole rounding residual goes to the recipient with the largest reference (then lowest
// Tiebreak, then lowest Key), independent of insertion order. Every reference zero gives the total
// to that same recipient; a negative total is handled identically.
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

	// Proportions are undefined, so the whole total goes to the residual recipient.
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

	// Assigned in one place rather than redistributed: spreading it would re-introduce the same
	// rounding problem one level down.
	if residual := total.Sub(allocated); !residual.IsZero() {
		results[residualIndex].Amount = results[residualIndex].Amount.Add(residual)
	}
	return results
}

// residualRecipient picks who absorbs the rounding difference: largest reference, then lowest
// tiebreak, then lowest key. All three are needed to stay total on duplicate inputs.
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

func AllocationSum(results []AllocationResult) decimal.Decimal {
	total := decimal.Zero
	for _, result := range results {
		total = total.Add(result.Amount)
	}
	return total
}

// SortAllocationsByKey orders results deterministically.
func SortAllocationsByKey(results []AllocationResult) {
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})
}
