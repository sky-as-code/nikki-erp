// Package pricing is the sales pricing engine: a pure function from a basket to a priced result.
//
// It performs NO I/O (D-13). Every catalogue price, pricelist, combo, promotion and voucher is
// loaded by the caller and passed in. That is what makes BR §13's "same input ⇒ same output, always"
// testable at all, and it is why the same engine serves both a quote preview and a confirm: neither
// can drift from the other, because there is only one calculation.
//
// The nine steps of BR §13 run in a fixed order, and the order is load-bearing — discounts do not
// commute, so a percentage applied before a fixed amount gives a different total than the reverse.
// Every step that changes a number emits an Adjustment recording what it did, so the result can be
// replayed rather than merely trusted.
package pricing

import (
	"github.com/shopspring/decimal"
)

// InternalScale is the number of decimal places carried through the calculation (D-01).
//
// Four rather than the currency's own scale, deliberately: a chain of proportional allocations
// rounded to whole dong at every step accumulates error, so the working precision stays finer than
// the presentation precision and rounding happens once, at step 8.
const InternalScale int32 = 4

// LineInput is one thing the customer wants to buy.
type LineInput struct {
	// Key identifies this line to the caller. Passed through untouched, so the caller can match
	// results back without depending on slice order.
	Key string

	// LineNumber orders the lines and breaks allocation ties (D-04).
	LineNumber int32

	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal

	// CatalogueUnitPrice is the fallback price, used when no pricelist item matches. The caller
	// resolves it from Inventory; the engine never looks one up.
	CatalogueUnitPrice decimal.Decimal

	// ProductCode and ProductName are snapshotted onto the result so the caller can write them to
	// the order line without a second lookup.
	ProductCode string
	ProductName string

	// ComboId, when set, marks this line as a combo parent. Its components are expanded from the
	// ComboDefinition the caller supplied.
	ComboId string

	// RequiresFulfillment is false for a service or a fee. Carried through untouched — the engine
	// does not price differently for it, but the caller needs it to derive fulfilment status.
	RequiresFulfillment bool
}

// PricelistItem is one applicable price the caller has already filtered to this context.
//
// The engine does not decide which pricelists apply — that is scope resolution, which needs the
// channel and point and belongs to the caller. It only picks the best matching item per line.
type PricelistItem struct {
	ProductVariantId string
	UomId            string
	UnitPrice        decimal.Decimal

	// MinQuantity is the quantity break this price applies from, inclusive.
	MinQuantity decimal.Decimal

	// Specificity ranks the pricelist this item came from: higher wins. The caller computes it
	// from the pricelist's scope, so the engine needs no knowledge of channels or points.
	Specificity int32

	// Priority breaks ties between items of equal specificity, higher winning.
	Priority int32
}

// ComboDefinition is a bundle the caller resolved for a combo line.
type ComboDefinition struct {
	ComboId    string
	ComboPrice decimal.Decimal
	Components []ComboComponentInput
}

// ComboComponentInput is one real product inside a bundle.
type ComboComponentInput struct {
	Key              string
	Sequence         int32
	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal
	ProductCode      string
	ProductName      string

	// ReferencePrice is what this component would cost standalone. It is the BASIS for allocating
	// the combo price across components (BR §18) — never an input to the combo price itself, which
	// is independent by BR §15.
	ReferencePrice decimal.Decimal
}

// AppliedProgram is a promotion the caller has already found eligible and resolved for conflicts.
//
// Eligibility and conflict resolution happen outside the engine, in ResolvePromotions and the
// caller's condition evaluation. The engine's job is to apply what it is given, in the order it is
// given, which keeps it a pure function of its inputs.
type AppliedProgram struct {
	ProgramId   string
	ProgramName string

	// VoucherCode is set when this program was activated by a code, for the adjustment description.
	VoucherCode string

	Rewards []RewardInput
}

// RewardInput is one thing a program grants.
type RewardInput struct {
	RewardId string
	Sequence int32

	// Type is one of the models.PromotionReward* values. A string rather than the typed constant so
	// this package does not import the models package, keeping the engine free of schema concerns.
	Type string

	Value decimal.Decimal

	// TargetScope is "order" or "line".
	TargetScope string

	// TargetLineKeys names the lines a line-scoped reward applies to. Empty means every line.
	TargetLineKeys []string

	// FreeProductVariantId and FreeUom describe the giveaway line a free_quantity reward creates
	// (D-11). A free item is a real order line rather than an adjustment, because Inventory must
	// physically fulfil it and its VAT treatment is a line-level question.
	FreeProductVariantId string
	FreeUomId            string
	FreeProductCode      string
	FreeProductName      string
}

// Context is what the engine needs to know about the sale, beyond the basket.
type Context struct {
	// CurrencyScale is the currency's own number of decimal places — 0 for VND. The final rounding
	// step rounds to this, not to InternalScale.
	CurrencyScale int32

	// TaxByLineKey carries the tax Accounting computed for each line, keyed by LineResult.Key.
	//
	// The engine does not call Accounting itself. It is pure (D-13) — no I/O, no clock — and a
	// calculation that reached into another module could not be replayed from its inputs, which is
	// the whole property that makes a historical sale reproducible. The caller resolves tax first
	// and hands the answer in; see services.ResolveBasketTax.
	//
	// A line absent from the map is taxed at zero. That is the correct reading for a giveaway line
	// Accounting was not asked about, and it is safe because a document Accounting could not resolve
	// never reaches this point at all — ResolveBasketTax refuses it.
	TaxByLineKey map[string]LineTax
}

