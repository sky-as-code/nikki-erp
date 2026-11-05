package promotion

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func anyRule() []Rule {
	return []Rule{applyRule("r1", "0", "0")}
}

func TestDiscountAmounts(t *testing.T) {
	cases := []struct {
		name     string
		lines    []Line
		reward   Reward
		expected string
		byLine   map[string]string
	}{
		{
			name:  "AC-11 percentage capped on untaxed value",
			lines: []Line{buyLine("l1", 1, "1", "300000")}, reward: percentReward("rw", ScopeOrder, "20", "40000"),
			expected: "40000", byLine: map[string]string{"l1": "40000"},
		},
		{
			name:  "AC-12 fixed amount never exceeds what remains",
			lines: []Line{buyLine("l1", 1, "1", "30000")}, reward: fixedReward("rw", ScopeOrder, "50000"),
			expected: "30000", byLine: map[string]string{"l1": "30000"},
		},
		{
			name:   "AC-38 fixed amount split proportionally",
			lines:  []Line{buyLine("l1", 1, "1", "100000"), buyLine("l2", 2, "1", "200000")},
			reward: fixedReward("rw", ScopeOrder, "30000"), expected: "30000",
			byLine: map[string]string{"l1": "10000", "l2": "20000"},
		},
		{
			name:   "largest remainder: 100 over three equal lines, tie by line order",
			lines:  []Line{buyLine("l1", 1, "1", "1000"), buyLine("l2", 2, "1", "1000"), buyLine("l3", 3, "1", "1000")},
			reward: fixedReward("rw", ScopeOrder, "100"), expected: "100",
			byLine: map[string]string{"l1": "34", "l2": "33", "l3": "33"},
		},
		{
			name:   "largest remainder favours the biggest fraction, not the biggest line",
			lines:  []Line{buyLine("l1", 1, "1", "600"), buyLine("l2", 2, "1", "300"), buyLine("l3", 3, "1", "100")},
			reward: fixedReward("rw", ScopeOrder, "7"), expected: "7",
			byLine: map[string]string{"l1": "4", "l2": "2", "l3": "1"},
		},
		{
			name:   "AC-13 cheapest gives one unit, not the whole line",
			lines:  []Line{buyLine("a", 1, "3", "10000"), buyLine("b", 2, "1", "20000")},
			reward: percentReward("rw", ScopeCheapestProduct, "100", "0"), expected: "10000",
			byLine: map[string]string{"a": "10000"},
		},
		{
			name:   "BR-R05 cheapest line below one unit offers only what it has",
			lines:  []Line{buyLine("a", 1, "0.5", "10000"), buyLine("b", 2, "1", "20000")},
			reward: percentReward("rw", ScopeCheapestProduct, "100", "0"), expected: "5000",
			byLine: map[string]string{"a": "5000"},
		},
		{
			name:   "BR-R05 cheapest tie broken by line order",
			lines:  []Line{buyLine("b", 2, "1", "10000"), buyLine("a", 1, "1", "10000")},
			reward: percentReward("rw", ScopeCheapestProduct, "50", "0"), expected: "5000",
			byLine: map[string]string{"a": "5000"},
		},
		{
			name:     "specific products receive the discount, not the qualifying lines",
			lines:    []Line{buyLine("a", 1, "1", "10000"), buyLine("b", 2, "1", "30000")},
			reward:   withRewardSelector(percentReward("rw", ScopeSpecificProducts, "10", "0"), "tpl-b"),
			expected: "3000", byLine: map[string]string{"b": "3000"},
		},
		{
			name:  "percentage rounds to the currency",
			lines: []Line{buyLine("l1", 1, "1", "335")}, reward: percentReward("rw", ScopeOrder, "10", "0"),
			expected: "34", byLine: map[string]string{"l1": "34"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := Apply(orderOf(testCase.lines, automatic("p1", 10, anyRule(), testCase.reward)))
			program := programResult(t, result, "p1")
			require.Equal(t, OutcomeApplied, program.Outcome)
			requireDec(t, testCase.expected, program.DiscountUntaxedAmount, "discount")
			sum := decimal.Zero
			for _, allocation := range program.Allocations {
				sum = sum.Add(allocation.DiscountUntaxedAmount)
				require.False(t, allocation.DiscountUntaxedAmount.GreaterThan(allocation.BaseUntaxedAmount))
			}
			requireDec(t, testCase.expected, sum, "allocation sum")
			for lineKey, amount := range testCase.byLine {
				requireDec(t, amount, allocationOn(program, lineKey), lineKey)
			}
			require.Len(t, program.Allocations, len(testCase.byLine))
		})
	}
}

