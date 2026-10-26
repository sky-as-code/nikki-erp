package pricing

import (
	"github.com/shopspring/decimal"
)

// Proportional allocation with an exact residual rule. Deliberately duplicates services.Allocate:
// this package is a leaf that imports nothing from its parent, and importing it would cycle. The two
// must stay in agreement: allocate proportionally, round each share, assign the whole residual to the
// largest reference, ties broken by lowest tiebreak then lowest key.

type allocationInput struct {
	key       string
	reference decimal.Decimal
	tiebreak  int32
}

// allocate distributes total across the inputs, guaranteeing the shares sum to total EXACTLY.
// Independently rounded shares rarely sum back, so the residual is assigned rather than dropped.
func allocate(
	total decimal.Decimal, inputs []allocationInput, scale int32,
) map[string]decimal.Decimal {
	shares := make(map[string]decimal.Decimal, len(inputs))
	if len(inputs) == 0 {
		return shares
	}

	referenceTotal := decimal.Zero
	for _, input := range inputs {
		referenceTotal = referenceTotal.Add(input.reference)
	}
	residualIndex := residualRecipient(inputs)

	// Proportions are undefined when every reference is zero; the whole total goes to the residual
	// recipient so the sum stays exact.
	if referenceTotal.IsZero() {
		for _, input := range inputs {
			shares[input.key] = decimal.Zero
		}
		shares[inputs[residualIndex].key] = total.Round(scale)
		return shares
	}

	allocated := decimal.Zero
	for _, input := range inputs {
		share := total.Mul(input.reference).Div(referenceTotal).Round(scale)
		shares[input.key] = share
		allocated = allocated.Add(share)
	}

	if residual := total.Sub(allocated); !residual.IsZero() {
		key := inputs[residualIndex].key
		shares[key] = shares[key].Add(residual)
	}
	return shares
}

// residualRecipient picks who absorbs the rounding difference: largest reference, then lowest
// tiebreak, then lowest key. All three levels are needed to stay total under duplicate references
// and duplicate tiebreaks.
func residualRecipient(inputs []allocationInput) int {
	best := 0
	for index := 1; index < len(inputs); index++ {
		candidate, incumbent := inputs[index], inputs[best]

		if comparison := candidate.reference.Cmp(incumbent.reference); comparison != 0 {
			if comparison > 0 {
				best = index
			}
			continue
		}
		if candidate.tiebreak != incumbent.tiebreak {
			if candidate.tiebreak < incumbent.tiebreak {
				best = index
			}
			continue
		}
		if candidate.key < incumbent.key {
			best = index
		}
	}
	return best
}
