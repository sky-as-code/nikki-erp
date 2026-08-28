// Package vendorpricing decides which vendor price applies to a purchase (section 27).
//
// It is a PURE function, exactly like sales/domain/services/pricing: no repository, no clock, no
// context. The caller loads every candidate and normalises the quantity, then hands both in; this
// package only chooses and reports. That is what makes the resolution rules testable without a
// database, and it is the reason the sales engine was built the same way.
//
// The one thing it deliberately does NOT do is fall back. When no vendor price applies the answer
// is "none", never the product's cost and never a price from some other vendor (section 28,
// AC-PRICE-036, TS-PRICE-10). A cost is what the business has paid and valued; a vendor price is
// what somebody is offering. Substituting one for the other would put a number on a purchase order
// that no supplier ever quoted.
package vendorpricing

import (
	"github.com/shopspring/decimal"
)

// Candidate is one vendor price row the caller has already read.
//
// The caller filters by vendor and organization before calling — those are indexed reads, and
// doing them here would mean passing the whole table in. What this package decides is which of the
// remaining rows wins.
type Candidate struct {
	// Id is echoed onto the result, so a purchase order line can record which quote it resolved
	// through, and so that two otherwise identical rows still resolve the same way on every run.
	Id string

	ProductTemplateId string

	// ProductVariantId is empty for a price covering the whole template. Emptiness IS the
	// specificity mechanism: a variant-specific price beats a template-wide one from the same
	// vendor (PRICE-INV-018).
	ProductVariantId string

	PurchaseUomId string
	CurrencyId    string

	// MinQuantity is the break this price applies from, inclusive, expressed in PurchaseUomId.
	MinQuantity decimal.Decimal
	UnitPrice   decimal.Decimal

	// Applicable says whether this row's commercial window covers the pricing date.
	//
	// A boolean rather than the dates themselves, because comparing them needs a clock and this
	// package has none — a pure function that read the time would give a different answer on two
	// runs with the same input, which is the property the whole design exists to preserve. The
	// caller evaluates the window against its own pricing date and reports the verdict.
	Applicable bool

	LeadTimeDays int32

	// Sequence breaks ties between rows that are equally specific and equally applicable, LOWEST
	// winning (section 27 step 6).
	Sequence int32
}

// Request is what the caller wants priced.
type Request struct {
	ProductTemplateId string
	ProductVariantId  string

	// QuantityByUom is the requested quantity expressed in EACH candidate's unit, keyed by uom id.
	//
	// A map rather than a single quantity because the conversion is Essential's to do and this
	// package must not do arithmetic on units (BR-PRICE-UOM-002, PRICE-INV-025). Candidates quote
	// in different units — one per case, another per piece — and a quantity break is compared
	// against the break's OWN unit (BR-PRICE-UOM-004), so the caller converts once per distinct
	// unit and passes the answers in.
	//
	// A unit absent from the map means the conversion failed or was not attempted; candidates in
	// that unit are skipped rather than compared against a number that means something else.
	QuantityByUom map[string]decimal.Decimal
}

// Resolution is the winning candidate, reported in full.
//
// The unit and currency travel with the price because neither is implied. A purchase order line
// records what was quoted, in the unit and currency it was quoted in — reconciling a foreign
// currency is the caller's problem, and there is no FX service that could do it here.
type Resolution struct {
	VendorProductPriceId string
	ProductVariantId     string
	UnitPrice            decimal.Decimal
	PurchaseUomId        string
	CurrencyId           string
	LeadTimeDays         int32
}
