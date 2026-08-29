// Package vendorpricing decides which vendor price applies to a purchase. It is pure: no
// repository, no clock, no context. The caller loads every candidate and normalises the quantity;
// this package only chooses. It never falls back — when no vendor price applies the answer is
// "none", never the product's cost and never another vendor's price, which would put a number on a
// purchase order that no supplier quoted.
package vendorpricing

import (
	"github.com/shopspring/decimal"
)

// Candidate is one vendor price row the caller has already read; the caller filters by vendor and
// organization first, and this package decides which of the remaining rows wins.
type Candidate struct {
	// Echoed onto the result so a line records which quote it resolved through, and so two
	// otherwise identical rows resolve the same way on every run.
	Id string

	ProductTemplateId string

	// Empty for a price covering the whole template. Emptiness is the specificity mechanism: a
	// variant-specific price beats a template-wide one from the same vendor.
	ProductVariantId string

	PurchaseUomId string
	CurrencyId    string

	// MinQuantity is the break this price applies from, inclusive, expressed in PurchaseUomId.
	MinQuantity decimal.Decimal
	UnitPrice   decimal.Decimal

	// Whether this row's commercial window covers the pricing date. A boolean rather than the dates
	// themselves, because comparing them needs a clock and this package has none; the caller
	// evaluates the window against its own pricing date.
	Applicable bool

	LeadTimeDays int32

	// Breaks ties between equally specific, equally applicable rows; lowest wins.
	Sequence int32
}

// Request is what the caller wants priced.
type Request struct {
	ProductTemplateId string
	ProductVariantId  string

	// The requested quantity expressed in each candidate's unit, keyed by uom id. A map because
	// this package does no unit arithmetic: candidates quote in different units and a quantity
	// break is compared in the break's own unit, so the caller converts once per distinct unit. A
	// unit absent from the map means the conversion failed, and candidates in it are skipped rather
	// than compared against a number meaning something else.
	QuantityByUom map[string]decimal.Decimal
}

// Resolution is the winning candidate. The unit and currency travel with the price because neither
// is implied: a line records what was quoted, in the unit and currency it was quoted in. There is
// no FX service here, so reconciling a foreign currency is the caller's problem.
type Resolution struct {
	VendorProductPriceId string
	ProductVariantId     string
	UnitPrice            decimal.Decimal
	PurchaseUomId        string
	CurrencyId           string
	LeadTimeDays         int32
}
