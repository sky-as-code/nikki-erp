package pricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

// Rule targeting, the resolution ladder and the three calculation methods. The package is a pure
// function, so every case is exercised for real: the inputs are the whole world the engine sees.

// aLine is a plain product line, placed in a two-level category tree.
func aLine() LineInput {
	return LineInput{
		Key:                "L1",
		ProductVariantId:   "VAR1",
		ProductTemplateId:  "TPL1",
		CategoryPath:       []string{"CAT_SOFT_DRINKS", "CAT_BEVERAGES"},
		UomId:              "UOM_BOTTLE",
		Quantity:           dec("1"),
		CatalogueUnitPrice: dec("100"),
	}
}

func fixedRule(id, appliesTo, targetId, price string) PricelistItem {
	item := PricelistItem{
		Id:                id,
		AppliesTo:         appliesTo,
		UomId:             "UOM_BOTTLE",
		UnitPrice:         dec(price),
		MinQuantity:       decimal.Zero,
		CalculationMethod: MethodFixedPrice,
	}
	switch appliesTo {
	case AppliesToVariant:
		item.ProductVariantId = targetId
	case AppliesToTemplate:
		item.ProductTemplateId = targetId
	case AppliesToCategory:
		item.ProductCategoryId = targetId
	}
	return item
}

// A variant rule beats a template rule.
func TestVariantRuleBeatsTemplateRule(t *testing.T) {
	items := []PricelistItem{
		fixedRule("R_TPL", AppliesToTemplate, "TPL1", "90"),
		fixedRule("R_VAR", AppliesToVariant, "VAR1", "80"),
	}

	best, found := bestRule(items, aLine())

	if !found || best.Id != "R_VAR" {
		t.Fatalf("the variant rule must win; got %+v (found=%v)", best.Id, found)
	}
}

// A template rule beats a category rule.
func TestTemplateRuleBeatsCategoryRule(t *testing.T) {
	items := []PricelistItem{
		fixedRule("R_CAT", AppliesToCategory, "CAT_SOFT_DRINKS", "90"),
		fixedRule("R_TPL", AppliesToTemplate, "TPL1", "85"),
	}

	best, found := bestRule(items, aLine())

	if !found || best.Id != "R_TPL" {
		t.Fatalf("the template rule must win; got %q (found=%v)", best.Id, found)
	}
}

// The nearest ancestor category wins: a rule on Soft Drinks beats one on its parent Beverages.
func TestNearestAncestorCategoryWins(t *testing.T) {
	items := []PricelistItem{
		fixedRule("R_PARENT", AppliesToCategory, "CAT_BEVERAGES", "70"),
		fixedRule("R_NEAR", AppliesToCategory, "CAT_SOFT_DRINKS", "90"),
	}

	best, found := bestRule(items, aLine())

	if !found || best.Id != "R_NEAR" {
		t.Fatalf("the nearer category must win even though it prices higher; got %q", best.Id)
	}
}

// Every category match must still outrank ALL_PRODUCTS, however distant the ancestor.
func TestDistantCategoryStillBeatsAllProducts(t *testing.T) {
	items := []PricelistItem{
		fixedRule("R_ALL", AppliesToAllProducts, "", "50"),
		fixedRule("R_FAR", AppliesToCategory, "CAT_BEVERAGES", "95"),
	}

	best, found := bestRule(items, aLine())

	if !found || best.Id != "R_FAR" {
		t.Fatalf("a category rule must beat ALL_PRODUCTS; got %q", best.Id)
	}
}

// The scope ladder outranks the target ladder: a point-scoped list saying "everything at 10% off"
// beats the global list's rule for this exact variant.
func TestPricelistSpecificityOutranksTargetSpecificity(t *testing.T) {
	global := fixedRule("R_VAR_GLOBAL", AppliesToVariant, "VAR1", "80")
	global.Specificity = 1
	local := fixedRule("R_ALL_LOCAL", AppliesToAllProducts, "", "95")
	local.Specificity = 3

	best, found := bestRule([]PricelistItem{global, local}, aLine())

	if !found || best.Id != "R_ALL_LOCAL" {
		t.Fatalf("the point-scoped list must win whatever either rule targets; got %q", best.Id)
	}
}

