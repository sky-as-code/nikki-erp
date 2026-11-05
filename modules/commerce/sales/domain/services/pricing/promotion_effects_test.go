package pricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

// Promotions land on a finished calculation: each effect lowers its line by the untaxed discount
// plus the tax reduction, is recorded as a line adjustment, and the totals are re-rounded once.

func TestPromotionEffectLowersTheLineAndIsExplained(t *testing.T) {
	base := Calculate(Input{Lines: []LineInput{lineOf("a", 1, "1", "110000"), lineOf("b", 2, "1", "55000")},
		Context: vndContext()})
	tax := map[string]LineTax{
		"a": {RateSnapshot: dec("0.1"), Amount: dec("9000")},
		"b": {RateSnapshot: dec("0.1"), Amount: dec("5000")},
	}
	result := ApplyPromotionEffects(base, PromotionEffects{
		Lines: []PromotionLineEffect{{
			LineKey: "a", RewardId: "rw-1", AdjustmentType: AdjustmentPercentageDiscount,
			BaseUntaxed: dec("100000"), DiscountUntaxed: dec("10000"), TaxReduction: dec("1000"),
		}},
		TaxByLineKey: tax,
	}, 0)

	line := lineByKey(t, result, "a")
	if !line.NetAmount.Equal(dec("99000")) || !line.DiscountAmount.Equal(dec("11000")) {
		t.Fatalf("line a net %s discount %s, want 99000 and 11000", line.NetAmount, line.DiscountAmount)
	}
	if !line.TaxAmount.Equal(dec("9000")) {
		t.Errorf("tax = %s, want Accounting's 9000 on the discounted line", line.TaxAmount)
	}
	if !result.GrandTotal.Equal(dec("154000")) {
		t.Errorf("grand total = %s, want 154000", result.GrandTotal)
	}
	found := false
	for _, adjustment := range result.Adjustments {
		if adjustment.SourceType == AdjustmentSourcePromotion {
			found = adjustment.SourceId == "rw-1" && adjustment.LineKey == "a" && adjustment.Amount.Equal(dec("-11000"))
		}
	}
	if !found {
		t.Errorf("no promotion adjustment explains the discount: %+v", result.Adjustments)
	}
}

// The first pass's rounding adjustment rounded a total that no longer exists; only one remains.
func TestPromotionEffectsReRoundOnce(t *testing.T) {
	base := Calculate(Input{Lines: []LineInput{lineOf("a", 1, "3", "333.4")}, Context: vndContext()})
	result := ApplyPromotionEffects(base, PromotionEffects{Lines: []PromotionLineEffect{{
		LineKey: "a", RewardId: "rw", AdjustmentType: AdjustmentFixedDiscount,
		DiscountUntaxed: dec("0.5"), TaxReduction: decimal.Zero,
	}}}, 0)
	rounding := 0
	for _, adjustment := range result.Adjustments {
		if adjustment.Type == AdjustmentRounding {
			rounding++
		}
	}
	if rounding > 1 {
		t.Fatalf("%d rounding adjustments, want at most one", rounding)
	}
	if !result.GrandTotal.Equal(result.GrandTotal.Round(0)) {
		t.Errorf("grand total %s is not rounded to the currency", result.GrandTotal)
	}
}

func TestNoEffectLeavesTheResultUntouched(t *testing.T) {
	base := Calculate(Input{Lines: []LineInput{lineOf("a", 1, "1", "1000")}, Context: vndContext()})
	result := ApplyPromotionEffects(base, PromotionEffects{}, 0)
	if !result.GrandTotal.Equal(base.GrandTotal) || len(result.Adjustments) != len(base.Adjustments) {
		t.Errorf("an empty effect set changed the result")
	}
}
