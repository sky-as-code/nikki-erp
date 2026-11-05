package promotion

import (
	"slices"
	"sort"

	"github.com/shopspring/decimal"
)

// plan carries one program through Apply. reward is nil until a reward is chosen; an empty Outcome
// means "still to apply".
type plan struct {
	program   Program
	evaluator ProgramEvaluator
	reward    *Reward
	result    ProgramResult
}

// Apply runs the default registry. See Registry.Apply.
func Apply(input Input) Result {
	return DefaultRegistry().Apply(input)
}

// Apply follows BR-R15 in one pass, never iterating to a fixed point:
//  1. qualify every candidate once on the pre-promotion base and choose its reward (BR-R12);
//  2. keep only the best order-scope discount, measured on that same base (BR-R14);
//  3. apply the rest in (sequence, id) order, each over what earlier ones left (BR-R03);
//  4. drop zero benefits (BR-R15) and reconcile gift lines (BR-R09).
func (this *Registry) Apply(input Input) Result {
	programs := sortedPrograms(input.Programs)
	baseContext := EvalContext{Input: &input, Remaining: baseRemaining(&input)}
	selections := selectionsByPromotion(input.Selections)

	plans := make([]*plan, len(programs))
	for i, program := range programs {
		plans[i] = this.prepare(baseContext, program, selections)
	}
	resolveOrderWinner(baseContext, plans)

	remaining := baseRemaining(&input)
	applyPlans(&input, plans, remaining)
	return assemble(&input, plans, remaining)
}

func sortedPrograms(programs []Program) []Program {
	sorted := slices.Clone(programs)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Sequence != sorted[j].Sequence {
			return sorted[i].Sequence < sorted[j].Sequence
		}
		return sorted[i].Id < sorted[j].Id
	})
	// One benefit per promotion per order (BR-R13): a repeated id, e.g. a second voucher of the same
	// promotion, is ignored rather than doubling the reward. The caller owns the replace flow.
	seen := map[string]bool{}
	unique := sorted[:0]
	for _, program := range sorted {
		if !seen[program.Id] {
			seen[program.Id] = true
			unique = append(unique, program)
		}
	}
	return unique
}

func baseRemaining(input *Input) map[string]decimal.Decimal {
	remaining := map[string]decimal.Decimal{}
	for _, line := range purchaseLines(input) {
		remaining[line.Key] = line.UntaxedAmount
	}
	return remaining
}

func selectionsByPromotion(selections []Selection) map[string]Selection {
	byPromotion := make(map[string]Selection, len(selections))
	for _, selection := range selections {
		byPromotion[selection.PromotionId] = selection
	}
	return byPromotion
}

func (this *Registry) prepare(ctx EvalContext, program Program, selections map[string]Selection) *plan {
	current := &plan{program: program, result: newProgramResult(program)}
	evaluator, found := this.Evaluator(program.ProgramType)
	if !found {
		current.settle(OutcomeNotApplicable, ReasonUnknownProgramType)
		return current
	}
	current.evaluator = evaluator
	if program.CurrencyId != "" && program.CurrencyId != ctx.Input.CurrencyId {
		current.settle(OutcomeNotEligible, ReasonCurrencyMismatch)
		return current
	}
	qualification := evaluator.Qualify(ctx, program)
	current.result.Qualification = qualification
	if !qualification.Qualified {
		current.settle(OutcomeNotEligible, qualification.ReasonCode)
		return current
	}
	if len(program.Rewards) == 0 {
		current.settle(OutcomeNotApplicable, ReasonNoReward)
		return current
	}
	current.result.Rewards = rewardOptions(ctx, current)
	chooseReward(current, selections)
	return current
}

func newProgramResult(program Program) ProgramResult {
	return ProgramResult{
		PromotionId:           program.Id,
		ProgramType:           program.ProgramType,
		Sequence:              program.Sequence,
		ConfigVersion:         program.ConfigVersion,
		VoucherId:             program.VoucherId,
		DiscountUntaxedAmount: decimal.Zero,
		TaxReductionAmount:    decimal.Zero,
	}
}

func (this *plan) settle(outcome string, reason string) {
	this.result.Outcome = outcome
	this.result.ReasonCode = reason
}

func (this *plan) pending() bool {
	return this.reward != nil && this.result.Outcome == ""
}

func sortedRewards(program Program) []Reward {
	rewards := make([]Reward, len(program.Rewards))
	copy(rewards, program.Rewards)
	sort.SliceStable(rewards, func(i, j int) bool { return rewards[i].Id < rewards[j].Id })
	return rewards
}

// rewardOptions previews each alternative; a gift's quantity does not depend on the remaining base,
// so the preview equals what applying it would grant.
func rewardOptions(ctx EvalContext, current *plan) []RewardOption {
	rewards := sortedRewards(current.program)
	options := make([]RewardOption, len(rewards))
	for i, reward := range rewards {
		options[i] = RewardOption{RewardId: reward.Id, RewardType: reward.RewardType, GiftQuantity: decimal.Zero}
		if reward.RewardType != RewardFreeProduct {
			continue
		}
		benefit := current.evaluator.Benefit(ctx, current.program, reward, current.result.Qualification)
		if benefit.Gift != nil {
			options[i].GiftQuantity = benefit.Gift.Quantity
		}
	}
	return options
}

// chooseReward is BR-R12: an explicit selection wins; a lone discount or shipping reward applies by
// itself; a gift or several alternatives wait for a choice, which never blocks Confirm.
func chooseReward(current *plan, selections map[string]Selection) {
	rewards := sortedRewards(current.program)
	selection, hasSelection := selections[current.program.Id]
	if hasSelection && selection.Declined {
		current.settle(OutcomeDeclined, "")
		return
	}
	if hasSelection && selection.RewardId != "" {
		for i := range rewards {
			if rewards[i].Id == selection.RewardId {
				current.pick(rewards[i])
				return
			}
		}
	}
	if len(rewards) == 1 && rewards[0].RewardType != RewardFreeProduct {
		current.pick(rewards[0])
		return
	}
	current.settle(OutcomeSelectionRequired, "")
}

func (this *plan) pick(reward Reward) {
	chosen := reward
	this.reward = &chosen
	this.result.RewardId = reward.Id
	this.result.RewardType = reward.RewardType
	for i := range this.result.Rewards {
		this.result.Rewards[i].Selected = this.result.Rewards[i].RewardId == reward.Id
	}
}