// The highest quantity break the line actually reaches wins.
func TestHighestReachedQuantityBreakWins(t *testing.T) {
	line := aLine()
	line.Quantity = dec("12")

	one := fixedRule("R_1", AppliesToVariant, "VAR1", "100")
	ten := fixedRule("R_10", AppliesToVariant, "VAR1", "90")
	ten.MinQuantity = dec("10")
	hundred := fixedRule("R_100", AppliesToVariant, "VAR1", "80")
	hundred.MinQuantity = dec("100")

	best, found := bestRule([]PricelistItem{one, ten, hundred}, line)

	if !found || best.Id != "R_10" {
		t.Fatalf("a line of 12 takes the ten-break, not the one- or hundred-break; got %q", best.Id)
	}
}

// Two rules alike in every ranked respect still resolve the same way every run.
func TestIdBreaksTheFinalTie(t *testing.T) {
	first := fixedRule("R_AAA", AppliesToVariant, "VAR1", "90")
	second := fixedRule("R_BBB", AppliesToVariant, "VAR1", "80")

	forward, _ := bestRule([]PricelistItem{first, second}, aLine())
	reversed, _ := bestRule([]PricelistItem{second, first}, aLine())

	if forward.Id != reversed.Id {
		t.Fatalf("resolution must not depend on input order: %q vs %q", forward.Id, reversed.Id)
	}
	if forward.Id != "R_AAA" {
		t.Fatalf("the lowest id wins the final tie; got %q", forward.Id)
	}
}

// Sequence is checked before id, and LOWEST wins — the opposite direction to Priority.
func TestLowestSequenceWins(t *testing.T) {
	late := fixedRule("R_AAA", AppliesToVariant, "VAR1", "90")
	late.Sequence = 10
	early := fixedRule("R_ZZZ", AppliesToVariant, "VAR1", "80")
	early.Sequence = 1

	best, _ := bestRule([]PricelistItem{late, early}, aLine())

	if best.Id != "R_ZZZ" {
		t.Fatalf("the lowest sequence wins even with a higher id; got %q", best.Id)
	}
}

// A rule targeting another product must not match at all.
func TestRuleForAnotherProductDoesNotMatch(t *testing.T) {
	items := []PricelistItem{fixedRule("R_OTHER", AppliesToVariant, "VAR_OTHER", "10")}

	if _, found := bestRule(items, aLine()); found {
		t.Fatal("a rule naming a different variant must not price this line")
	}
}

// A rule with no applies_to means the variant it names, and must not silently widen to the whole
// catalogue.
func TestMissingAppliesToIsReadAsVariant(t *testing.T) {
	legacy := PricelistItem{
		Id: "R_LEGACY", ProductVariantId: "VAR1", UomId: "UOM_BOTTLE",
		UnitPrice: dec("77"), MinQuantity: decimal.Zero,
	}

	best, found := bestRule([]PricelistItem{legacy}, aLine())
	if !found {
		t.Fatal("a legacy rule for this variant must still match")
	}

	price, computed := rulePrice(best, aLine(), InternalScale)
	if !computed || !price.Equal(dec("77")) {
		t.Fatalf("a legacy rule must still quote its price; got %s (computed=%v)", price, computed)
	}
}

// Base 100, discount 10% resolves to 90.
func TestDiscountMethod(t *testing.T) {
	rule := fixedRule("R_D", AppliesToVariant, "VAR1", "0")
	rule.CalculationMethod = MethodDiscount
	rule.DiscountPercent = dec("10")

	price, computed := rulePrice(rule, aLine(), InternalScale)

	if !computed || !price.Equal(dec("90")) {
		t.Fatalf("100 less 10%% is 90; got %s", price)
	}
}