// LineTax is Accounting's answer for one line, in the units the engine and the schema use.
//
// Converted at the boundary rather than carried in Accounting's own units on purpose. Accounting
// stores a rate as a PERCENTAGE (8 means 8%); sales_order_line.tax_rate_snapshot is documented as a
// FRACTION. Keeping the two apart in one struct would leave every reader to remember which is which,
// and the failure is silent and hundredfold.
type LineTax struct {
	// RateSnapshot is the effective rate as a FRACTION: 0.1 for 10%. It is the sum of the line's
	// component rates, recorded so a historical line can be read without the tax master.
	RateSnapshot decimal.Decimal

	// Amount is the rounded tax Accounting charged on this line.
	Amount decimal.Decimal

	// ComponentAmounts is the per-combo-component split, keyed by ComponentResult.Key, when the
	// caller asked Accounting to tax components individually. Empty means the engine allocates the
	// line's tax across its components itself.
	ComponentAmounts map[string]decimal.Decimal
}

// AdjustmentKind mirrors the sales_order_adjustment types, as strings so this package imports no
// schema.
const (
	AdjustmentComboPrice         = "combo_price"
	AdjustmentConditionalPrice   = "conditional_price"
	AdjustmentPercentageDiscount = "percentage_discount"
	AdjustmentFixedDiscount      = "fixed_discount"
	AdjustmentVoucher            = "voucher"
	AdjustmentManualDiscount     = "manual_discount"
	AdjustmentRounding           = "rounding"
)

// Adjustment is one step of the calculation, recorded so the price can be explained.
//
// BR §13 requires the engine to return a LIST of these rather than a total: discounts do not
// commute, so a total alone could never be reproduced. Sequence is what makes the replay exact.
type Adjustment struct {
	Sequence int32

	// LineKey is empty for an order-level adjustment.
	LineKey string

	Type        string
	SourceType  string
	SourceId    string
	Description string

	// BaseAmount is what the adjustment was calculated FROM. Stored because it depends on where in
	// the sequence the adjustment fell: ten percent applied third operates on a different base than
	// the same ten percent applied first.
	BaseAmount decimal.Decimal

	// Amount is SIGNED: negative for a discount, positive for a clawback or a rounding increase.
	// The order's discount_total is positive by convention, but that is a presentation choice; here
	// the sign says what the step did to the number.
	Amount decimal.Decimal
}

// LineResult is one priced line.
type LineResult struct {
	Key        string
	LineNumber int32

	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal
	ProductCode      string
	ProductName      string

	LineType      string
	PricingSource string

	// SourcePromotionProgramId is set on a giveaway line (D-11), tying the free item back to the
	// campaign that gave it away — which is the question asked when a campaign's cost is reviewed.
	SourcePromotionProgramId string
	ComboId                  string

	BaseUnitPrice      decimal.Decimal
	EffectiveUnitPrice decimal.Decimal
	GrossAmount        decimal.Decimal
	DiscountAmount     decimal.Decimal
	NetAmount          decimal.Decimal
	TaxRateSnapshot    decimal.Decimal
	TaxAmount          decimal.Decimal
	FinalAmount        decimal.Decimal

	RequiresFulfillment bool

	Components []ComponentResult
}

// ComponentResult is one combo component with its allocated share.
type ComponentResult struct {
	Key              string
	Sequence         int32
	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal
	ProductCode      string
	ProductName      string

	AllocatedNetAmount decimal.Decimal
	AllocatedTaxAmount decimal.Decimal
}

// Result is what the engine returns.
type Result struct {
	Lines       []LineResult
	Adjustments []Adjustment

	Subtotal      decimal.Decimal
	DiscountTotal decimal.Decimal
	TaxTotal      decimal.Decimal
	GrandTotal    decimal.Decimal
}

// Input is everything Calculate needs.
type Input struct {
	Lines          []LineInput
	PricelistItems []PricelistItem
	Combos         []ComboDefinition
	Programs       []AppliedProgram

	// ManualDiscounts are the operator overrides stored against this order (BR 87.4).
	//
	// They are ENGINE INPUT, replayed on every reprice, rather than adjustments written once
	// afterwards. That is forced by how repricing works: it deletes the whole adjustment chain and
	// rewrites it from engine output, so an adjustment written outside the engine would vanish on
	// the next line edit - and confirm reprices unconditionally, so it would vanish before the sale
	// completed. Feeding them in is what makes them survive.
	ManualDiscounts []ManualDiscountInput

	Context Context
}

// ManualDiscountInput is one operator override (BR 87.4).
//
// It NEVER touches the base price. The line keeps the price the catalogue and the pricelist gave it,
// and this rides on top as an adjustment - so BR 87.9's explanation still reads as a chain from
// catalogue price to final amount, with the override as one visible link rather than as an
// unexplained difference at the start.
type ManualDiscountInput struct {
	// Id identifies the stored override, so the adjustment it produces can point back at it and an
	// operator can see which record authorised the change.
	Id string

	// LineKey names the line this applies to. Empty means the whole order, which is how a
	// goodwill discount on a basket is expressed.
	LineKey string

	// Amount is what comes off, as a POSITIVE number. The engine subtracts it; a negative value
	// here would be a surcharge, which BR 87.4 does not authorise and which the caller rejects
	// before it reaches the engine.
	Amount decimal.Decimal

	// Reason is mandatory (BR 87.4) and travels into the adjustment description, so the price
	// explanation says WHY rather than merely that somebody was allowed to.
	Reason string
}
