package vendorpricing

import (
	"testing"
)

// The named purchase-side pricing scenarios, asserted by identifier. resolve_test.go covers the
// same ground from the rules; the overlap is deliberate so an auditor can find a scenario by name.

// Breaks at 1/10/100 priced 250/240/220, a quantity of 120, and the answer is 220: the highest
// break reached wins, not the cheapest row.
func TestScenarioVendorQuantityBreak(t *testing.T) {
	candidates := []Candidate{
		row("R1", "250", "1"),
		row("R10", "240", "10"),
		row("R100", "220", "100"),
	}

	got, found := Resolve(request("120"), candidates)

	if !found || !got.UnitPrice.Equal(dec("220")) {
		t.Fatalf("120 units take the 100+ break; got %s (found=%v)", got.UnitPrice, found)
	}
}

// The same ladder at each rung, so a resolver that merely happened to pick 220 for 120 is caught.
func TestScenarioEachBreakResolvesAtItsOwnRung(t *testing.T) {
	candidates := []Candidate{
		row("R1", "250", "1"),
		row("R10", "240", "10"),
		row("R100", "220", "100"),
	}

	testCases := []struct{ quantity, want string }{
		{"1", "250"},
		{"9", "250"},
		{"10", "240"},
		{"99", "240"},
		{"100", "220"},
		{"1000", "220"},
	}

	for _, testCase := range testCases {
		got, found := Resolve(request(testCase.quantity), candidates)
		if !found || !got.UnitPrice.Equal(dec(testCase.want)) {
			t.Errorf("quantity %s resolved to %s (found=%v), want %s",
				testCase.quantity, got.UnitPrice, found, testCase.want)
		}
	}
}

// A negotiated price does not change the vendor's quote. This package cannot write anywhere, so
// what is asserted is that resolving an unchanged candidate set twice gives the same answer.
func TestScenarioNegotiatingDoesNotChangeTheQuote(t *testing.T) {
	candidates := []Candidate{row("R1", "9500", "1")}

	first, foundFirst := Resolve(request("1"), candidates)
	// A negotiation happens on the order line — 9,200 agreed — then the same product is priced
	// again. The vendor still says 9,500.
	second, foundSecond := Resolve(request("1"), candidates)

	if !foundFirst || !foundSecond {
		t.Fatal("both resolutions must find the quote")
	}
	if !first.UnitPrice.Equal(dec("9500")) || !second.UnitPrice.Equal(dec("9500")) {
		t.Fatalf("the quote must stay 9500; got %s then %s", first.UnitPrice, second.UnitPrice)
	}
	if !candidates[0].UnitPrice.Equal(dec("9500")) {
		t.Fatalf("resolution mutated the candidate to %s", candidates[0].UnitPrice)
	}
}

// No vendor price means no price — never the product's cost, never another vendor's. A fallback to
// cost would put a number on a purchase order that no supplier ever quoted.
func TestScenarioNoVendorPriceDoesNotFallBackToCost(t *testing.T) {
	// Nothing applies: the only row is for a different variant of the same template.
	other := row("R_OTHER", "100", "1")
	other.ProductVariantId = varTwo

	got, found := Resolve(request("1"), []Candidate{other})

	if found {
		t.Fatalf("no quote applies, so nothing may be returned; got %s", got.UnitPrice)
	}
	if !got.UnitPrice.IsZero() || got.VendorProductPriceId != "" {
		t.Fatal("the zero result must be empty rather than a usable-looking price")
	}
}

// This package never sees another organization's rows; the caller filters by org first. What it
// enforces one level down is that a row for a different product is not a candidate, so a row
// escaping the org filter would still have to name this exact product to price it.
func TestScenarioARowForAnotherProductIsNotACandidate(t *testing.T) {
	elsewhere := row("R_ELSEWHERE", "1", "1")
	elsewhere.ProductTemplateId = "TPL_ELSEWHERE"

	if _, found := Resolve(request("1"), []Candidate{elsewhere}); found {
		t.Fatal("a quote for another product must not price this one")
	}
}

// An archived quote must not price anything new. Archived rows are excluded by the caller's read,
// so with nothing else applicable the answer is "none" rather than a stale price. The row stays
// readable for orders that already resolved through it.
func TestScenarioAnArchivedQuoteResolvesToNothing(t *testing.T) {
	if _, found := Resolve(request("1"), nil); found {
		t.Fatal("with the archived row filtered out there is nothing to price with")
	}
}

// A live row still prices, so retiring one quote does not retire the product.
func TestScenarioALiveQuoteStillPricesAfterAnotherIsArchived(t *testing.T) {
	got, found := Resolve(request("1"), []Candidate{row("R_LIVE", "250", "1")})

	if !found || !got.UnitPrice.Equal(dec("250")) {
		t.Fatalf("the surviving quote must price; got %s (found=%v)", got.UnitPrice, found)
	}
}
