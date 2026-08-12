package services

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Price selection for BR §6.12.
//
// Rule 1 — a price attaches to exactly one of a template or a variant — is enforced by the
// schema's exclusive_required_fields group, so nothing here re-checks it. What is left is rule 2,
// the read-time question: given every price row that mentions a product, which one applies?
//
// The answer is deliberately nullable. A product with no applicable rule has *no price*, which is
// not the same as a price of zero: a consumer that substitutes 0 sells the product for nothing.
// SelectApplicablePrice returns nil for that case and every caller must carry the nil through.

// SelectApplicablePrice picks the price that applies to a variant on the given day, from rows
// that target either the variant or its template.
//
// Precedence, per BR §6.12 rule 2:
//  1. rows that cannot apply at all are discarded — archived, not approved, or outside their
//     effective date range;
//  2. a row targeting the variant beats one targeting the template, however old it is, because a
//     variant rule exists precisely to override the line's base price;
//  3. among rows at the same level, the one that came into effect most recently wins. A row with
//     no effective_from is the standing price and loses to any dated row, since a dated row is a
//     deliberate later decision.
//
// Returns nil when no row applies.
func SelectApplicablePrice(
	prices []*models.ProductPrice,
	variantId string,
	templateId string,
	on time.Time,
) *decimal.Decimal {
	var best *models.ProductPrice
	var bestTargetsVariant bool

	for _, price := range prices {
		if !isPriceApplicable(price, on) {
			continue
		}

		targetsVariant, matches := priceTarget(price, variantId, templateId)
		if !matches {
			continue
		}

		if best == nil || winsOver(price, targetsVariant, best, bestTargetsVariant) {
			best, bestTargetsVariant = price, targetsVariant
		}
	}

	if best == nil {
		return nil
	}
	return best.GetPrice()
}

// winsOver reports whether candidate should replace best. Variant level beats template level
// outright; within a level the later effective_from wins.
func winsOver(
	candidate *models.ProductPrice, candidateTargetsVariant bool,
	best *models.ProductPrice, bestTargetsVariant bool,
) bool {
	if candidateTargetsVariant != bestTargetsVariant {
		return candidateTargetsVariant
	}
	return startsLaterThan(candidate.GetEffectiveFrom(), best.GetEffectiveFrom())
}

// startsLaterThan orders two effective_from bounds. A nil bound is the standing price, in effect
// since forever, so it sorts before any dated row.
func startsLaterThan(candidate *model.ModelDate, best *model.ModelDate) bool {
	if candidate == nil {
		return false
	}
	if best == nil {
		return true
	}
	return candidate.After(*best)
}

// priceTarget reports which level a row applies to, and whether it concerns this product at all.
func priceTarget(price *models.ProductPrice, variantId string, templateId string) (targetsVariant bool, matches bool) {
	if id := derefModelId(price.GetProductVariantId()); id != "" {
		return true, variantId != "" && id == variantId
	}
	if id := derefModelId(price.GetProductTemplateId()); id != "" {
		return false, templateId != "" && id == templateId
	}
	// Rule 1 makes this unreachable through the engine; a row loaded from elsewhere that targets
	// nothing is simply ignored rather than treated as applying to everything.
	return false, false
}

// isPriceApplicable reports whether a row is in force on the given day.
func isPriceApplicable(price *models.ProductPrice, on time.Time) bool {
	if price == nil {
		return false
	}
	if derefBool(price.IsArchived()) {
		return false
	}

	// Only an approved row prices a product. A draft is a price being prepared, and an expired one
	// is kept for audit; applying either would charge a customer a price nobody signed off.
	status := price.GetStatus()
	if status == nil || *status != models.ProductPriceStatusApproved {
		return false
	}

	if from := price.GetEffectiveFrom(); from != nil && from.GoTime().After(on) {
		return false
	}
	// effective_to is the last day the price applies, so the comparison is against the end of that
	// day rather than its midnight — otherwise a price expires a day early.
	if to := price.GetEffectiveTo(); to != nil && endOfDay(to.GoTime()).Before(on) {
		return false
	}
	return true
}

func endOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

func derefModelId(id *model.Id) string {
	if id == nil {
		return ""
	}
	return string(*id)
}
