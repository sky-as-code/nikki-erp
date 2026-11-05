package promotion

import (
	"github.com/shopspring/decimal"
)

type allocationRef struct {
	program    int
	allocation int
}

// AttributeTax fills the tax reductions from Accounting's answer. This package never computes tax
// (BR-R03): the caller resolves tax on the pre-promotion base and again on the discounted lines, and
// passes both per line key. Each line's tax delta is split over that line's allocations in
// proportion to their untaxed discount, largest remainder, so the per-line mapping is kept and
// tax groups are never pooled (BR-R06). Program totals are the sums of their allocations.
func AttributeTax(
	result *Result, taxBefore map[string]decimal.Decimal, taxAfter map[string]decimal.Decimal, scale int32,
) {
	byLine := map[string][]allocationRef{}
	for programIndex := range result.Programs {
		for allocationIndex, allocation := range result.Programs[programIndex].Allocations {
			ref := allocationRef{program: programIndex, allocation: allocationIndex}
			byLine[allocation.LineKey] = append(byLine[allocation.LineKey], ref)
		}
	}
	for lineKey, refs := range byLine {
		attributeLineTax(result, refs, taxBefore[lineKey].Sub(taxAfter[lineKey]), scale)
	}
	for programIndex := range result.Programs {
		total := decimal.Zero
		for _, allocation := range result.Programs[programIndex].Allocations {
			total = total.Add(allocation.TaxReductionAmount)
		}
		result.Programs[programIndex].TaxReductionAmount = total
	}
}

// attributeLineTax splits one line's delta. A negative delta (tax rose through rounding) is split
// by magnitude and negated, so the sign survives.
func attributeLineTax(result *Result, refs []allocationRef, delta decimal.Decimal, scale int32) {
	targets := make([]shareTarget, len(refs))
	for i, ref := range refs {
		allocation := result.Programs[ref.program].Allocations[ref.allocation]
		targets[i] = shareTarget{
			key:        result.Programs[ref.program].PromotionId,
			lineNumber: allocation.CalculationSequence,
			weight:     allocation.DiscountUntaxedAmount,
		}
	}
	shares := allocateLargestRemainder(delta.Abs().Round(scale), targets, scale)
	for i, ref := range refs {
		share := shares[i]
		if delta.IsNegative() {
			share = share.Neg()
		}
		result.Programs[ref.program].Allocations[ref.allocation].TaxReductionAmount = share
	}
}
