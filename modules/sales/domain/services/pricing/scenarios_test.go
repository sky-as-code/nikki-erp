package pricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

// The named pricing scenarios, asserted here for traceability. Some are also covered by the
// rule-driven tests elsewhere; they are re-asserted here rather than cross-referenced, because a
// comment pointing at another test does not fail when that test is renamed or narrowed.

// A base sales price of 100 with no rule matching resolves to 100. The catalogue price standing is
// what makes a pricelist optional; resolving to zero would give unmatched products away.
func TestScenarioBaseSalesPriceAppliesWhenNoRuleMatches(t *testing.T) {
	result := Calculate(Input{
		Lines:   []LineInput{lineOf("a", 1, "1", "100")},
		Context: vndContext(),
	})

	line := lineByKey(t, result, "a")
	if !line.EffectiveUnitPrice.Equal(dec("100")) {
		t.Fatalf("effective unit price = %s, want the base sales price 100", line.EffectiveUnitPrice)
	}
	if line.PricingSource != "catalogue" {
		t.Fatalf("pricing_source = %q, want catalogue — the price came from the product", line.PricingSource)
	}
}

// Base 100 less a 10% discount resolves to 90.
func TestScenarioSalesDiscount(t *testing.T) {
	item := fixedRule("R", AppliesToVariant, "VAR1", "0")
	item.CalculationMethod = MethodDiscount
	item.DiscountPercent = dec("10")

	price, ok := rulePrice(item, aLine(), InternalScale)

	if !ok || !price.Equal(dec("90")) {
		t.Fatalf("100 less 10%% = %s (ok=%v), want 90", price, ok)
	}
}

// A cost of 60 plus 50% resolves to 90, and Sales does not touch the cost: a formula READS it, and
// writing back a "selling cost" would make the margin computed from it meaningless.
func TestScenarioSalesFormulaOnCost(t *testing.T) {
	line := aLine()
	line.UnitCost = dec("60")
	line.HasCost = true

	item := fixedRule("R", AppliesToVariant, "VAR1", "0")
	item.CalculationMethod = MethodFormula
	item.BasePriceSource = BaseSourceCost
	// A markup is a negative discount: +50% on the base.
	item.DiscountPercent = dec("-50")

	price, ok := rulePrice(item, line, InternalScale)

	if !ok || !price.Equal(dec("90")) {
		t.Fatalf("cost 60 plus 50%% = %s (ok=%v), want 90", price, ok)
	}
	if !line.UnitCost.Equal(dec("60")) {
		t.Fatalf("the cost was modified to %s; pricing must never write it back", line.UnitCost)
	}
}

// A formula whose cost is unavailable does not price at zero. Zero is a legitimate cost for a
// giveaway, so the number alone cannot say whether one was configured — hence HasCost.
func TestScenarioFormulaWithNoCostDoesNotPriceAtZero(t *testing.T) {
	line := aLine()
	line.HasCost = false

	item := fixedRule("R", AppliesToVariant, "VAR1", "0")
	item.CalculationMethod = MethodFormula
	item.BasePriceSource = BaseSourceCost
	item.DiscountPercent = dec("-50")

	_, ok := rulePrice(item, line, InternalScale)

	if ok {
		t.Fatal("a formula with no cost must not resolve; the catalogue price stands instead")
	}
}

// A cost that IS zero must still price: the other half of the test above, and why HasCost exists.
func TestScenarioAZeroCostStillPrices(t *testing.T) {
	line := aLine()
	line.UnitCost = decimal.Zero
	line.HasCost = true

	item := fixedRule("R", AppliesToVariant, "VAR1", "0")
	item.CalculationMethod = MethodFormula
	item.BasePriceSource = BaseSourceCost
	item.DiscountPercent = decimal.Zero

	price, ok := rulePrice(item, line, InternalScale)

	if !ok || !price.IsZero() {
		t.Fatalf("a configured cost of zero must price at zero; got %s (ok=%v)", price, ok)
	}
}

// A rule from an archived pricelist must not price a line. Archived rules are excluded by the
// CALLER; this asserts the consequence, that the line falls back to the catalogue price.
func TestScenarioArchivedRuleLeavesTheCataloguePriceStanding(t *testing.T) {
	result := Calculate(Input{
		// The archived list's rule is simply absent, as the repository read produces.
		Lines:   []LineInput{lineOf("a", 1, "1", "100")},
		Context: vndContext(),
	})

	line := lineByKey(t, result, "a")
	if !line.EffectiveUnitPrice.Equal(dec("100")) {
		t.Fatalf("unit price = %s, want 100 — an archived rule must not price, and must not zero it either",
			line.EffectiveUnitPrice)
	}
}

// A rule targeting another organization's product must not match. Org scoping is enforced by the
// repository read; what is asserted here is the property making that safe — a rule whose target does
// not match the line is not a candidate at all.
func TestScenarioARuleForAnotherProductNeverMatches(t *testing.T) {
	items := []PricelistItem{
		fixedRule("R_OTHER_VARIANT", AppliesToVariant, "VAR_ELSEWHERE", "1"),
		fixedRule("R_OTHER_TEMPLATE", AppliesToTemplate, "TPL_ELSEWHERE", "1"),
		fixedRule("R_OTHER_CATEGORY", AppliesToCategory, "CAT_ELSEWHERE", "1"),
	}

	_, found := bestRule(items, aLine())

	if found {
		t.Fatal("no rule names this product, so none may price it")
	}
}
