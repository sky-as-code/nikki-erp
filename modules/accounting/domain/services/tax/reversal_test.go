package tax

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// AC-TAX-29 and AC-TAX-SUP-16: a full refund reverses the original charge exactly.
//
// The sale was at 8%; the current rate is irrelevant, and is never consulted.
func TestFullReversalNegatesOriginal(t *testing.T) {
	results := ReverseFull([]ReversalComponentInput{{
		OriginalComponentReference: "L1/VAT8",
		OriginalTaxAmount:          dec("8000"),
		OriginalReversibleBasis:    dec("100000"),
	}})

	assert.Len(t, results, 1)
	assert.True(t, results[0].ReversalTaxAmount.Equal(dec("-8000")),
		"reversal = %s", results[0].ReversalTaxAmount)
	assert.True(t, results[0].RemainingTaxAmount.IsZero())
}

// A full refund after a partial one reverses only what is left.
func TestFullReversalAfterPartial(t *testing.T) {
	results := ReverseFull([]ReversalComponentInput{{
		OriginalComponentReference: "L1/VAT10",
		OriginalTaxAmount:          dec("10000"),
		AlreadyReversedTaxAmount:   dec("4000"),
	}})

	assert.True(t, results[0].ReversalTaxAmount.Equal(dec("-6000")),
		"reversal = %s", results[0].ReversalTaxAmount)
}

// A partial refund reverses in proportion to the basis returned.
func TestPartialReversalIsProportional(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "0.01")

	results := ReversePartial([]ReversalComponentInput{{
		OriginalComponentReference: "L1/VAT10",
		OriginalReversibleBasis:    dec("10"),
		OriginalTaxAmount:          dec("100"),
		RequestedReversalBasis:     dec("3"),
	}}, rounder)

	assert.True(t, results[0].ReversalTaxAmount.Equal(dec("-30")),
		"reversal = %s", results[0].ReversalTaxAmount)
	assert.True(t, results[0].RemainingTaxAmount.Equal(dec("70")))
}

// AC-TAX-30 and AC-TAX-SUP-18: the final reversal absorbs the rounding residual, so the refunds
// sum to exactly the original charge.
//
// Three refunds of a third of 10.00 round to 3.33 each and leave a cent stranded. The last one
// takes what remains instead of its proportional share.
func TestFinalPartialReversalAbsorbsResidual(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "0.01")
	original := dec("10.00")
	basis := dec("3")

	reversedSoFar := decimal.Zero
	basisSoFar := decimal.Zero

	for step := 1; step <= 3; step++ {
		isFinal := step == 3
		results := ReversePartial([]ReversalComponentInput{{
			OriginalComponentReference: "L1/VAT10",
			OriginalReversibleBasis:    basis,
			OriginalTaxAmount:          original,
			AlreadyReversedBasis:       basisSoFar,
			AlreadyReversedTaxAmount:   reversedSoFar,
			RequestedReversalBasis:     dec("1"),
			IsFinalReversal:            isFinal,
		}}, rounder)

		reversedSoFar = reversedSoFar.Add(results[0].ReversalTaxAmount.Neg())
		basisSoFar = basisSoFar.Add(dec("1"))
	}

	assert.True(t, reversedSoFar.Equal(original),
		"total reversed = %s, want exactly %s", reversedSoFar, original)
}

// AC-TAX-SUP-17 and TAX-INV-11: no sequence of partial refunds may reverse more than was charged.
//
// Here the caller asks to reverse the whole basis twice over — a duplicated refund request, which
// is the realistic way this happens.
func TestPartialReversalNeverExceedsOriginal(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "0.01")

	results := ReversePartial([]ReversalComponentInput{{
		OriginalComponentReference: "L1/VAT10",
		OriginalReversibleBasis:    dec("10"),
		OriginalTaxAmount:          dec("100"),
		AlreadyReversedTaxAmount:   dec("80"),
		RequestedReversalBasis:     dec("10"), // would prorate to 100, but only 20 is left
	}}, rounder)

	assert.True(t, results[0].ReversalTaxAmount.Equal(dec("-20")),
		"reversal = %s, want the remaining 20 only", results[0].ReversalTaxAmount)
	assert.True(t, results[0].RemainingTaxAmount.IsZero())
}

// A malformed input with no basis to prorate against reverses nothing, rather than assuming the
// whole remainder and over-refunding on the strength of a bad request.
func TestPartialReversalWithoutBasisReversesNothing(t *testing.T) {
	rounder := policy(models.RoundingHalfUp, "0.01")

	results := ReversePartial([]ReversalComponentInput{{
		OriginalComponentReference: "L1/VAT10",
		OriginalReversibleBasis:    decimal.Zero,
		OriginalTaxAmount:          dec("100"),
		RequestedReversalBasis:     dec("5"),
	}}, rounder)

	assert.True(t, results[0].ReversalTaxAmount.IsZero())
	assert.True(t, results[0].RemainingTaxAmount.Equal(dec("100")))
}

// A reversal is signed negative, so a caller summing charges and refunds gets the net without
// having to know which is which.
func TestReversalsAreNegative(t *testing.T) {
	full := ReverseFull([]ReversalComponentInput{{
		OriginalComponentReference: "L1",
		OriginalTaxAmount:          dec("500"),
	}})
	assert.True(t, full[0].ReversalTaxAmount.IsNegative())

	partial := ReversePartial([]ReversalComponentInput{{
		OriginalComponentReference: "L1",
		OriginalReversibleBasis:    dec("10"),
		OriginalTaxAmount:          dec("500"),
		RequestedReversalBasis:     dec("5"),
	}}, policy(models.RoundingHalfUp, "0.01"))
	assert.True(t, partial[0].ReversalTaxAmount.IsNegative())
}