func withRewardSelector(reward Reward, templateIds ...string) Reward {
	reward.Selector = Selector{ProductTemplateIds: templateIds}
	return reward
}

// AC-10: 10% off an untaxed 200,000 at 10% VAT. The engine discounts 20,000 untaxed; Accounting's
// re-resolved tax gives the 2,000 reduction, 22,000 in total.
func TestTaxReductionFromAccounting(t *testing.T) {
	lines := []Line{withTax(buyLine("l1", 1, "1", "200000"), "0.1")}
	result := Apply(orderOf(lines, automatic("p1", 10, anyRule(), percentReward("rw", ScopeOrder, "10", "0"))))
	before := map[string]decimal.Decimal{"l1": dec("20000")}
	after := map[string]decimal.Decimal{"l1": dec("18000")}
	AttributeTax(&result, before, after, 0)

	program := programResult(t, result, "p1")
	requireDec(t, "20000", program.DiscountUntaxedAmount, "untaxed")
	requireDec(t, "2000", program.TaxReductionAmount, "tax reduction")
	requireDec(t, "22000", program.DiscountTotalAmount(), "total")
}

// AC-38 tax per line: two programs on one line share its tax delta by their untaxed discount.
func TestTaxReductionSplitPerLineAndProgram(t *testing.T) {
	lines := []Line{buyLine("l1", 1, "1", "100000"), buyLine("l2", 2, "1", "200000")}
	fixed := automatic("p1", 10, anyRule(), fixedReward("rw1", ScopeOrder, "30000"))
	onLine2 := withRewardSelector(fixedReward("rw2", ScopeSpecificProducts, "10000"), "tpl-l2")
	line2Only := automatic("p2", 20, anyRule(), onLine2)
	result := Apply(orderOf(lines, fixed, line2Only))
	AttributeTax(&result, map[string]decimal.Decimal{"l1": dec("10000"), "l2": dec("20000")},
		map[string]decimal.Decimal{"l1": dec("9000"), "l2": dec("17000")}, 0)

	first, second := programResult(t, result, "p1"), programResult(t, result, "p2")
	requireDec(t, "3000", first.TaxReductionAmount, "p1 tax")
	requireDec(t, "1000", second.TaxReductionAmount, "p2 tax")
	for _, allocation := range first.Allocations {
		if allocation.LineKey == "l2" {
			requireDec(t, "2000", allocation.TaxReductionAmount, "p1 on l2")
		}
	}
}

// BR-R03: a later program discounts what an earlier one left, in (sequence, id) order.
func TestStackingOverRemainingBase(t *testing.T) {
	lines := []Line{buyLine("a", 1, "1", "100000"), buyLine("b", 2, "1", "100000")}
	onA := automatic("p1", 10, anyRule(), withRewardSelector(fixedReward("rw1", ScopeSpecificProducts, "60000"), "tpl-a"))
	order := automatic("p2", 20, anyRule(), percentReward("rw2", ScopeOrder, "50", "0"))
	result := Apply(orderOf(lines, order, onA))

	first, second := programResult(t, result, "p1"), programResult(t, result, "p2")
	require.Equal(t, int32(1), first.CalculationSequence)
	require.Equal(t, int32(2), second.CalculationSequence)
	requireDec(t, "70000", second.DiscountUntaxedAmount, "50% of 140,000 remaining")
	requireDec(t, "20000", allocationOn(second, "a"), "a share")
	requireDec(t, "50000", allocationOn(second, "b"), "b share")
	requireDec(t, "20000", result.Lines[0].RemainingUntaxedAmount, "a left")
	requireDec(t, "50000", result.Lines[1].RemainingUntaxedAmount, "b left")
}

// Gift lines are never discounted (BR-R02) even by an order-scope reward.
func TestGiftLinesAreNotDiscounted(t *testing.T) {
	lines := []Line{buyLine("a", 1, "1", "1000"), giftLine("g", 2, "x", "y", "1")}
	result := Apply(orderOf(lines, automatic("p1", 10, anyRule(), percentReward("rw", ScopeOrder, "10", "0"))))
	program := programResult(t, result, "p1")
	require.Len(t, program.Allocations, 1)
	require.Equal(t, "a", program.Allocations[0].LineKey)
}
