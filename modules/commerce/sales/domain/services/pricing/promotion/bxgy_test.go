package promotion

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func bxgyRule(id string, minQuantity string, variantIds ...string) Rule {
	rule := applyRule(id, minQuantity, "0")
	rule.QualificationUomId = "uom-bottle"
	rule.Selector = Selector{ProductVariantIds: variantIds}
	return rule
}

func bottles(key string, number int32, variantId string, quantity string) Line {
	line := withQualificationUom(buyLine(key, number, quantity, "10000"), "uom-bottle", quantity)
	line.ProductVariantId = variantId
	return line
}

func buyXGetY(id string, rules []Rule, reward Reward) Program {
	program := automatic(id, 10, rules, reward)
	program.ProgramType = ProgramBuyXGetY
	return program
}

func confirmGift(input Input, promotionId string, rewardId string) Input {
	input.Selections = append(input.Selections, Selection{PromotionId: promotionId, RewardId: rewardId})
	return input
}

func TestBuyXGetYSets(t *testing.T) {
	cases := []struct {
		name     string
		lines    []Line
		rules    []Rule
		outcome  string
		sets     int64
		quantity string
	}{
		{
			name:  "AC-14 seven bottles, buy 3 get 1: two gifts, the leftover is not carried",
			lines: []Line{bottles("l1", 1, "var-a", "7")}, rules: []Rule{bxgyRule("r1", "3", "var-a")},
			outcome: OutcomeApplied, sets: 2, quantity: "2",
		},
		{
			name:    "AC-15 two overlapping rules with 2 and 3 sets give 3, not 5",
			lines:   []Line{bottles("l1", 1, "var-a", "6")},
			rules:   []Rule{bxgyRule("r1", "3", "var-a"), bxgyRule("r2", "2", "var-a")},
			outcome: OutcomeApplied, sets: 3, quantity: "3",
		},
		{
			name:  "below X",
			lines: []Line{bottles("l1", 1, "var-a", "2")}, rules: []Rule{bxgyRule("r1", "3", "var-a")},
			outcome: OutcomeNotEligible,
		},
		{
			name: "BR-C07 a line not convertible to the qualification UoM does not count",
			lines: []Line{bottles("l1", 1, "var-a", "2"),
				func() Line { line := buyLine("kg", 2, "5", "1000"); line.ProductVariantId = "var-a"; return line }()},
			rules: []Rule{bxgyRule("r1", "3", "var-a")}, outcome: OutcomeNotEligible,
		},
		{
			name:  "BR-C08 a BXGY rule without positive X never qualifies",
			lines: []Line{bottles("l1", 1, "var-a", "5")}, rules: []Rule{bxgyRule("r1", "0", "var-a")},
			outcome: OutcomeNotEligible,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			input := orderOf(testCase.lines, buyXGetY("bx", testCase.rules, giftReward("gift", "var-a", "1")))
			result := Apply(confirmGift(input, "bx", "gift"))
			program := programResult(t, result, "bx")
			require.Equal(t, testCase.outcome, program.Outcome)
			if testCase.outcome != OutcomeApplied {
				require.Empty(t, result.Gifts)
				return
			}
			require.Equal(t, testCase.sets, program.Qualification.Sets)
			require.NotNil(t, program.Gift)
			requireDec(t, testCase.quantity, program.Gift.Quantity, "gift quantity")
			require.Empty(t, program.Allocations)
		})
	}
}

// AC-16: the gift is the same variant the customer buys. Repricing with the gift line present must
// neither raise the set count nor add a second gift line.
func TestBuyXGetYGiftDoesNotFeedItself(t *testing.T) {
	program := buyXGetY("bx", []Rule{bxgyRule("r1", "3", "var-a")}, giftReward("gift", "var-a", "1"))
	first := Apply(confirmGift(orderOf([]Line{bottles("l1", 1, "var-a", "5")}, program), "bx", "gift"))
	require.Len(t, first.Gifts, 1)
	require.Empty(t, first.Gifts[0].ExistingLineKey)

	existing := giftLine("g1", 2, "bx", "gift", "1")
	existing.ProductVariantId = "var-a"
	existing.QualificationQuantities = map[string]decimal.Decimal{"uom-bottle": dec("1")}
	second := Apply(confirmGift(orderOf([]Line{bottles("l1", 1, "var-a", "5"), existing}, program), "bx", "gift"))
	require.Equal(t, int64(1), programResult(t, second, "bx").Qualification.Sets)
	require.Len(t, second.Gifts, 1)
	require.Equal(t, "g1", second.Gifts[0].ExistingLineKey)
	require.Empty(t, second.RemovedGiftLineKeys)
}

// BR-R09: once the condition is lost, the draft gift line is removed; a duplicate is removed too.
func TestGiftLineRemovedWhenNoLongerQualified(t *testing.T) {
	program := buyXGetY("bx", []Rule{bxgyRule("r1", "3", "var-a")}, giftReward("gift", "var-a", "1"))
	lines := []Line{bottles("l1", 1, "var-a", "2"), giftLine("g1", 2, "bx", "gift", "1")}
	result := Apply(confirmGift(orderOf(lines, program), "bx", "gift"))
	require.Empty(t, result.Gifts)
	require.Equal(t, []string{"g1"}, result.RemovedGiftLineKeys)

	lines = []Line{
		bottles("l1", 1, "var-a", "3"), giftLine("g2", 3, "bx", "gift", "1"), giftLine("g1", 2, "bx", "gift", "1"),
	}
	result = Apply(confirmGift(orderOf(lines, program), "bx", "gift"))
	require.Equal(t, "g1", result.Gifts[0].ExistingLineKey)
	require.Equal(t, []string{"g2"}, result.RemovedGiftLineKeys)
}

// BR-R12: a gift always waits for confirmation; the option previews the server-computed quantity.
func TestGiftNeedsConfirmation(t *testing.T) {
	program := buyXGetY("bx", []Rule{bxgyRule("r1", "3", "var-a")}, giftReward("gift", "var-a", "2"))
	result := Apply(orderOf([]Line{bottles("l1", 1, "var-a", "9")}, program))
	outcome := programResult(t, result, "bx")
	require.Equal(t, OutcomeSelectionRequired, outcome.Outcome)
	require.Len(t, outcome.Rewards, 1)
	requireDec(t, "6", outcome.Rewards[0].GiftQuantity, "preview")
	require.Empty(t, result.Gifts)
}

// Outside BXGY a free product is granted once per order, whatever the quantity bought.
func TestAutomaticGiftOncePerOrder(t *testing.T) {
	program := automatic("p1", 10, []Rule{applyRule("r1", "2", "0")}, giftReward("gift", "var-g", "1"))
	result := Apply(confirmGift(orderOf([]Line{buyLine("l1", 1, "10", "1000")}, program), "p1", "gift"))
	requireDec(t, "1", programResult(t, result, "p1").Gift.Quantity, "gift quantity")
}
