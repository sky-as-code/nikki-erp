package services

import (
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing/promotion"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
)

// Running the promotion engine inside a reprice (plan §3.1, §3.3). The order of the passes:
//
//  1. Calculate prices the lines up to manual discounts (pricing.Calculate);
//  2. Accounting taxes those nets tax-inclusively, giving N_i and T_i;
//  3. the pure engine qualifies and computes the benefits (pricing/promotion.Apply);
//  4. Accounting taxes the discounted untaxed values tax-exclusively, and the delta per line is the
//     tax reduction (promotion.AttributeTax);
//  5. the effects become line adjustments and the totals are re-rounded;
//  6. the stateless calculation reference is computed over the result (D4).
//
// An order with no candidate, no attached voucher and no gift line skips 2-5 entirely, and one
// where no promotion yields a money benefit skips 4-5: both price and tax exactly as before
// promotions existed.

// PromotionQuoteResult is what one evaluation concluded, kept for the quote, the draft write and,
// on confirm, the commit.
type PromotionQuoteResult struct {
	Reference    string
	CalculatedAt time.Time
	StageVersion int32

	pc      *promotionContext
	engine  promotion.Input
	outcome promotion.Result
	issues  []promotion.IssueDecision
	facts   map[string]itExt.PromotionProductFacts
	scale   int32
	totals  pricing.Result

	// giftLineIds names the order line each applied gift was written to, by promotion id.
	giftLineIds map[string]string
}

type promotionRun struct {
	order   dmodel.DynamicFields
	input   pricing.Input
	base    pricing.Result
	baseTax *BasketTax
	taxSvc  itExt.TaxCalculationExtService
	policy  SalesPolicy
}

func runPromotions(
	ctx corectx.Context, run promotionRun,
) (pricing.Result, *BasketTax, *PromotionQuoteResult, *ft.ClientErrors, error) {
	now := time.Now().UTC()
	pc, err := loadPromotionContext(ctx, run.order, now)
	if err != nil {
		return pricing.Result{}, nil, nil, nil, err
	}
	quote := &PromotionQuoteResult{
		CalculatedAt: now, StageVersion: int32Of(run.order, models.SalesOrderFieldStageVersion),
		pc: pc, scale: run.policy.RoundingScale, giftLineIds: map[string]string{},
	}
	final, finalTax := run.base, run.baseTax
	if pc.hasWork() {
		var vErrs *ft.ClientErrors
		final, finalTax, vErrs, err = evaluatePromotions(ctx, run, quote)
		if err != nil || vErrs != nil {
			return pricing.Result{}, nil, nil, vErrs, err
		}
	}
	quote.totals = final
	quote.Reference = promotion.CalculationReference(promotion.ReferenceInput{
		OrderId: pc.orderId, StageVersion: quote.StageVersion, CurrencyScale: quote.scale,
		Programs: quote.outcome.Programs, GrandTotal: final.GrandTotal,
	})
	return final, finalTax, quote, nil, nil
}

func evaluatePromotions(
	ctx corectx.Context, run promotionRun, quote *PromotionQuoteResult,
) (pricing.Result, *BasketTax, *ft.ClientErrors, error) {
	pc := quote.pc
	taxBefore, vErrs, err := resolveLineTax(ctx, run, run.base.Lines, itExt.PriceInclusionIncluded)
	if err != nil || vErrs != nil {
		return pricing.Result{}, nil, vErrs, err
	}
	withIssuers := append(append([]*promotionCandidate{}, pc.candidates...), pc.issuers...)
	quote.facts = resolvePromotionFacts(ctx, run.base.Lines, withIssuers)
	params := engineInputParams{
		pc: pc, order: run.order, pricingInput: run.input, base: run.base,
		baseTax: taxBefore.ByLineKey, scale: quote.scale, facts: quote.facts,
	}
	quote.engine = buildEngineInput(ctx, params, pc.candidates)
	quote.outcome = promotion.Apply(quote.engine)
	quote.issues = evaluateIssuePhase(quote.engine, pc)
	// No money benefit (none qualified, a gift only, a zero benefit): the order keeps exactly the
	// legacy price and tax. Accounting's inclusive answer above served only as the engine's T_i.
	if !hasMerchandiseDiscount(quote.outcome) {
		return run.base, run.baseTax, nil, nil
	}
	taxAfter, vErrs, err := resolveLineTax(ctx, run, untaxedAfter(run.base.Lines, quote), itExt.PriceInclusionExcluded)
	if err != nil || vErrs != nil {
		return pricing.Result{}, nil, vErrs, err
	}
	promotion.AttributeTax(&quote.outcome, taxAmounts(taxBefore), taxAmounts(taxAfter), quote.scale)
	effects := pricing.PromotionEffects{Lines: lineEffectsOf(quote), TaxByLineKey: taxAfter.ByLineKey}
	return pricing.ApplyPromotionEffects(run.base, effects, quote.scale), taxAfter, nil, nil
}

