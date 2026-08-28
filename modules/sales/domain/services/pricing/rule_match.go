package pricing

import (
	"github.com/shopspring/decimal"
)

// How a pricelist rule is matched to a line, and how the winner turns into a price (sections 12–14,
// 18).
//
// Two independent ladders decide which rule wins, and keeping them apart is the point:
//
//   - PRICELIST specificity — which policy applies HERE, from the list's channel and point scope.
//     The caller computes it, because it needs the channel and the point.
//   - TARGET specificity — which rule within that policy applies to THIS PRODUCT: variant, then
//     template, then the nearest ancestor category, then everything.
//
// Pricelist specificity outranks target specificity, and that ordering is a business decision
// rather than an implementation detail. A point-scoped list saying "everything at 10% off" must
// beat the global list's rule for this exact variant, because the point-scoped list is the one that
// applies at this till. Reversing them would let a global variant rule silently override every
// local price, which is what the scope mechanism exists to prevent.

// Target specificity ranks, most specific first. Only their ORDER is meaningful; the numbers are
// compared and never stored.
const (
	targetRankVariant  = 4
	targetRankTemplate = 3
	targetRankCategory = 2 // Refined by category depth; see categoryRank.
	targetRankAll      = 1
	targetRankNoMatch  = 0
)

// categoryDepthScale spreads category matches between the category and template ranks.
//
// A category match's rank is targetRankCategory plus a fraction of a rank derived from how near the
// matched category is to the product. That keeps every category match above ALL_PRODUCTS and below
// a template match, while still ordering them among themselves — the nearest ancestor winning
// (PRICE-INV-017).
const categoryDepthScale = 1000

// ruleTargetRank scores how specifically a rule targets a line, or targetRankNoMatch if it does not
// apply to it at all.
//
// An empty AppliesTo is read as PRODUCT_VARIANT, which is what every rule written before targeting
// existed meant: they all named a variant, and treating a missing enum as the broadest target would
// silently widen them to the whole catalogue.
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

// categoryRank scores a category match by how near the matched category is to the product.
//
// CategoryPath runs from the product's own category outward to the root, so an earlier position is
// nearer and must win. The score therefore counts DOWN from the top of the category band as the
// match gets more distant, leaving every category match inside its band.
//
// A path longer than the scale would collide with the band below. Rather than let that corrupt the
// ordering silently, distant ancestors past the scale all score at the bottom of the band: they are
// still category matches and still beat ALL_PRODUCTS, they simply stop being ordered among
// themselves. A category tree that deep is not a real one.
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

// appliesToOf reads the rule's target, defaulting to the variant target for a rule written before
// the field existed.
func appliesToOf(item PricelistItem) string {
	if item.AppliesTo == "" {
		return AppliesToVariant
	}
	return item.AppliesTo
}

// ruleApplies reports whether a rule may price this line at all, before any ranking.
//
// The unit check is deliberately conditional. A FIXED_PRICE rule quotes an amount PER a unit, so a
// rule in cases cannot price a line in bottles — the number would mean something else. A DISCOUNT
// or FORMULA rule carries no unit of its own: it adjusts a base that is already in the line's unit,
// so it applies whatever the line is counted in.
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

// methodOf reads the rule's calculation method, defaulting to FIXED_PRICE.
//
// Every row written before this change request quoted an amount in `price` and meant exactly that,
// so the default preserves their behaviour without rewriting them.
func methodOf(item PricelistItem) string {
	if item.CalculationMethod == "" {
		return MethodFixedPrice
	}
	return item.CalculationMethod
}

// betterRule reports whether candidate beats incumbent. Both are already known to apply.
//
// The order is the requirement's, and every step matters:
//
//  1. pricelist specificity — which policy applies here;
//  2. pricelist priority — the tie-break WITHIN that, higher winning;
//  3. target specificity — variant over template over nearest category over everything;
//  4. quantity break — the highest the line actually reaches, so a line of 12 takes the ten-break
//     rather than the one-break when both are eligible;
//  5. sequence, LOWEST winning;
//  6. id, so two otherwise identical rules still resolve the same way on every run.
//
// Step 6 is what makes the outcome testable at all (PRICE-INV-020): without a final total order,
// two indistinguishable rules would resolve by whatever order the database happened to return.
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

// rulePrice computes what a matched rule charges, and whether it could compute anything at all.
//
// The false return is not an error but an absence: a FORMULA whose base cannot be resolved has no
// price to offer, and the caller falls back to the base sales price exactly as it would if no rule
// had matched. Reporting it as a failure would refuse the sale over a misconfigured rule, which is
// a worse answer than pricing at the catalogue price.
func rulePrice(item PricelistItem, line LineInput, scale int32) (decimal.Decimal, bool) {
	switch methodOf(item) {
	case MethodFixedPrice:
		return item.UnitPrice.Round(scale), true

	case MethodDiscount:
		// A discount is always off the line's own base price. It never reads cost or another list:
		// those are FORMULA's business, and a DISCOUNT rule that could reach them would make the
		// two methods indistinguishable.
		return applyDiscount(line.CatalogueUnitPrice, item.DiscountPercent).Round(scale), true

	case MethodFormula:
		base, ok := formulaBase(item, line)
		if !ok {
			return decimal.Zero, false
		}
		// Order fixed by section 13: discount, then round, then surcharge. Surcharge last so an
		// amount meant to be exact is not rounded into something that is not it.
		price := applyDiscount(base, item.DiscountPercent)
		price = roundToIncrement(price, item.RoundingIncrement)
		price = price.Add(item.SurchargeAmount)
		return floorAtZero(price).Round(scale), true
	}
	return decimal.Zero, false
}

// formulaBase resolves what a FORMULA rule starts from.
//
// OTHER_PRICELIST is resolved by the CALLER and handed back in ResolvedBasePrice: chasing it here
// would mean reading another list, and this package reads nothing. A cycle among lists is refused
// when the rule is written, not here.
func formulaBase(item PricelistItem, line LineInput) (decimal.Decimal, bool) {
	switch item.BasePriceSource {
	case BaseSourceCost:
		// An unset cost must NOT price at zero. Zero is a real cost for a giveaway, so the two are
		// indistinguishable in the number alone — which is why HasCost exists.
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
		// Empty defaults to the base sales price: it is the only source needing nothing but the
		// line itself, so a rule that names none still resolves rather than failing.
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

// floorAtZero keeps a computed price from going negative.
//
// A discount over 100% combined with a negative surcharge can produce one, and a negative unit
// price would flow into the order totals as money owed TO the customer — a refund arising from a
// pricing rule, which no rule should be able to create.
func floorAtZero(value decimal.Decimal) decimal.Decimal {
	if value.IsNegative() {
		return decimal.Zero
	}
	return value
}
