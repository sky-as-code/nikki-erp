package promotion

import (
	"github.com/shopspring/decimal"
)

// Outcomes and reason codes, matching the PromotionQuote contract (plan §3.2).
const (
	OutcomeApplied           = "applied"
	OutcomeSelectionRequired = "selection_required"
	OutcomeDeclined          = "declined"
	OutcomeNotEligible       = "not_eligible"
	OutcomeNotApplicable     = "not_applicable"

	ReasonUnknownProgramType   = "unknown_program_type"
	ReasonCurrencyMismatch     = "currency_mismatch"
	ReasonVoucherRequired      = "voucher_required"
	ReasonVoucherSourceOrder   = "voucher_source_order"
	ReasonNoRule               = "no_rule"
	ReasonNoEligibleLine       = "no_eligible_line"
	ReasonPricelistMismatch    = "pricelist_mismatch"
	ReasonConditionNotMet      = "condition_not_met"
	ReasonNoReward             = "no_reward"
	ReasonNoShippingLine       = "no_shipping_line"
	ReasonZeroBenefit          = "zero_benefit"
	ReasonOrderDiscountNotBest = "order_discount_not_best"
	ReasonCustomerRequired     = "customer_required"

	AllocationMerchandiseDiscount = "merchandise_discount"
	AllocationShippingDiscount    = "shipping_discount"
	AllocationFreeProduct         = "free_product"
)

// Result is what Apply returns.
type Result struct {
	// Programs holds one entry per input program, in application order (Sequence, then Id).
	Programs []ProgramResult

	// Lines is every input purchase line with its promotion discount, in line order.
	Lines []LineOutcome

	// Gifts is the full set of gift lines the order must carry after this calculation.
	Gifts []GiftIntent

	// RemovedGiftLineKeys are existing gift lines no applied program wants any more (BR-R09).
	RemovedGiftLineKeys []string
}

// ProgramResult is one program's outcome.
type ProgramResult struct {
	PromotionId   string
	ProgramType   string
	Sequence      int32
	ConfigVersion int32
	VoucherId     string

	Outcome    string
	ReasonCode string

	Qualification Qualification

	// Rewards lists every alternative so the caller can offer a choice (BR-R12).
	Rewards []RewardOption

	// RewardId/RewardType are the reward actually evaluated (selected or automatic).
	RewardId   string
	RewardType string

	// CalculationSequence is the 1-based application order among applied programs.
	CalculationSequence int32

	DiscountUntaxedAmount decimal.Decimal

	// TaxReductionAmount is zero until AttributeTax distributes Accounting's tax delta.
	TaxReductionAmount decimal.Decimal

	Allocations []Allocation
	Gift        *GiftIntent
}

// DiscountTotalAmount is untaxed discount plus tax reduction.
func (this ProgramResult) DiscountTotalAmount() decimal.Decimal {
	return this.DiscountUntaxedAmount.Add(this.TaxReductionAmount)
}

// Qualification is what Qualify decided, kept for the application's qualification_snapshot.
type Qualification struct {
	Qualified  bool
	ReasonCode string

	// RuleId is the deciding rule: the first qualifying rule by id, or the max-sets rule for BXGY.
	RuleId string

	EligibleLineKeys []string
	EligibleQuantity decimal.Decimal
	EligibleAmount   decimal.Decimal

	// Sets is the BXGY set count, max over qualifying rules (BR-R08); zero for other types.
	Sets int64
}

// RewardOption is one reward offered to the caller.
type RewardOption struct {
	RewardId   string
	RewardType string
	Selected   bool

	// GiftQuantity previews what a free_product reward would give, so the choice is informed.
	GiftQuantity decimal.Decimal
}

// Allocation is one line's share of a program's benefit (BR §8.9).
type Allocation struct {
	LineKey        string
	AllocationType string

	EligibleQuantity      decimal.Decimal
	BaseUntaxedAmount     decimal.Decimal
	DiscountUntaxedAmount decimal.Decimal
	TaxReductionAmount    decimal.Decimal

	CalculationSequence int32
}

// GiftIntent is a gift line the caller must upsert. ExistingLineKey names the current gift line of
// the same promotion and reward when there is one; empty means create.
type GiftIntent struct {
	PromotionId      string
	RewardId         string
	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal
	ExistingLineKey  string
}

// LineOutcome is one purchase line after every promotion.
type LineOutcome struct {
	Key                    string
	BaseUntaxedAmount      decimal.Decimal
	DiscountUntaxedAmount  decimal.Decimal
	RemainingUntaxedAmount decimal.Decimal
}
