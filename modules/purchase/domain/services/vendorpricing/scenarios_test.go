package vendorpricing

import (
	"testing"
)

// The change request's purchase-side named scenarios (section 43), asserted by identifier.
//
// resolve_test.go covers the same ground from the rules; this file exists so that somebody auditing
// the requirement can find TS-PRICE-06 by name rather than working out which behavioural test
// happens to imply it. The overlap is deliberate: a comment pointing at another test does not fail
// when that test is renamed.

// TS-PRICE-06: breaks at 1/10/100 priced 250/240/220, a quantity of 120, and the answer is 220.
//
// The highest break REACHED wins, which is the opposite of the cheapest row winning. A resolver
// that simply took the lowest price would give the 100+ rate to a buyer ordering one.
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

// The same ladder at each rung, so that a resolver which merely happened to pick 220 for 120 is not
// mistaken for one that resolves correctly.
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

// TS-PRICE-07: a negotiated price does not change the vendor's quote.
//
// This package cannot write anywhere, which is most of the guarantee. What it can be asked is
// whether resolving twice returns the same answer — an override on an order line must leave the
// master exactly as it was, so the second resolution of an unchanged candidate set must be
// identical to the first.
func TestScenarioNegotiatingDoesNotChangeTheQuote(t *testing.T) {
	candidates := []Candidate{row("R1", "9500", "1")}

	first, foundFirst := Resolve(request("1"), candidates)
	// A negotiation happens on the order line — 9,200 agreed — and then the same product is priced
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

// TS-PRICE-10: no vendor price means NO price. Never the product's cost, never another vendor's.
//
// The most important assertion in this file. A resolver that fell back to a cost would put a number
// on a purchase order that no supplier ever quoted, and it would look entirely ordinary to whoever
// approved it.
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

// TS-PRICE-11: this package never sees another organization's rows — the caller filters by org
// before reading. What it does enforce is the same shape of protection one level down: a row for a
// different PRODUCT is not a candidate, so a row that escaped the org filter would still have to
// name this exact product to price it.
func TestScenarioARowForAnotherProductIsNotACandidate(t *testing.T) {
	elsewhere := row("R_ELSEWHERE", "1", "1")
	elsewhere.ProductTemplateId = "TPL_ELSEWHERE"

	if _, found := Resolve(request("1"), []Candidate{elsewhere}); found {
		t.Fatal("a quote for another product must not price this one")
	}
}

// TS-PRICE-12: an archived quote must not price anything new.
//
// Archived rows are excluded by the caller's read, so the assertion here is the consequence: with
// the archived row absent and nothing else applicable, the answer is "none" rather than a stale
// price. The row stays readable for the orders that already resolved through it (PRICE-INV-024) —
// readable and usable are different things.
func TestScenarioAnArchivedQuoteResolvesToNothing(t *testing.T) {
	if _, found := Resolve(request("1"), nil); found {
		t.Fatal("with the archived row filtered out there is nothing to price with")
	}
}

// And the case that makes archiving safe rather than destructive: a live row still prices, so
// retiring one quote does not retire the product.
func TestScenarioALiveQuoteStillPricesAfterAnotherIsArchived(t *testing.T) {
	got, found := Resolve(request("1"), []Candidate{row("R_LIVE", "250", "1")})

	if !found || !got.UnitPrice.Equal(dec("250")) {
		t.Fatalf("the surviving quote must price; got %s (found=%v)", got.UnitPrice, found)
	}
}
