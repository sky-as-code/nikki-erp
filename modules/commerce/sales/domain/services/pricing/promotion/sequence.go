package promotion

import (
	"maps"

	"github.com/shopspring/decimal"
)

func isOrderDiscount(current *plan) bool {
	return current.pending() && current.reward.RewardType == RewardDiscount &&
		current.reward.DiscountScope == ScopeOrder
}

// resolveOrderWinner keeps one order-scope discount: the largest D on the pre-promotion base, then
// the smaller sequence, then the smaller id (BR-R14). Plans are already in (sequence, id) order, so
// a strict comparison keeps the earliest on a tie. Losers are not applicable, so their vouchers are
// never consumed (AC-19).
func resolveOrderWinner(ctx EvalContext, plans []*plan) {
	winner := -1
	winnerAmount := decimal.Zero
	for i, current := range plans {
		if !isOrderDiscount(current) {
			continue
		}
		benefit := current.evaluator.Benefit(ctx, current.program, *current.reward, current.result.Qualification)
		if winner < 0 || benefit.DiscountUntaxedAmount.GreaterThan(winnerAmount) {
			winner, winnerAmount = i, benefit.DiscountUntaxedAmount
		}
	}
	for i, current := range plans {
		if i != winner && isOrderDiscount(current) {
			current.settle(OutcomeNotApplicable, ReasonOrderDiscountNotBest)
		}
	}
}

// applyPlans applies the pending plans in order, each over the base its predecessors left.
func applyPlans(input *Input, plans []*plan, remaining map[string]decimal.Decimal) {
	calculationSequence := int32(0)
	for _, current := range plans {
		if !current.pending() {
			continue
		}
		ctx := EvalContext{Input: input, Remaining: maps.Clone(remaining)}
		benefit := current.evaluator.Benefit(ctx, current.program, *current.reward, current.result.Qualification)
		if benefit.NotApplicableReason != "" {
			current.settle(OutcomeNotApplicable, benefit.NotApplicableReason)
			continue
		}
		// Qualifying is independent of the benefit; a zero benefit creates no application (AC-40).
		if !benefit.DiscountUntaxedAmount.IsPositive() && benefit.Gift == nil {
			current.settle(OutcomeNotApplicable, ReasonZeroBenefit)
			continue
		}
		calculationSequence++
		commitBenefit(current, benefit, calculationSequence, remaining)
	}
}

func commitBenefit(
	current *plan, benefit Benefit, calculationSequence int32, remaining map[string]decimal.Decimal,
) {
	current.settle(OutcomeApplied, "")
	current.result.CalculationSequence = calculationSequence
	current.result.DiscountUntaxedAmount = benefit.DiscountUntaxedAmount
	current.result.Gift = benefit.Gift
	for _, allocation := range benefit.Allocations {
		allocation.CalculationSequence = calculationSequence
		current.result.Allocations = append(current.result.Allocations, allocation)
		remaining[allocation.LineKey] = remaining[allocation.LineKey].Sub(allocation.DiscountUntaxedAmount)
	}
}

func assemble(input *Input, plans []*plan, remaining map[string]decimal.Decimal) Result {
	result := Result{Programs: make([]ProgramResult, len(plans))}
	for i, current := range plans {
		result.Programs[i] = current.result
	}
	for _, line := range purchaseLines(input) {
		left := remaining[line.Key]
		result.Lines = append(result.Lines, LineOutcome{
			Key:                    line.Key,
			BaseUntaxedAmount:      line.UntaxedAmount,
			DiscountUntaxedAmount:  line.UntaxedAmount.Sub(left),
			RemainingUntaxedAmount: left,
		})
	}
	result.Gifts, result.RemovedGiftLineKeys = reconcileGifts(input, result.Programs)
	return result
}
