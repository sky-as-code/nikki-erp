package pricing

import (
	"github.com/shopspring/decimal"
)

// AdjustmentSourcePromotion is the source_type of a promotion adjustment; its source_id names the
// sales_promotion_reward that granted it (plan D3).
const AdjustmentSourcePromotion = "promotion"

// PromotionLineEffect is one promotion's share of one line. DiscountUntaxed is what the BR engine
// allocated (pricing/promotion); TaxReduction is the line's tax delta attributed to it. Sales prices
// tax-inclusive, so the line's net falls by both.
type PromotionLineEffect struct {
	LineKey         string
	RewardId        string
	AdjustmentType  string
	Description     string
	BaseUntaxed     decimal.Decimal
	DiscountUntaxed decimal.Decimal
	TaxReduction    decimal.Decimal
}

// PromotionEffects is what the order flow hands back after running the promotion engine and the
// second tax resolution. TaxByLineKey, when set, replaces every line's tax: it is Accounting's answer
// on the discounted lines, and must cover every line.
type PromotionEffects struct {
	Lines        []PromotionLineEffect
	TaxByLineKey map[string]LineTax
}

// ApplyPromotionEffects is step 4, run on a finished Calculate result: it records each effect as a
// line adjustment, lowers the line, re-applies tax when given and re-rounds. The rounding adjustment
// of the first pass is dropped, since it rounded a total that no longer exists.
func ApplyPromotionEffects(result Result, effects PromotionEffects, currencyScale int32) Result {
	if len(effects.Lines) == 0 && effects.TaxByLineKey == nil {
		return result
	}
	adjustments, sequence := withoutRounding(result.Adjustments)
	lines := append([]LineResult(nil), result.Lines...)
	for _, effect := range effects.Lines {
		index := indexOfKey(lines, effect.LineKey)
		amount := effect.DiscountUntaxed.Add(effect.TaxReduction)
		if index < 0 || !amount.IsPositive() {
			continue
		}
		lines[index].DiscountAmount = lines[index].DiscountAmount.Add(amount)
		lines[index].NetAmount = lines[index].NetAmount.Sub(amount)
		sequence++
		adjustments = append(adjustments, Adjustment{
			Sequence: sequence, LineKey: effect.LineKey, Type: effect.AdjustmentType,
			SourceType: AdjustmentSourcePromotion, SourceId: effect.RewardId,
			Description: effect.Description, BaseAmount: effect.BaseUntaxed, Amount: amount.Neg(),
		})
	}
	if effects.TaxByLineKey != nil {
		lines = applyTax(lines, effects.TaxByLineKey, InternalScale)
	} else {
		for index := range lines {
			lines[index].FinalAmount = lines[index].NetAmount
		}
	}
	return finish(lines, adjustments, sequence, currencyScale)
}

func withoutRounding(adjustments []Adjustment) ([]Adjustment, int32) {
	kept := make([]Adjustment, 0, len(adjustments)+4)
	sequence := int32(0)
	for _, adjustment := range adjustments {
		if adjustment.Type == AdjustmentRounding {
			continue
		}
		kept = append(kept, adjustment)
		if adjustment.Sequence > sequence {
			sequence = adjustment.Sequence
		}
	}
	return kept, sequence
}
