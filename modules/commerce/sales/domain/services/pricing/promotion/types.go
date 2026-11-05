package promotion

import (
	"github.com/shopspring/decimal"
)

// The vocabulary below is byte-identical to the enums in the sales promotion JSON models, kept as
// strings so this package imports no schema.
const (
	ProgramAutomatic       = "automatic"
	ProgramDiscountCode    = "discount_code"
	ProgramCoupon          = "coupon"
	ProgramBuyXGetY        = "buy_x_get_y"
	ProgramNextOrderCoupon = "next_order_coupon"

	PhaseApply = "apply"
	PhaseIssue = "issue"

	TaxBasisIncluded = "tax_included"
	TaxBasisExcluded = "tax_excluded"

	RewardDiscount     = "discount"
	RewardFreeProduct  = "free_product"
	RewardFreeShipping = "free_shipping"

	MethodPercentage  = "percentage"
	MethodFixedAmount = "fixed_amount"

	ScopeOrder            = "order"
	ScopeSpecificProducts = "specific_products"
	ScopeCheapestProduct  = "cheapest_product"

	LineTypeProduct       = "product"
	LineTypeCombo         = "combo"
	LineTypePromotionGift = "promotion_gift"
)

// Input is everything Apply needs. The caller (reprice/confirm) loads and filters candidates; this
// package performs no I/O, so the same Input always yields the same Result.
type Input struct {
	OrderId      string
	StageVersion int32
	CurrencyId   string

	// CurrencyScale is the currency's decimal places (0 for VND). Discounts are rounded to it.
	CurrencyScale int32

	// CustomerId is the order's sold-to party; only the issue phase reads it (BR-V10).
	CustomerId string

	Lines []Line

	// Programs are the candidates that already passed BR-C01/BR-S01 filtering (status, window,
	// channel AND point, usage). Order is irrelevant: Apply sorts by (Sequence, Id).
	Programs []Program

	// Selections are the persisted draft reward choices (D6), at most one per promotion.
	Selections []Selection
}

// Line is one order line as it stands BEFORE any promotion: after pricelist and manual discounts.
type Line struct {
	// Key is echoed back on every allocation and gift intent (the order line id when it exists).
	Key        string
	LineNumber int32

	// LineType is LineTypeProduct, LineTypeCombo or LineTypePromotionGift. Only product and combo
	// lines with a positive quantity are purchase lines (BR-C05); gift lines never qualify nor
	// receive a discount (BR-R02, BR-R08).
	LineType string

	ProductVariantId  string
	ProductTemplateId string

	// CategoryIds holds the variant's own category AND every ancestor, so a category selector
	// matches its descendants without a lookup (BR-C03).
	CategoryIds []string

	// PricelistId is the pricelist whose item priced this line (D2); empty when none did.
	PricelistId string

	UomId    string
	Quantity decimal.Decimal

	// BaseQuantity is Quantity in the product's base UoM, summed by non-BXGY rules (BR-C07).
	BaseQuantity decimal.Decimal

	// QualificationQuantities is Quantity expressed in each BXGY rule's qualification UoM, keyed by
	// that UoM id. The caller fills it only where the UoMs share a category; a missing key means the
	// line cannot count toward that rule (BR-C07).
	QualificationQuantities map[string]decimal.Decimal

	// UntaxedAmount is N_i (BR-C06); TaxAmount is T_i computed on N_i by Accounting.
	UntaxedAmount decimal.Decimal
	TaxAmount     decimal.Decimal

	// SourcePromotionId/SourceRewardId identify an existing gift line, so a reprice updates it
	// instead of adding a duplicate (BR-R09).
	SourcePromotionId string
	SourceRewardId    string
}

// Program is one candidate promotion with its rules and rewards.
type Program struct {
	Id            string
	ProgramType   string
	Sequence      int32
	ConfigVersion int32
	CurrencyId    string

	// PricelistIds empty means no pricelist restriction; otherwise only lines priced by one of them
	// are eligible (D2).
	PricelistIds []string

	// VoucherId is the attached, still-available voucher that gates discount_code, coupon and
	// next_order_coupon. VoucherSourceOrderId is that voucher's source order: a coupon is never
	// applied on its own source order (BR-V09).
	VoucherId            string
	VoucherSourceOrderId string

	Rules   []Rule
	Rewards []Reward
}

// Rule is one OR-alternative for qualifying (BR-C04).
type Rule struct {
	Id    string
	Phase string

	MinQuantity decimal.Decimal
	MinAmount   decimal.Decimal

	// AmountTaxBasis is TaxBasisIncluded or TaxBasisExcluded; empty reads as excluded.
	AmountTaxBasis string

	// QualificationUomId is set for buy_x_get_y only; the rule then counts QualificationQuantities.
	QualificationUomId string

	Selector Selector
}

// Selector lists structured targets. OR inside one list, AND across the two dimensions; an empty
// dimension does not restrict. Tags are not supported (Inventory has no product tag).
type Selector struct {
	ProductTemplateIds []string
	ProductVariantIds  []string
	CategoryIds        []string
}

// Reward is one alternative benefit (BR-R12).
type Reward struct {
	Id         string
	RewardType string

	DiscountMethod string
	DiscountValue  decimal.Decimal
	DiscountScope  string

	// MaxDiscountAmount is an untaxed cap; zero means uncapped (BR-R04).
	MaxDiscountAmount decimal.Decimal

	// Selector restricts specific_products and cheapest_product targets only.
	Selector Selector

	GiftProductVariantId string
	GiftQuantity         decimal.Decimal
	GiftUomId            string
}

// Selection is the caller's persisted choice for one promotion.
type Selection struct {
	PromotionId string
	RewardId    string
	Declined    bool
}
