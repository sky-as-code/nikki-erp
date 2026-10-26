package pricing

import (
	"github.com/shopspring/decimal"
)

// How a pricelist rule is matched to a line, and how the winner turns into a price. Two ladders
// decide: PRICELIST specificity (which policy applies here, from the list's channel/point scope,
// computed by the caller) outranks TARGET specificity (variant, then template, then nearest ancestor
// category, then everything). Reversing them would let a global variant rule override every local
// price, which is exactly what scoping exists to prevent.

// Target specificity ranks, most specific first. Only their order is meaningful; never stored.
const (
	targetRankVariant  = 4
	targetRankTemplate = 3
	targetRankCategory = 2 // Refined by category depth; see categoryRank.
	targetRankAll      = 1
	targetRankNoMatch  = 0
)

// categoryDepthScale spreads category matches between the category and template ranks: a category
// match scores targetRankCategory plus a nearness fraction, so it stays above ALL_PRODUCTS and below
// a template match while the nearest ancestor still wins among categories.
const categoryDepthScale = 1000

// ruleTargetRank scores how specifically a rule targets a line, or targetRankNoMatch if it does not
// apply. An empty AppliesTo reads as PRODUCT_VARIANT: rules predating the field all named a variant,
// and defaulting to the broadest target would silently widen them to the whole catalogue.
func ruleTargetRank(item PricelistItem, line LineInput) int {
	switch appliesToOf(item) {
	case AppliesToVariant:
		if item.ProductVariantId != "" && item.ProductVariantId == line.ProductVariantId {
			return targetRankVariant * categoryDepthScale
		}
	case AppliesToTemplate:
		if item.ProductTemplateId != "" && item.ProductTemplateId == line.ProductTemplateId {
			return targetRankTemplate * categoryDepthScale
		}
	case AppliesToCategory:
		return categoryRank(item.ProductCategoryId, line.CategoryPath)
	case AppliesToAllProducts:
		return targetRankAll * categoryDepthScale
	}
	return targetRankNoMatch
}

// categoryRank scores a category match by nearness: CategoryPath runs from the product's own
// category outward to the root, so an earlier position wins and the score counts down from the top
// of the category band. Ancestors beyond categoryDepthScale all score at the bottom of the band
// rather than colliding with the band below — they still beat ALL_PRODUCTS, just stop being ordered
// among themselves.
func categoryRank(ruleCategoryId string, categoryPath []string) int {
	if ruleCategoryId == "" {
		return targetRankNoMatch
	}
	for distance, categoryId := range categoryPath {
		if categoryId != ruleCategoryId {
			continue
		}
		nearness := categoryDepthScale - 1 - distance
		if nearness < 1 {
			nearness = 1
		}
		return targetRankCategory*categoryDepthScale + nearness
	}
	return targetRankNoMatch
}

// appliesToOf reads the rule's target, defaulting to the variant target.
func appliesToOf(item PricelistItem) string {
	if item.AppliesTo == "" {
		return AppliesToVariant
	}
	return item.AppliesTo
}

// ruleApplies reports whether a rule may price this line at all, before any ranking. The unit check
// is deliberately conditional: a FIXED_PRICE rule quotes an amount PER unit, so a rule in cases
// cannot price a line in bottles, while DISCOUNT and FORMULA carry no unit and adjust a base already
// in the line's unit.
func ruleApplies(item PricelistItem, line LineInput) bool {
	if ruleTargetRank(item, line) == targetRankNoMatch {
		return false
	}
	if line.Quantity.LessThan(item.MinQuantity) {
		return false
	}
	if methodOf(item) == MethodFixedPrice && item.UomId != "" && item.UomId != line.UomId {
		return false
	}
	return true
}

// methodOf reads the rule's calculation method, defaulting to FIXED_PRICE to preserve the behaviour
// of rows that predate the field.
func methodOf(item PricelistItem) string {
	if item.CalculationMethod == "" {
		return MethodFixedPrice
	}
	return item.CalculationMethod
}

