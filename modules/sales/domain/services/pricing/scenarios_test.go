package pricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

// The change request's own named scenarios (section 43), asserted here by name.
//
// Several are already covered by tests written from the rules rather than from the scenario list —
// TS-PRICE-02 and -03 in rule_match_test.go, TS-PRICE-01 as TestCataloguePriceWhenNoPricelistMatches
// in engine_test.go. What this file adds is the ones with no home, plus a named entry point for the
// ones that had one, so that somebody auditing the requirement can find each scenario by its
// identifier instead of inferring which test happens to cover it.
//
// Traceability is the point. A scenario that is genuinely covered elsewhere is asserted here too
// rather than merely referenced, because a comment pointing at another test does not fail when that
// test is renamed or narrowed.

// TS-PRICE-01: a base sales price of 100 with no rule matching resolves to 100.
//
// The catalogue price standing when nothing matches is what makes a pricelist optional. If an
// unmatched line resolved to zero, a product with no rule would be given away, and every new
// product would need a rule before it could be sold at all.
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

// TS-PRICE-02: base 100 less a 10% discount resolves to 90.
func TestScenarioSalesDiscount(t *testing.T) {
	item := fixedRule("R", AppliesToVariant, "VAR1", "0")
	item.CalculationMethod = MethodDiscount
	item.DiscountPercent = dec("10")

	price, ok := rulePrice(item, aLine(), InternalScale)

	if !ok || !price.Equal(dec("90")) {
		t.Fatalf("100 less 10%% = %s (ok=%v), want 90", price, ok)
	}
}

// TS-PRICE-03: a cost of 60 plus 50% resolves to 90, and Sales does not touch the cost.
//
// The second half is the half that matters. A formula READS the cost; writing back a "selling cost"
// would make the margin computed from it meaningless, and the engine has no way to write anywhere
// anyway — which is exactly why the calculation lives in a pure function.
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

// TS-PRICE-10, on the SALES side: a formula whose cost is unavailable does not price at zero.
//
// Zero is a legitimate cost for a giveaway, so the number alone cannot say whether one was
// configured — hence HasCost beside UnitCost. Without that distinction an unset cost would price
// every line in the rule's scope at a 100% markup on nothing.
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

// A cost that IS zero must still price, because a giveaway is a real configuration. This is the
// other half of the test above, and the pair is what makes HasCost worth having.
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

// TS-PRICE-12: a rule from an archived pricelist must not price a line.
//
// Archived rules are excluded by the CALLER, which reads only live lists — the engine is handed
// candidates and ranks them. What this asserts is the consequence: with the archived rule filtered
// out, the line falls back to its catalogue price rather than to nothing.
func TestScenarioArchivedRuleLeavesTheCataloguePriceStanding(t *testing.T) {
	result := Calculate(Input{
		// The archived list's rule is simply absent, which is what the repository read produces.
		Lines:   []LineInput{lineOf("a", 1, "1", "100")},
		Context: vndContext(),
	})

	line := lineByKey(t, result, "a")
	if !line.EffectiveUnitPrice.Equal(dec("100")) {
		t.Fatalf("unit price = %s, want 100 — an archived rule must not price, and must not zero it either",
			line.EffectiveUnitPrice)
	}
}

// TS-PRICE-11: a rule targeting another organization's product must not match.
//
// Org scoping is enforced by the repository read, not here — the engine never sees another org's
// rows. What CAN be asserted at this level is the property that makes that safe: a rule whose
// target does not match the line is not a candidate at all, so a row that slipped through would
// still have to name this exact product to affect it.
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
