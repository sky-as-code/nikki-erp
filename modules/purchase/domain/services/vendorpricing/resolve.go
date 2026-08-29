package vendorpricing

// Resolve picks the vendor price that applies, or reports that none does. The precedence, in order:
//
//  1. Applicable: the row's window covers the pricing date and its unit could be converted to, both
//     being the caller's verdicts.
//  2. Product: this exact variant or a template-wide row; a row for another variant prices
//     something else.
//  3. Quantity: the row's break must be one the request reaches, measured in that row's own unit.
//  4. Specificity: a variant-specific row beats a template-wide one.
//  5. Quantity break: within the same specificity, the highest break reached wins, so 120 takes the
//     100+ price rather than the 10+ one.
//  6. Sequence: lowest wins.
//  7. Id: a final total order, so two indistinguishable rows resolve identically rather than by
//     whatever order the database returned.
//
// A false return means no vendor price applies, and the caller must not substitute one from
// elsewhere: no fallback to product cost, none to another vendor. A user with the entitlement may
// type a price on the line, which is a deliberate human act.
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
		// The request's variant, not the candidate's: a template-wide row prices a specific variant,
		// and echoing the row's empty variant would lose which product was priced.
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
	// Only an exact variant match or a template-wide row (empty variant) is a candidate.
	if candidate.ProductVariantId != "" &&
		candidate.ProductVariantId != request.ProductVariantId {
		return false
	}

	// The break is compared in the candidate's own unit. A unit missing from the map means the
	// caller could not convert into it, which is not a quantity of zero: comparing against zero
	// would make every break look reachable.
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
