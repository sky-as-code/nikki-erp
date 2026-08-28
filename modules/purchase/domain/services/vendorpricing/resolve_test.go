package vendorpricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

// Purchase price resolution (section 27), exercised for real: the package is a pure function, so
// the inputs below are the whole world it sees.

const (
	tpl     = "TPL1"
	varOne  = "VAR1"
	varTwo  = "VAR2"
	uomCase = "UOM_CASE"
	uomPc   = "UOM_PC"
)

func dec(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}

// row is a template-wide, currently-applicable candidate; the tests narrow it as needed.
func row(id, price, minQty string) Candidate {
	return Candidate{
		Id:                id,
		ProductTemplateId: tpl,
		PurchaseUomId:     uomCase,
		CurrencyId:        "VND",
		MinQuantity:       dec(minQty),
		UnitPrice:         dec(price),
		Applicable:        true,
		LeadTimeDays:      3,
	}
}

func request(quantity string) Request {
	return Request{
		ProductTemplateId: tpl,
		ProductVariantId:  varOne,
		QuantityByUom:     map[string]decimal.Decimal{uomCase: dec(quantity)},
	}
}

// TS-PRICE-06, the change request's own worked example: breaks at 1/10/100 priced 250/240/220, a
// quantity of 120, and the answer must be 220.
func TestHighestReachedBreakWins(t *testing.T) {
	candidates := []Candidate{
		row("R1", "250", "1"),
		row("R10", "240", "10"),
		row("R100", "220", "100"),
	}

	got, found := Resolve(request("120"), candidates)

	if !found || !got.UnitPrice.Equal(dec("220")) {
		t.Fatalf("a quantity of 120 takes the 100+ break; got %s (found=%v)", got.UnitPrice, found)
	}
}

// A quantity below every break resolves to nothing rather than to the cheapest row.
func TestQuantityBelowEveryBreakFindsNothing(t *testing.T) {
	candidates := []Candidate{row("R10", "240", "10"), row("R100", "220", "100")}

	if _, found := Resolve(request("5"), candidates); found {
		t.Fatal("no break is reached, so no price applies")
	}
}

// PRICE-INV-018: a variant-specific row beats a template-wide one, even when the template row has
// the higher quantity break — specificity is checked before quantity.
func TestVariantBeatsTemplate(t *testing.T) {
	templateWide := row("R_TPL", "220", "100")
	variantSpecific := row("R_VAR", "235", "1")
	variantSpecific.ProductVariantId = varOne

	got, found := Resolve(request("120"), []Candidate{templateWide, variantSpecific})

	if !found || got.VendorProductPriceId != "R_VAR" {
		t.Fatalf("the variant-specific row must win; got %q", got.VendorProductPriceId)
	}
}

// A row naming a different variant prices a different product and must not be considered at all.
func TestRowForAnotherVariantIsNotACandidate(t *testing.T) {
	other := row("R_OTHER", "100", "1")
	other.ProductVariantId = varTwo

	if _, found := Resolve(request("120"), []Candidate{other}); found {
		t.Fatal("a price for another variant must not price this one")
	}
}

// The window verdict is the caller's; an inapplicable row is simply skipped. This covers both an
// expired price and a future one, which is why the flag is a boolean rather than two dates.
func TestInapplicableRowsAreSkipped(t *testing.T) {
	expired := row("R_OLD", "180", "1")
	expired.Applicable = false
	live := row("R_NOW", "250", "1")

	got, found := Resolve(request("120"), []Candidate{expired, live})

	if !found || got.VendorProductPriceId != "R_NOW" {
		t.Fatalf("only the applicable row may price; got %q", got.VendorProductPriceId)
	}
}

