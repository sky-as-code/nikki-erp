package promotion

import (
	"github.com/shopspring/decimal"
)

// buyXGetYEvaluator counts sets within this order only: no carry-over between orders, no points.
type buyXGetYEvaluator struct{}

// Qualify takes sets = max(sets_r) over qualifying rules, never their sum (BR-R08, AC-15). Ties go
// to the smaller rule id.
func (this buyXGetYEvaluator) Qualify(ctx EvalContext, program Program) Qualification {
	rules := rulesOfPhase(program, PhaseApply)
	if len(rules) == 0 {
		return Qualification{ReasonCode: ReasonNoRule}
	}
	evaluations := make([]ruleEvaluation, 0, len(rules))
	best := -1
	for _, rule := range rules {
		evaluation := measureRule(ctx.Input, program, rule, true)
		evaluations = append(evaluations, evaluation)
		if evaluation.qualified && (best < 0 || evaluation.sets > evaluations[best].sets) {
			best = len(evaluations) - 1
		}
	}
	if best < 0 {
		return Qualification{ReasonCode: strongestReason(evaluations)}
	}
	return qualificationOf(evaluations[best])
}

// Benefit grants sets × gift_quantity of the gift variant. Only free_product is valid for BXGY
// (BR §8.5); any other reward type is reported as not applicable rather than guessed.
func (this buyXGetYEvaluator) Benefit(
	_ EvalContext, program Program, reward Reward, qualification Qualification,
) Benefit {
	if reward.RewardType != RewardFreeProduct {
		return Benefit{RewardType: reward.RewardType, NotApplicableReason: ReasonNoReward}
	}
	quantity := reward.GiftQuantity.Mul(decimal.NewFromInt(qualification.Sets))
	return giftBenefit(program, reward, quantity)
}