// betterRule reports whether candidate beats incumbent; both are already known to apply. The ladder
// is pricelist specificity, pricelist priority (higher wins), target specificity, highest quantity
// break the line reaches, sequence (LOWEST wins), then id. The final id step makes the outcome a
// total order, so two indistinguishable rules do not resolve by database row order.
func betterRule(candidate, incumbent PricelistItem, line LineInput) bool {
	if candidate.Specificity != incumbent.Specificity {
		return candidate.Specificity > incumbent.Specificity
	}
	if candidate.Priority != incumbent.Priority {
		return candidate.Priority > incumbent.Priority
	}

	candidateRank, incumbentRank := ruleTargetRank(candidate, line), ruleTargetRank(incumbent, line)
	if candidateRank != incumbentRank {
		return candidateRank > incumbentRank
	}
	if !candidate.MinQuantity.Equal(incumbent.MinQuantity) {
		return candidate.MinQuantity.GreaterThan(incumbent.MinQuantity)
	}
	if candidate.Sequence != incumbent.Sequence {
		return candidate.Sequence < incumbent.Sequence
	}
	return candidate.Id < incumbent.Id
}

// bestRule picks the rule that prices a line, or reports that none does.
func bestRule(items []PricelistItem, line LineInput) (PricelistItem, bool) {
	var best PricelistItem
	found := false

	for _, item := range items {
		if !ruleApplies(item, line) {
			continue
		}
		if !found || betterRule(item, best, line) {
			best, found = item, true
		}
	}
	return best, found
}

// rulePrice computes what a matched rule charges. The false return is an absence, not an error: a
// FORMULA whose base cannot be resolved leaves the caller falling back to the catalogue price rather
// than refusing the sale over a misconfigured rule.
func rulePrice(item PricelistItem, line LineInput, scale int32) (decimal.Decimal, bool) {
	switch methodOf(item) {
	case MethodFixedPrice:
		return item.UnitPrice.Round(scale), true

	case MethodDiscount:
		// A discount is always off the line's own base price; cost and other lists are FORMULA's
		// business.
		return applyDiscount(line.CatalogueUnitPrice, item.DiscountPercent).Round(scale), true

	case MethodFormula:
		base, ok := formulaBase(item, line)
		if !ok {
			return decimal.Zero, false
		}
		// Fixed order: discount, round, then surcharge — surcharge last so an exact amount stays exact.
		price := applyDiscount(base, item.DiscountPercent)
		price = roundToIncrement(price, item.RoundingIncrement)
		price = price.Add(item.SurchargeAmount)
		return floorAtZero(price).Round(scale), true
	}
	return decimal.Zero, false
}

// formulaBase resolves what a FORMULA rule starts from. OTHER_PRICELIST is resolved by the caller
// into ResolvedBasePrice, since this package reads nothing; list cycles are refused at write time.
func formulaBase(item PricelistItem, line LineInput) (decimal.Decimal, bool) {
	switch item.BasePriceSource {
	case BaseSourceCost:
		// An unset cost must NOT price at zero; zero is a real cost for a giveaway, hence HasCost.
		if !line.HasCost {
			return decimal.Zero, false
		}
		return line.UnitCost, true

	case BaseSourceOtherPricelist:
		if !item.HasResolvedBase {
			return decimal.Zero, false
		}
		return item.ResolvedBasePrice, true

	case BaseSourceBaseSalesPrice, "":
		// Empty defaults to the base sales price, the only source needing nothing but the line itself.
		return line.CatalogueUnitPrice, true
	}
	return decimal.Zero, false
}

// applyDiscount takes a signed percentage off a base. Negative marks up.
func applyDiscount(base decimal.Decimal, percent decimal.Decimal) decimal.Decimal {
	if percent.IsZero() {
		return base
	}
	multiplier := decimal.NewFromInt(1).Sub(percent.Div(decimal.NewFromInt(100)))
	return base.Mul(multiplier)
}

// roundToIncrement snaps a price to a commercial step: 1000 to the nearest thousand, 0.05 to the
// nearest five cents. A zero or negative increment means no rounding.
func roundToIncrement(value decimal.Decimal, increment decimal.Decimal) decimal.Decimal {
	if increment.LessThanOrEqual(decimal.Zero) {
		return value
	}
	return value.Div(increment).Round(0).Mul(increment)
}

// floorAtZero keeps a computed price from going negative — a discount over 100% plus a negative
// surcharge can produce one, and that would reach the order totals as money owed to the customer.
func floorAtZero(value decimal.Decimal) decimal.Decimal {
	if value.IsNegative() {
		return decimal.Zero
	}
	return value
}
