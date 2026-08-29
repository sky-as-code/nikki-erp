// Package pricing is the sales pricing engine: a pure function from a basket to a priced result.
// It performs no I/O — the caller loads every price, combo, promotion and voucher and passes them in
// — so the same input always gives the same output and quote and confirm cannot drift apart. Its
// nine steps run in a fixed order because discounts do not commute, and every step that changes a
// number emits an Adjustment so the result can be replayed.
package pricing

import (
	"github.com/shopspring/decimal"
)

// InternalScale is the number of decimal places carried through the calculation. It is finer than
// any currency's own scale on purpose: rounding to the currency happens once, at step 8, because
// rounding at every allocation accumulates error.
const InternalScale int32 = 4

// LineInput is one thing the customer wants to buy.
type LineInput struct {
	// Key identifies this line to the caller; passed through untouched so results match back without
	// depending on slice order.
	Key string

	// LineNumber orders the lines and breaks allocation ties.
	LineNumber int32

	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal

	// CatalogueUnitPrice is the fallback price when no pricelist item matches. The caller resolves it
	// from Inventory; the engine never looks one up.
	CatalogueUnitPrice decimal.Decimal

	// ProductTemplateId and CategoryPath place the variant in the product hierarchy so template- and
	// category-targeted rules match without a lookup. CategoryPath runs from the variant's OWN
	// category outward to the root, and that order IS the precedence: the nearest ancestor wins. The
	// caller builds it by walking parent_category_id, keeping the I/O outside this package.
	ProductTemplateId string
	CategoryPath      []string

	// UnitCost is the variant's cost, for a FORMULA rule based on COST. Sales only reads it; the
	// number belongs to Inventory. HasCost distinguishes a real zero cost (a giveaway) from an absent
	// one, which must refuse rather than price at zero.
	UnitCost decimal.Decimal
	HasCost  bool

	// ProductCode and ProductName are snapshotted onto the result, sparing the caller a lookup.
	ProductCode string
	ProductName string

	// ComboId, when set, marks this line a combo parent, expanded from the caller's ComboDefinition.
	ComboId string

	// RequiresFulfillment is false for a service or a fee. Carried through untouched; the caller
	// needs it to derive fulfilment status.
	RequiresFulfillment bool
}

// PricelistItem is one applicable price the caller has already filtered to this context. Scope
// resolution belongs to the caller; the engine only picks the best matching item per line.
type PricelistItem struct {
	// Id is the rule, echoed onto the result for provenance and as the final deterministic tiebreak
	// between otherwise identical rules.
	Id string

	// AppliesTo names which of the three target fields below is meaningful, or ALL_PRODUCTS. The
	// schema enum, as a string so this package depends on no model.
	AppliesTo string

	ProductVariantId  string
	ProductTemplateId string
	ProductCategoryId string

	UomId     string
	UnitPrice decimal.Decimal

	// MinQuantity is the quantity break this price applies from, inclusive.
	MinQuantity decimal.Decimal

	// Specificity ranks the PRICELIST this item came from: higher wins. The caller computes it from
	// the pricelist's scope. It is distinct from and outranks target specificity — a point-scoped
	// list beats a channel-scoped one whatever either rule targets.
	Specificity int32

	// Priority breaks ties between items of equal specificity, higher winning.
	Priority int32

	// Sequence breaks ties between equally specific, equally applicable rules, LOWEST winning — the
	// opposite direction to Priority above.
	Sequence int32

	// CalculationMethod is FIXED_PRICE, DISCOUNT or FORMULA. Empty means FIXED_PRICE, preserving the
	// behaviour callers had before this field existed.
	CalculationMethod string

	// DiscountPercent is taken off the base for DISCOUNT and FORMULA. Signed: negative marks up,
	// expressing "cost plus 50%" without a second field.
	DiscountPercent decimal.Decimal

	// BasePriceSource is what a FORMULA rule starts from: BASE_SALES_PRICE, COST or OTHER_PRICELIST.
	// For OTHER_PRICELIST the caller resolves the other list and supplies ResolvedBasePrice.
	BasePriceSource   string
	ResolvedBasePrice decimal.Decimal
	HasResolvedBase   bool

	// SurchargeAmount is added after rounding, so an amount meant to be exact stays exact.
	SurchargeAmount decimal.Decimal

	// RoundingIncrement is the step a FORMULA result is rounded to. Zero means no rounding.
	RoundingIncrement decimal.Decimal
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

	// ReferencePrice is what this component would cost standalone: the basis for allocating the combo
	// price across components, never an input to the combo price itself.
	ReferencePrice decimal.Decimal
}