// Cost 60 plus 50% resolves to 90, expressed as a negative discount.
func TestFormulaOnCostMarksUp(t *testing.T) {
	line := aLine()
	line.UnitCost, line.HasCost = dec("60"), true

	rule := fixedRule("R_F", AppliesToVariant, "VAR1", "0")
	rule.CalculationMethod = MethodFormula
	rule.BasePriceSource = BaseSourceCost
	rule.DiscountPercent = dec("-50")

	price, computed := rulePrice(rule, line, InternalScale)

	if !computed || !price.Equal(dec("90")) {
		t.Fatalf("cost 60 marked up 50%% is 90; got %s", price)
	}
}

// A COST formula with no cost available must decline rather than price at zero: zero is a real cost
// for a giveaway, so the number alone cannot say which case this is.
func TestFormulaOnMissingCostDeclines(t *testing.T) {
	rule := fixedRule("R_F", AppliesToVariant, "VAR1", "0")
	rule.CalculationMethod = MethodFormula
	rule.BasePriceSource = BaseSourceCost

	if _, computed := rulePrice(rule, aLine(), InternalScale); computed {
		t.Fatal("a formula with no cost must decline, not price at zero")
	}
}

// Rounding happens before the surcharge, so an amount meant to be exact stays exact.
func TestFormulaRoundsThenAddsSurcharge(t *testing.T) {
	line := aLine()
	line.CatalogueUnitPrice = dec("98765")

	rule := fixedRule("R_F", AppliesToVariant, "VAR1", "0")
	rule.CalculationMethod = MethodFormula
	rule.BasePriceSource = BaseSourceBaseSalesPrice
	rule.RoundingIncrement = dec("1000")
	rule.SurchargeAmount = dec("500")

	price, computed := rulePrice(rule, line, InternalScale)

	// 98,765 rounds to 99,000, then the exact 500 is added.
	if !computed || !price.Equal(dec("99500")) {
		t.Fatalf("expected 99500 (99000 rounded, plus an exact 500); got %s", price)
	}
}

// A rule must never produce a negative unit price: it would reach the totals as money owed to the
// customer.
func TestPriceNeverGoesNegative(t *testing.T) {
	rule := fixedRule("R_F", AppliesToVariant, "VAR1", "0")
	rule.CalculationMethod = MethodFormula
	rule.BasePriceSource = BaseSourceBaseSalesPrice
	rule.DiscountPercent = dec("100")
	rule.SurchargeAmount = dec("-50")

	price, computed := rulePrice(rule, aLine(), InternalScale)

	if !computed || price.IsNegative() {
		t.Fatalf("a price must not be negative; got %s", price)
	}
}

// A FIXED_PRICE rule quotes an amount PER unit, so it cannot price a line counted in another.
func TestFixedPriceRuleInAnotherUnitDoesNotMatch(t *testing.T) {
	rule := fixedRule("R_CASE", AppliesToVariant, "VAR1", "1000")
	rule.UomId = "UOM_CASE"

	if _, found := bestRule([]PricelistItem{rule}, aLine()); found {
		t.Fatal("a price per case must not be applied to a line counted in bottles")
	}
}

// A DISCOUNT rule carries no unit: it adjusts a base already in the line's unit.
func TestDiscountRuleAppliesRegardlessOfUnit(t *testing.T) {
	rule := fixedRule("R_D", AppliesToVariant, "VAR1", "0")
	rule.CalculationMethod = MethodDiscount
	rule.DiscountPercent = dec("10")
	rule.UomId = "UOM_CASE"

	if _, found := bestRule([]PricelistItem{rule}, aLine()); !found {
		t.Fatal("a percentage discount is unit-agnostic and must still apply")
	}
}
