package pricing

import (
	"github.com/shopspring/decimal"
)

// Proportional allocation with an exact residual rule (D-04).
//
// This duplicates services.Allocate deliberately. The engine is a leaf package that imports nothing
// from its parent — that is what keeps it pure and independently testable (D-13) — and importing the
// parent would create a cycle the moment the parent calls the engine, which it will.
//
// The two must agree, and both are tested against the same D-04 rule: allocate proportionally,
// round each share, assign the entire residual to the largest reference, ties broken by lowest
// tiebreak then lowest key.

type allocationInput struct {
	key       string
	reference decimal.Decimal
	tiebreak  int32
}

// allocate distributes total across the inputs, guaranteeing the shares sum to total EXACTLY.
//
// The exactness is the contract. Rounding each share independently almost never sums back to the
// original, so the difference is computed and assigned rather than left to vanish — which is what
// makes a combo's components reconcile against its price and a discount against its lines.
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

	// Every reference is zero, so proportions are undefined. The whole total goes to the recipient
	// the residual rule picks, which keeps the sum exact rather than losing the amount.
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
// tiebreak, then lowest key (D-04).
//
// All three levels are needed. Reference alone ties when two components cost the same; the tiebreak
// alone would ignore the requirement's stated basis; the key is the last resort that makes the
// answer total even when a caller passes duplicate tiebreaks.
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
