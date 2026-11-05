package promotion

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func dec(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}

func requireDec(t *testing.T, expected string, actual decimal.Decimal, label string) {
	t.Helper()
	require.Truef(t, dec(expected).Equal(actual), "%s: expected %s, got %s", label, expected, actual)
}

// buyLine is a purchase line: quantity units at an untaxed unit price, base UoM equal to the line UoM.
func buyLine(key string, number int32, quantity string, unitPrice string) Line {
	qty := dec(quantity)
	return Line{
		Key:               key,
		LineNumber:        number,
		LineType:          LineTypeProduct,
		ProductVariantId:  "var-" + key,
		ProductTemplateId: "tpl-" + key,
		UomId:             "uom-each",
		Quantity:          qty,
		BaseQuantity:      qty,
		UntaxedAmount:     qty.Mul(dec(unitPrice)),
		TaxAmount:         decimal.Zero,
	}
}

func withTax(line Line, rate string) Line {
	line.TaxAmount = line.UntaxedAmount.Mul(dec(rate))
	return line
}

func withCategories(line Line, categoryIds ...string) Line {
	line.CategoryIds = categoryIds
	return line
}

func withQualificationUom(line Line, uomId string, quantity string) Line {
	line.QualificationQuantities = map[string]decimal.Decimal{uomId: dec(quantity)}
	return line
}

func giftLine(key string, number int32, promotionId string, rewardId string, quantity string) Line {
	return Line{
		Key: key, LineNumber: number, LineType: LineTypePromotionGift,
		ProductVariantId: "var-gift", Quantity: dec(quantity), BaseQuantity: dec(quantity),
		UntaxedAmount: decimal.Zero, TaxAmount: decimal.Zero,
		SourcePromotionId: promotionId, SourceRewardId: rewardId,
	}
}

func applyRule(id string, minQuantity string, minAmount string) Rule {
	return Rule{
		Id: id, Phase: PhaseApply, MinQuantity: dec(minQuantity), MinAmount: dec(minAmount),
		AmountTaxBasis: TaxBasisExcluded,
	}
}

func percentReward(id string, scope string, value string, capAmount string) Reward {
	return Reward{
		Id: id, RewardType: RewardDiscount, DiscountMethod: MethodPercentage,
		DiscountValue: dec(value), DiscountScope: scope, MaxDiscountAmount: dec(capAmount),
	}
}

func fixedReward(id string, scope string, value string) Reward {
	return Reward{
		Id: id, RewardType: RewardDiscount, DiscountMethod: MethodFixedAmount,
		DiscountValue: dec(value), DiscountScope: scope, MaxDiscountAmount: decimal.Zero,
	}
}

func giftReward(id string, variantId string, quantity string) Reward {
	return Reward{
		Id: id, RewardType: RewardFreeProduct, GiftProductVariantId: variantId,
		GiftQuantity: dec(quantity), GiftUomId: "uom-each", MaxDiscountAmount: decimal.Zero,
	}
}

func automatic(id string, sequence int32, rules []Rule, rewards ...Reward) Program {
	return Program{
		Id: id, ProgramType: ProgramAutomatic, Sequence: sequence, ConfigVersion: 1,
		CurrencyId: "VND", Rules: rules, Rewards: rewards,
	}
}

func orderOf(lines []Line, programs ...Program) Input {
	return Input{
		OrderId: "order-1", StageVersion: 1, CurrencyId: "VND", CurrencyScale: 0,
		Lines: lines, Programs: programs,
	}
}

func programResult(t *testing.T, result Result, promotionId string) ProgramResult {
	t.Helper()
	for _, program := range result.Programs {
		if program.PromotionId == promotionId {
			return program
		}
	}
	require.Failf(t, "program missing", "no result for %s", promotionId)
	return ProgramResult{}
}

func allocationOn(program ProgramResult, lineKey string) decimal.Decimal {
	total := decimal.Zero
	for _, allocation := range program.Allocations {
		if allocation.LineKey == lineKey {
			total = total.Add(allocation.DiscountUntaxedAmount)
		}
	}
	return total
}
