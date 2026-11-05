package promotion

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQualificationThresholds(t *testing.T) {
	tenPercent := percentReward("rw-1", ScopeOrder, "10", "0")
	cases := []struct {
		name    string
		lines   []Line
		rules   []Rule
		outcome string
		reason  string
		ruleId  string
	}{
		{
			name:  "AC-06 quantity met, amount missed",
			lines: []Line{buyLine("l1", 1, "3", "50000")},
			rules: []Rule{applyRule("r1", "3", "200000")}, outcome: OutcomeNotEligible, reason: ReasonConditionNotMet,
		},
		{
			name:  "AC-06 amount met, quantity missed",
			lines: []Line{buyLine("l1", 1, "2", "150000")},
			rules: []Rule{applyRule("r1", "3", "200000")}, outcome: OutcomeNotEligible, reason: ReasonConditionNotMet,
		},
		{
			name:  "AC-06 both met",
			lines: []Line{buyLine("l1", 1, "3", "70000")},
			rules: []Rule{applyRule("r1", "3", "200000")}, outcome: OutcomeApplied, ruleId: "r1",
		},
		{
			name:    "AC-07 only the second OR rule holds, thresholds are not summed",
			lines:   []Line{buyLine("l1", 1, "1", "120000")},
			rules:   []Rule{applyRule("r1", "5", "0"), applyRule("r2", "0", "100000")},
			outcome: OutcomeApplied, ruleId: "r2",
		},
		{
			name:    "AC-07 neither rule alone holds although their sum would",
			lines:   []Line{buyLine("l1", 1, "3", "40000")},
			rules:   []Rule{applyRule("r1", "4", "0"), applyRule("r2", "0", "150000")},
			outcome: OutcomeNotEligible, reason: ReasonConditionNotMet,
		},
		{
			name:    "BR-C08 zero thresholds still need one purchase line",
			lines:   []Line{giftLine("g1", 1, "other", "rw", "1")},
			rules:   []Rule{applyRule("r1", "0", "0")},
			outcome: OutcomeNotEligible, reason: ReasonNoEligibleLine,
		},
		{
			name:    "BR-C08 zero thresholds with one purchase line",
			lines:   []Line{buyLine("l1", 1, "1", "1000")},
			rules:   []Rule{applyRule("r1", "0", "0")},
			outcome: OutcomeApplied, ruleId: "r1",
		},
		{
			name:    "no apply rule",
			lines:   []Line{buyLine("l1", 1, "1", "1000")},
			rules:   []Rule{{Id: "r1", Phase: PhaseIssue}},
			outcome: OutcomeNotEligible, reason: ReasonNoRule,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := Apply(orderOf(testCase.lines, automatic("p1", 10, testCase.rules, tenPercent)))
			program := programResult(t, result, "p1")
			require.Equal(t, testCase.outcome, program.Outcome)
			require.Equal(t, testCase.reason, program.ReasonCode)
			require.Equal(t, testCase.ruleId, program.Qualification.RuleId)
		})
	}
}

// AC-08 without the tag dimension (Inventory has no product tag): (A OR B) AND category C,
// descendants included through the ancestor ids the caller supplies.
func TestSelectorOrInsideAndAcross(t *testing.T) {
	rule := applyRule("r1", "0", "0")
	rule.Selector = Selector{
		ProductVariantIds: []string{"var-a"}, ProductTemplateIds: []string{"tpl-b"}, CategoryIds: []string{"cat-c"},
	}
	lineA := withCategories(buyLine("a", 1, "1", "100"), "cat-c-child", "cat-c")
	lineB := withCategories(buyLine("b", 2, "1", "200"), "cat-c")
	lineAOutside := withCategories(buyLine("a2", 3, "1", "400"), "cat-x")
	lineAOutside.ProductVariantId = "var-a"
	lineD := withCategories(buyLine("d", 4, "1", "800"), "cat-c")

	result := Apply(orderOf([]Line{lineA, lineB, lineAOutside, lineD}, automatic("p1", 10, []Rule{rule},
		percentReward("rw", ScopeOrder, "10", "0"))))
	program := programResult(t, result, "p1")
	require.Equal(t, OutcomeApplied, program.Outcome)
	require.Equal(t, []string{"a", "b"}, program.Qualification.EligibleLineKeys)
	requireDec(t, "300", program.Qualification.EligibleAmount, "eligible amount")
}

// AC-09: base 100,000, tax 10%, threshold 110,000.
func TestTaxBasisThreshold(t *testing.T) {
	for _, testCase := range []struct {
		basis   string
		outcome string
	}{
		{TaxBasisIncluded, OutcomeApplied},
		{TaxBasisExcluded, OutcomeNotEligible},
	} {
		t.Run(testCase.basis, func(t *testing.T) {
			rule := applyRule("r1", "0", "110000")
			rule.AmountTaxBasis = testCase.basis
			lines := []Line{withTax(buyLine("l1", 1, "1", "100000"), "0.1")}
			result := Apply(orderOf(lines, automatic("p1", 10, []Rule{rule}, percentReward("rw", ScopeOrder, "10", "0"))))
			require.Equal(t, testCase.outcome, programResult(t, result, "p1").Outcome)
		})
	}
}

func TestGateCurrencyAndPricelist(t *testing.T) {
	reward := percentReward("rw", ScopeOrder, "10", "0")
	rules := []Rule{applyRule("r1", "0", "0")}
	priced := buyLine("l1", 1, "1", "1000")
	priced.PricelistId = "pl-other"

	code := automatic("code", 10, rules, reward)
	code.ProgramType = ProgramDiscountCode
	ownCoupon := automatic("own", 10, rules, reward)
	ownCoupon.ProgramType, ownCoupon.VoucherId, ownCoupon.VoucherSourceOrderId = ProgramNextOrderCoupon, "v1", "order-1"
	foreign := automatic("usd", 10, rules, reward)
	foreign.CurrencyId = "USD"
	scoped := automatic("pl", 10, rules, reward)
	scoped.PricelistIds = []string{"pl-linked"}
	unknown := automatic("odd", 10, rules, reward)
	unknown.ProgramType = "loyalty_card"

	result := Apply(orderOf([]Line{priced}, code, ownCoupon, foreign, scoped, unknown))
	for promotionId, reason := range map[string]string{
		"code": ReasonVoucherRequired, "own": ReasonVoucherSourceOrder, "usd": ReasonCurrencyMismatch,
		"pl": ReasonPricelistMismatch, "odd": ReasonUnknownProgramType,
	} {
		program := programResult(t, result, promotionId)
		require.NotEqual(t, OutcomeApplied, program.Outcome, promotionId)
		require.Equal(t, reason, program.ReasonCode, promotionId)
	}
}

func TestVoucherUnlocksButNeverBypassesConditions(t *testing.T) {
	coupon := automatic("cp", 10, []Rule{applyRule("r1", "0", "500")}, fixedReward("rw", ScopeOrder, "100"))
	coupon.ProgramType, coupon.VoucherId = ProgramCoupon, "v1"

	met := Apply(orderOf([]Line{buyLine("l1", 1, "1", "600")}, coupon))
	require.Equal(t, OutcomeApplied, programResult(t, met, "cp").Outcome)
	require.Equal(t, "v1", programResult(t, met, "cp").VoucherId)

	// AC-32 shape: a voucher whose apply phase is not met grants nothing.
	missed := Apply(orderOf([]Line{buyLine("l1", 1, "1", "400")}, coupon))
	require.Equal(t, OutcomeNotEligible, programResult(t, missed, "cp").Outcome)
}