// BR-PRICE-UOM-004: a break is compared in its OWN unit. Here the request converts to 2 cases or
// 48 pieces, and only the per-piece row's break of 24 is reached.
func TestBreakIsComparedInTheCandidatesOwnUnit(t *testing.T) {
	perCase := row("R_CASE", "250", "10")
	perPiece := row("R_PC", "11", "24")
	perPiece.PurchaseUomId = uomPc

	req := Request{
		ProductTemplateId: tpl,
		ProductVariantId:  varOne,
		QuantityByUom:     map[string]decimal.Decimal{uomCase: dec("2"), uomPc: dec("48")},
	}

	got, found := Resolve(req, []Candidate{perCase, perPiece})

	if !found || got.VendorProductPriceId != "R_PC" {
		t.Fatalf("only the per-piece break is reached; got %q", got.VendorProductPriceId)
	}
	if got.PurchaseUomId != uomPc {
		t.Fatalf("the price must carry the unit it is per; got %q", got.PurchaseUomId)
	}
}

// A unit the caller could not convert into is skipped, NOT treated as a quantity of zero — which
// would make every break in that unit look reachable.
func TestUnconvertibleUnitIsSkippedNotZeroed(t *testing.T) {
	perKilo := row("R_KG", "90", "0")
	perKilo.PurchaseUomId = "UOM_KG"

	if _, found := Resolve(request("120"), []Candidate{perKilo}); found {
		t.Fatal("a unit with no conversion must not price, even at a zero break")
	}
}

// Section 28 and TS-PRICE-10: no candidates means no price. The caller must not substitute a cost.
func TestNoCandidatesFindsNothing(t *testing.T) {
	if _, found := Resolve(request("120"), nil); found {
		t.Fatal("with nothing to choose from there is no price")
	}
}

func TestLowestSequenceBreaksATie(t *testing.T) {
	late := row("R_AAA", "250", "1")
	late.Sequence = 20
	early := row("R_ZZZ", "240", "1")
	early.Sequence = 5

	got, _ := Resolve(request("120"), []Candidate{late, early})

	if got.VendorProductPriceId != "R_ZZZ" {
		t.Fatalf("the lowest sequence wins even with a later id; got %q", got.VendorProductPriceId)
	}
}

// PRICE-INV-020: two rows alike in every ranked respect must resolve identically however the
// database happened to order them.
func TestResolutionDoesNotDependOnInputOrder(t *testing.T) {
	first := row("R_AAA", "250", "1")
	second := row("R_BBB", "240", "1")

	forward, _ := Resolve(request("120"), []Candidate{first, second})
	reversed, _ := Resolve(request("120"), []Candidate{second, first})

	if forward.VendorProductPriceId != reversed.VendorProductPriceId {
		t.Fatalf("order-dependent resolution: %q vs %q",
			forward.VendorProductPriceId, reversed.VendorProductPriceId)
	}
	if forward.VendorProductPriceId != "R_AAA" {
		t.Fatalf("the lowest id wins the final tie; got %q", forward.VendorProductPriceId)
	}
}

// The result reports the REQUEST's variant. A template-wide row prices a specific variant, and
// echoing the row's empty variant back would lose which product was actually priced.
func TestResolutionReportsTheRequestedVariant(t *testing.T) {
	got, found := Resolve(request("120"), []Candidate{row("R1", "250", "1")})

	if !found || got.ProductVariantId != varOne {
		t.Fatalf("the resolution must name the variant that was priced; got %q", got.ProductVariantId)
	}
}

// Currency and lead time travel with the price: neither is implied, and the lead time is half the
// content of a quote.
func TestResolutionCarriesCurrencyAndLeadTime(t *testing.T) {
	candidate := row("R1", "3.20", "1")
	candidate.CurrencyId = "USD"
	candidate.LeadTimeDays = 30

	got, _ := Resolve(request("120"), []Candidate{candidate})

	if got.CurrencyId != "USD" || got.LeadTimeDays != 30 {
		t.Fatalf("currency and lead time must survive resolution; got %q / %d",
			got.CurrencyId, got.LeadTimeDays)
	}
}