func evaluateIssuePhase(engine promotion.Input, pc *promotionContext) []promotion.IssueDecision {
	if len(pc.issuers) == 0 {
		return nil
	}
	issueInput := engine
	issueInput.Programs = nil
	for _, candidate := range pc.issuers {
		issueInput.Programs = append(issueInput.Programs, candidate.program)
	}
	decisions := promotion.EvaluateIssues(issueInput)
	for _, decision := range decisions {
		if decision.ReasonCode == promotion.ReasonCustomerRequired {
			pc.warnings = append(pc.warnings, PromotionWarning{
				Code: PromotionWarningCouponNeedsCustomer, PromotionId: decision.PromotionId,
			})
		}
	}
	return decisions
}

func hasMerchandiseDiscount(outcome promotion.Result) bool {
	for _, program := range outcome.Programs {
		if program.Outcome == promotion.OutcomeApplied && len(program.Allocations) > 0 {
			return true
		}
	}
	return false
}

// resolveLineTax asks Accounting for the tax on the given lines in one document request.
func resolveLineTax(
	ctx corectx.Context, run promotionRun, lines []pricing.LineResult, mode itExt.PriceInclusionMode,
) (*BasketTax, *ft.ClientErrors, error) {
	return ResolveBasketTax(ctx, run.taxSvc, TaxRequestContext{
		OrgId:               stringOf(run.order, basemodel.FieldOrgId),
		TaxDate:             taxDateOf(run.order),
		CurrencyCode:        stringOf(run.order, models.SalesOrderFieldCurrencyCode),
		TaxCode:             run.policy.DefaultSalesTaxCode,
		PriceMode:           mode,
		BusinessChannelCode: stringOf(run.order, models.SalesOrderFieldSalesChannelId),
		OutletReference:     stringOf(run.order, models.SalesOrderFieldSalesPointId),
	}, lines)
}

// untaxedAfter describes each line by its untaxed value after the promotions, for the exclusive
// tax pass: its tax is then the tax the discounted line really carries.
func untaxedAfter(lines []pricing.LineResult, quote *PromotionQuoteResult) []pricing.LineResult {
	remaining := map[string]decimal.Decimal{}
	for _, line := range quote.outcome.Lines {
		remaining[line.Key] = line.RemainingUntaxedAmount
	}
	untaxed := map[string]decimal.Decimal{}
	for _, line := range quote.engine.Lines {
		untaxed[line.Key] = line.UntaxedAmount
	}
	described := make([]pricing.LineResult, 0, len(lines))
	for _, line := range lines {
		amount, found := remaining[line.Key]
		if !found {
			amount = untaxed[line.Key]
		}
		unitPrice := decimal.Zero
		if line.Quantity.IsPositive() {
			unitPrice = amount.Div(line.Quantity)
		}
		described = append(described, pricing.LineResult{
			Key: line.Key, LineNumber: line.LineNumber, ProductVariantId: line.ProductVariantId,
			UomId: line.UomId, Quantity: line.Quantity, EffectiveUnitPrice: unitPrice,
			GrossAmount: amount, NetAmount: amount,
		})
	}
	return described
}

func taxAmounts(tax *BasketTax) map[string]decimal.Decimal {
	amounts := map[string]decimal.Decimal{}
	for key, lineTax := range tax.ByLineKey {
		amounts[key] = lineTax.Amount
	}
	return amounts
}

func lineEffectsOf(quote *PromotionQuoteResult) []pricing.PromotionLineEffect {
	effects := []pricing.PromotionLineEffect{}
	for _, program := range quote.outcome.Programs {
		if program.Outcome != promotion.OutcomeApplied {
			continue
		}
		candidate := quote.pc.byPromotionId(program.PromotionId)
		reward := rewardRecordOf(candidate, program.RewardId)
		adjustmentType := pricing.AdjustmentPercentageDiscount
		if stringOf(reward, models.SalesPromotionRewardFieldDiscountMethod) == promotion.MethodFixedAmount {
			adjustmentType = pricing.AdjustmentFixedDiscount
		}
		for _, allocation := range program.Allocations {
			effects = append(effects, pricing.PromotionLineEffect{
				LineKey: allocation.LineKey, RewardId: program.RewardId, AdjustmentType: adjustmentType,
				Description: rewardDescription(candidate, reward), BaseUntaxed: allocation.BaseUntaxedAmount,
				DiscountUntaxed: allocation.DiscountUntaxedAmount, TaxReduction: allocation.TaxReductionAmount,
			})
		}
	}
	return effects
}

func rewardRecordOf(candidate *promotionCandidate, rewardId string) dmodel.DynamicFields {
	if candidate == nil {
		return nil
	}
	for _, reward := range candidate.rewards {
		if stringOf(reward, basemodel.FieldId) == rewardId {
			return reward
		}
	}
	return nil
}

// rewardDescription is what the customer reads on the order: the reward's own wording, else the
// promotion's name.
func rewardDescription(candidate *promotionCandidate, reward dmodel.DynamicFields) string {
	if text := stringOf(reward, models.SalesPromotionRewardFieldOrderDescription); text != "" {
		return text
	}
	if candidate == nil {
		return ""
	}
	return stringOf(candidate.record, models.SalesPromotionFieldName)
}