// AppliedProgram is a promotion the caller has already found eligible and resolved for conflicts.
// The engine applies what it is given, in the order it is given.
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

	// Type is one of the models.PromotionReward* values, as a string so this package imports no
	// models.
	Type string

	Value decimal.Decimal

	// TargetScope is "order" or "line".
	TargetScope string

	// TargetLineKeys names the lines a line-scoped reward applies to. Empty means every line.
	TargetLineKeys []string

	// FreeProductVariantId and FreeUom describe the giveaway line a free_quantity reward creates. A
	// free item is a real order line, not an adjustment: Inventory must fulfil it and its VAT
	// treatment is a line-level question.
	FreeProductVariantId string
	FreeUomId            string
	FreeProductCode      string
	FreeProductName      string
}

// Context is what the engine needs to know about the sale, beyond the basket.
type Context struct {
	// CurrencyScale is the currency's own number of decimal places (0 for VND). The final rounding
	// step rounds to this, not to InternalScale.
	CurrencyScale int32

	// TaxByLineKey carries the tax Accounting computed for each line, keyed by LineResult.Key. The
	// engine never calls Accounting itself — that would break replayability; the caller resolves tax
	// first via services.ResolveBasketTax. A line absent from the map is taxed at zero, which is safe
	// because ResolveBasketTax refuses any document it could not resolve.
	TaxByLineKey map[string]LineTax
}

// LineTax is Accounting's answer for one line, converted at this boundary into the units the engine
// and schema use. Accounting stores a rate as a PERCENTAGE (8 means 8%) while
// sales_order_line.tax_rate_snapshot is a FRACTION; mixing them fails silently and hundredfold.
type LineTax struct {
	// RateSnapshot is the effective rate as a FRACTION (0.1 for 10%), the sum of the line's component
	// rates, recorded so a historical line reads without the tax master.
	RateSnapshot decimal.Decimal

	// Amount is the rounded tax Accounting charged on this line.
	Amount decimal.Decimal

	// ComponentAmounts is the per-combo-component split, keyed by ComponentResult.Key. Empty means
	// the engine allocates the line's tax across its components itself.
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

// Adjustment is one step of the calculation, recorded so the price can be explained. The engine
// returns a list rather than a total because discounts do not commute; Sequence makes the replay
// exact.
type Adjustment struct {
	Sequence int32

	// LineKey is empty for an order-level adjustment.
	LineKey string

	Type        string
	SourceType  string
	SourceId    string
	Description string

	// BaseAmount is what the adjustment was calculated FROM; it depends on where in the sequence the
	// adjustment fell.
	BaseAmount decimal.Decimal

	// Amount is SIGNED: negative for a discount, positive for a clawback or rounding increase. The
	// order's discount_total is positive by convention; the sign here says what the step did.
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

	// PricelistItemId names the rule that set this price, when one did, so a resolved price has
	// provenance somebody can go and read.
	PricelistItemId string

	// SourcePromotionProgramId ties a giveaway line back to the campaign that gave it away, for
	// campaign cost review.
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

	// ManualDiscounts are the operator overrides stored against this order. They must be engine
	// INPUT, replayed on every reprice: repricing deletes the whole adjustment chain and rewrites it
	// from engine output, so an adjustment written outside the engine would vanish on the next edit.
	ManualDiscounts []ManualDiscountInput

	Context Context
}

// ManualDiscountInput is one operator override. It NEVER touches the base price: the line keeps its
// catalogue/pricelist price and this rides on top as an adjustment, so the price explanation stays a
// visible chain.
type ManualDiscountInput struct {
	// Id identifies the stored override, so the adjustment can point back at the authorising record.
	Id string

	// LineKey names the line this applies to; empty means the whole order.
	LineKey string

	// Amount is what comes off, as a POSITIVE number; the engine subtracts it. A negative value would
	// be an unauthorised surcharge and the caller rejects it before this point.
	Amount decimal.Decimal

	// Reason is mandatory and travels into the adjustment description.
	Reason string
}

// The rule vocabulary, mirrored as plain strings rather than imported from models so this package
// depends on nothing. The values are byte-identical to the enum in sales_pricelist_item.json.
const (
	AppliesToVariant     = "PRODUCT_VARIANT"
	AppliesToTemplate    = "PRODUCT_TEMPLATE"
	AppliesToCategory    = "PRODUCT_CATEGORY"
	AppliesToAllProducts = "ALL_PRODUCTS"

	MethodFixedPrice = "FIXED_PRICE"
	MethodDiscount   = "DISCOUNT"
	MethodFormula    = "FORMULA"

	BaseSourceBaseSalesPrice = "BASE_SALES_PRICE"
	BaseSourceOtherPricelist = "OTHER_PRICELIST"
	BaseSourceCost           = "COST"
)
