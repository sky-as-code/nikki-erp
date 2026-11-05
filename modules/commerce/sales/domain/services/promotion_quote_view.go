package services

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing/promotion"
)

// The PromotionQuote every draft promotion action answers (plan §3.2). Its field names are a
// contract with the frontend (salesOrder/promotionQuote.ts); amounts are decimal strings at the
// currency scale.

// engineReasons maps the engine's internal reasons onto the codes the quote publishes.
var engineReasons = map[string]string{
	promotion.ReasonNoEligibleLine:       PromotionReasonConditionNotMet,
	promotion.ReasonNoRule:               PromotionReasonConditionNotMet,
	promotion.ReasonConditionNotMet:      PromotionReasonConditionNotMet,
	promotion.ReasonCurrencyMismatch:     PromotionReasonScopeMismatch,
	promotion.ReasonVoucherRequired:      PromotionReasonNotAvailable,
	promotion.ReasonVoucherSourceOrder:   PromotionReasonNotAvailable,
	promotion.ReasonNoReward:             PromotionReasonNotAvailable,
	promotion.ReasonUnknownProgramType:   PromotionReasonNotAvailable,
	promotion.ReasonPricelistMismatch:    promotion.ReasonPricelistMismatch,
	promotion.ReasonNoShippingLine:       promotion.ReasonNoShippingLine,
	promotion.ReasonOrderDiscountNotBest: promotion.ReasonOrderDiscountNotBest,
	promotion.ReasonZeroBenefit:          promotion.ReasonZeroBenefit,
	promotion.ReasonCustomerRequired:     promotion.ReasonCustomerRequired,
}

func publishedReason(reason string) string {
	if reason == "" {
		return ""
	}
	if mapped, found := engineReasons[reason]; found {
		return mapped
	}
	return reason
}

// View renders the quote.
func (this *PromotionQuoteResult) View() map[string]any {
	programs := []map[string]any{}
	versions := map[string]int32{}
	for _, program := range this.outcome.Programs {
		candidate := this.pc.byPromotionId(program.PromotionId)
		programs = append(programs, this.programView(candidate, program))
		versions[program.PromotionId] = program.ConfigVersion
	}
	for _, candidate := range this.pc.rejected {
		programs = append(programs, this.rejectedView(candidate))
		versions[candidate.program.Id] = candidate.program.ConfigVersion
	}
	return map[string]any{
		"calculation_reference":     this.Reference,
		"calculated_at":             this.CalculatedAt.Format(time.RFC3339),
		"order_revision":            this.StageVersion,
		"promotion_config_versions": versions,
		"programs":                  programs,
		"totals": map[string]any{
			"subtotal":       this.money(this.totals.Subtotal),
			"discount_total": this.money(this.totals.DiscountTotal),
			"tax_total":      this.money(this.totals.TaxTotal),
			"grand_total":    this.money(this.totals.GrandTotal),
		},
		"warnings": this.warnings(),
	}
}

func (this *PromotionQuoteResult) warnings() []PromotionWarning {
	if this.pc.warnings == nil {
		return []PromotionWarning{}
	}
	return this.pc.warnings
}

func (this *PromotionQuoteResult) money(amount decimal.Decimal) string {
	return amount.StringFixed(this.scale)
}

func (this *PromotionQuoteResult) programView(
	candidate *promotionCandidate, program promotion.ProgramResult,
) map[string]any {
	view := this.programHeader(candidate)
	view["outcome"] = program.Outcome
	if reason := publishedReason(program.ReasonCode); reason != "" {
		view["reason_code"] = reason
	}
	view["rewards"] = rewardViews(candidate, program.RewardId)
	view["discount_untaxed_amount"] = this.money(program.DiscountUntaxedAmount)
	view["tax_reduction_amount"] = this.money(program.TaxReductionAmount)
	view["discount_total_amount"] = this.money(program.DiscountTotalAmount())
	affected := []map[string]any{}
	for _, allocation := range program.Allocations {
		affected = append(affected, map[string]any{
			"order_line_id": allocation.LineKey, "discount_untaxed_amount": this.money(allocation.DiscountUntaxedAmount),
		})
	}
	view["affected_lines"] = affected
	if program.Gift != nil && program.Outcome == promotion.OutcomeApplied {
		view["gift"] = map[string]any{
			"product_variant_id": program.Gift.ProductVariantId, "uom_id": program.Gift.UomId,
			"quantity": program.Gift.Quantity.String(), "order_line_id": this.giftLineIds[program.PromotionId],
		}
	}
	return view
}

func (this *PromotionQuoteResult) rejectedView(candidate *promotionCandidate) map[string]any {
	view := this.programHeader(candidate)
	view["outcome"] = promotion.OutcomeNotEligible
	view["reason_code"] = candidate.rejectReason
	view["rewards"] = rewardViews(candidate, "")
	zero := this.money(decimal.Zero)
	view["discount_untaxed_amount"], view["tax_reduction_amount"], view["discount_total_amount"] = zero, zero, zero
	view["affected_lines"] = []map[string]any{}
	return view
}

func (this *PromotionQuoteResult) programHeader(candidate *promotionCandidate) map[string]any {
	view := map[string]any{}
	if candidate == nil {
		return view
	}
	view["promotion_id"] = candidate.program.Id
	view["reference"] = stringOf(candidate.record, models.SalesPromotionFieldReference)
	view["name"] = stringOf(candidate.record, models.SalesPromotionFieldName)
	view["program_type"] = candidate.program.ProgramType
	if candidate.voucherCode != "" {
		view["voucher_code"] = candidate.voucherCode
	}
	return view
}

// rewardViews lists every alternative of the promotion so a choice can be offered (BR-R12).
func rewardViews(candidate *promotionCandidate, selectedId string) []map[string]any {
	views := []map[string]any{}
	if candidate == nil {
		return views
	}
	for _, reward := range candidate.rewards {
		id := stringOf(reward, basemodel.FieldId)
		views = append(views, map[string]any{
			"reward_id":         id,
			"reward_type":       stringOf(reward, models.SalesPromotionRewardFieldRewardType),
			"order_description": stringOf(reward, models.SalesPromotionRewardFieldOrderDescription),
			"selected":          id != "" && id == selectedId,
		})
	}
	return views
}
