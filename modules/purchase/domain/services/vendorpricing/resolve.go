package vendorpricing

// Resolve picks the vendor price that applies, or reports that none does (section 27).
//
// The ladder, in order, and every step matters:
//
//  1. APPLICABLE — the row's commercial window covers the pricing date, and its unit could be
//     converted to. Both are the caller's verdicts; see Candidate.Applicable and
//     Request.QuantityByUom.
//  2. PRODUCT — a row for this exact variant, or a template-wide row. A row for a DIFFERENT
//     variant of the same template is not a candidate at all: it prices something else.
//  3. QUANTITY — the row's break must be one the request actually reaches, measured in that row's
//     own unit (BR-PRICE-UOM-004).
//  4. SPECIFICITY — a variant-specific row beats a template-wide one (PRICE-INV-018).
//  5. QUANTITY BREAK — within the same specificity, the highest break reached wins, so a request
//     for 120 takes the 100+ price rather than the 1+ or 10+ one.
//  6. SEQUENCE — lowest wins.
//  7. ID — so two otherwise indistinguishable rows still resolve identically on every run
//     (PRICE-INV-020). Without a final total order the answer would depend on whatever order the
//     database happened to return, and the outcome would not be testable.
//
// The false return is "no vendor price applies", and the caller must NOT convert it into a price
// from somewhere else. Section 28 is explicit: no fallback to Product.cost, none to another
// vendor. A user with the entitlement may type a price on the order line; that is a deliberate
// human act, not a substitution the system makes quietly.
func Resolve(request Request, candidates []Candidate) (Resolution, bool) {
	var best Candidate
	found := false

	for _, candidate := range candidates {
		if !candidateApplies(request, candidate) {
			continue
		}
		if !found || better(candidate, best) {
			best, found = candidate, true
		}
	}

	if !found {
		return Resolution{}, false
	}

	return Resolution{
		VendorProductPriceId: best.Id,
		// The REQUEST's variant, not the candidate's: a template-wide row prices a specific
		// variant, and reporting the row's empty variant back would lose which product was priced.
		ProductVariantId: request.ProductVariantId,
		UnitPrice:        best.UnitPrice,
		PurchaseUomId:    best.PurchaseUomId,
		CurrencyId:       best.CurrencyId,
		LeadTimeDays:     best.LeadTimeDays,
	}, true
}

// candidateApplies reports whether a row may price this request at all, before any ranking.
func candidateApplies(request Request, candidate Candidate) bool {
	if !candidate.Applicable {
		return false
	}
	if candidate.ProductTemplateId != request.ProductTemplateId {
		return false
	}
	// A row naming a DIFFERENT variant prices a different product. Only an exact match or a
	// template-wide row (empty variant) is a candidate.
	if candidate.ProductVariantId != "" &&
		candidate.ProductVariantId != request.ProductVariantId {
		return false
	}

	// The break is compared in the candidate's OWN unit (BR-PRICE-UOM-004). A unit missing from
	// the map means the caller could not convert into it, which is not the same as a quantity of
	// zero — comparing against zero would make every break look reachable.
	quantity, converted := request.QuantityByUom[candidate.PurchaseUomId]
	if !converted {
		return false
	}
	return !quantity.LessThan(candidate.MinQuantity)
}

// better reports whether candidate beats incumbent. Both already apply.
func better(candidate, incumbent Candidate) bool {
	candidateSpecific := candidate.ProductVariantId != ""
	incumbentSpecific := incumbent.ProductVariantId != ""
	if candidateSpecific != incumbentSpecific {
		return candidateSpecific
	}
	if !candidate.MinQuantity.Equal(incumbent.MinQuantity) {
		return candidate.MinQuantity.GreaterThan(incumbent.MinQuantity)
	}
	if candidate.Sequence != incumbent.Sequence {
		return candidate.Sequence < incumbent.Sequence
	}
	return candidate.Id < incumbent.Id
}
