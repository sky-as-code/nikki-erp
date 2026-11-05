package promotion

import (
	"github.com/shopspring/decimal"
)

var hundred = decimal.NewFromInt(100)

// discountEvaluator serves automatic, discount_code, coupon and the apply phase of
// next_order_coupon: qualify once by OR'd rules, then one reward of any type.
type discountEvaluator struct{}

func (this discountEvaluator) Qualify(ctx EvalContext, program Program) Qualification {
	return qualifyFirst(ctx.Input, program, PhaseApply)
}

func (this discountEvaluator) Benefit(
	ctx EvalContext, program Program, reward Reward, _ Qualification,
) Benefit {
	switch reward.RewardType {
	case RewardDiscount:
		return discountBenefit(ctx, program, reward)
	case RewardFreeProduct:
		// Outside BXGY the gift is granted once per program per order (BR-R08).
		return giftBenefit(program, reward, reward.GiftQuantity)
	default:
		return shippingBenefit(reward)
	}
}

// shippingBenefit: Sales has no shipping-fee line, so free_shipping never applies and is never
// converted into a merchandise discount (AC-39).
func shippingBenefit(reward Reward) Benefit {
	return Benefit{RewardType: reward.RewardType, NotApplicableReason: ReasonNoShippingLine}
}

func giftBenefit(program Program, reward Reward, quantity decimal.Decimal) Benefit {
	benefit := Benefit{RewardType: RewardFreeProduct, DiscountUntaxedAmount: decimal.Zero}
	if !quantity.IsPositive() {
		return benefit
	}
	benefit.Gift = &GiftIntent{
		PromotionId:      program.Id,
		RewardId:         reward.Id,
		ProductVariantId: reward.GiftProductVariantId,
		UomId:            reward.GiftUomId,
		Quantity:         quantity,
	}
	return benefit
}

// discountBenefit is BR-R03..R06 over the remaining untaxed base.
func discountBenefit(ctx EvalContext, program Program, reward Reward) Benefit {
	benefit := Benefit{RewardType: RewardDiscount, DiscountUntaxedAmount: decimal.Zero}
	targets := discountTargets(ctx, program, reward)
	if len(targets) == 0 {
		return benefit
	}
	scale := ctx.Input.CurrencyScale
	amount := discountAmount(reward, targets, scale)
	if !amount.IsPositive() {
		return benefit
	}

	shareTargets := make([]shareTarget, len(targets))
	for i, target := range targets {
		shareTargets[i] = shareTarget{
			key: target.line.Key, lineNumber: target.line.LineNumber,
			weight: target.basis, capacity: target.basis,
		}
	}
	shares := allocateLargestRemainder(amount, shareTargets, scale)
	for i, target := range targets {
		if !shares[i].IsPositive() {
			continue
		}
		benefit.Allocations = append(benefit.Allocations, Allocation{
			LineKey:               target.line.Key,
			AllocationType:        AllocationMerchandiseDiscount,
			EligibleQuantity:      target.quantity,
			BaseUntaxedAmount:     target.basis,
			DiscountUntaxedAmount: shares[i],
			TaxReductionAmount:    decimal.Zero,
		})
	}
	benefit.DiscountUntaxedAmount = amount
	return benefit
}

// discountAmount is D = min(D_raw, eligible_remaining, cap), rounded to the currency and never above
// what the targets can absorb on the currency grain (BR-R04).
func discountAmount(reward Reward, targets []discountTarget, scale int32) decimal.Decimal {
	base := decimal.Zero
	absorbable := decimal.Zero
	for _, target := range targets {
		base = base.Add(target.basis)
		absorbable = absorbable.Add(target.basis.Truncate(scale))
	}
	raw := reward.DiscountValue
	if reward.DiscountMethod == MethodPercentage {
		raw = base.Mul(reward.DiscountValue).Div(hundred)
	}
	amount := decimal.Min(raw, base)
	if reward.MaxDiscountAmount.IsPositive() {
		amount = decimal.Min(amount, reward.MaxDiscountAmount)
	}
	amount = amount.Round(scale)
	if amount.GreaterThan(absorbable) {
		amount = absorbable
	}
	if amount.IsNegative() {
		return decimal.Zero
	}
	return amount
}
