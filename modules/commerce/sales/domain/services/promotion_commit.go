package services

import (
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing/promotion"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
)

// Committing the promotions of a confirm (BR-T02..T05, plan §3.3). Runs inside the confirm's own
// transaction, so a failure anywhere — an exhausted use, the bill, the stage event — rolls back the
// uses taken and the coupons issued along with the status (BR-T04, AC-28).

// promotionRefusal travels out of the confirm transaction as an error so the transaction rolls back;
// confirm turns it back into client errors.
type promotionCommitRefusal struct {
	cErrs ft.ClientErrors
}

func (this *promotionCommitRefusal) Error() string {
	return this.cErrs.ToError().Error()
}

func commitRefusalOf(err error) *ft.ClientErrors {
	var refusal *promotionCommitRefusal
	if errors.As(err, &refusal) {
		return &refusal.cErrs
	}
	return nil
}

// ConfirmedPromotions is what a confirm committed, answered again on a replay (BR-T05).
type ConfirmedPromotions struct {
	ApplicationIds []string        `json:"application_ids"`
	IssuedVouchers []IssuedVoucher `json:"issued_vouchers"`
}

// confirmationRevision is always 1: orders have no reset-to-draft (D1).
const confirmationRevision int32 = 1

func commitPromotions(
	ctx corectx.Context, order dmodel.DynamicFields, quote *PromotionQuoteResult, at time.Time,
) (*ConfirmedPromotions, error) {
	committed := &ConfirmedPromotions{ApplicationIds: []string{}, IssuedVouchers: []IssuedVoucher{}}
	if quote == nil {
		return committed, nil
	}
	applied := appliedPrograms(quote.outcome)
	if err := consumeUsages(ctx, applied); err != nil {
		return nil, err
	}
	for _, program := range applied {
		applicationId, err := insertApplication(ctx, order, quote, program, at)
		if err != nil {
			return nil, err
		}
		committed.ApplicationIds = append(committed.ApplicationIds, applicationId)
	}
	issued, err := issueNextOrderCoupons(ctx, order, quote, at)
	if err != nil {
		return nil, err
	}
	committed.IssuedVouchers = issued
	repo, err := repoFor(models.SalesOrderSchemaName)
	if err != nil {
		return nil, err
	}
	_, err = repo.Update(ctx, dmodel.DynamicFields{
		models.SalesOrderFieldId:                   stringOf(order, models.SalesOrderFieldId),
		models.SalesOrderFieldConfirmationRevision: confirmationRevision,
	})
	return committed, err
}

// appliedPrograms are the programs that become applications, in promotion id order so concurrent
// confirms take their row locks in the same order.
func appliedPrograms(outcome promotion.Result) []promotion.ProgramResult {
	applied := []promotion.ProgramResult{}
	for _, program := range outcome.Programs {
		if program.Outcome == promotion.OutcomeApplied {
			applied = append(applied, program)
		}
	}
	sort.SliceStable(applied, func(i, j int) bool { return applied[i].PromotionId < applied[j].PromotionId })
	return applied
}

// consumeUsages takes one use of each promotion, then of each voucher (D5). A refusal is a client
// error: the confirm rolls back and the order stays a draft.
func consumeUsages(ctx corectx.Context, applied []promotion.ProgramResult) error {
	for _, program := range applied {
		if err := consumeOne(ctx, models.SalesPromotionSchemaName, program.PromotionId,
			PromotionErrUsageExhausted, "the promotion has no use left"); err != nil {
			return err
		}
	}
	for _, program := range applied {
		if program.VoucherId == "" {
			continue
		}
		if err := consumeOne(ctx, models.SalesVoucherSchemaName, program.VoucherId,
			PromotionErrVoucherUsageExhausted, "the code has no use left"); err != nil {
			return err
		}
	}
	return nil
}

func usageCounterOf(schemaName string) (itPromotion.UsageCounterRepository, error) {
	repo, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}
	counter, ok := repo.(itPromotion.UsageCounterRepository)
	if !ok {
		return nil, errors.New("the " + schemaName + " repository counts no usage")
	}
	return counter, nil
}

func consumeOne(ctx corectx.Context, schemaName, id, key, message string) error {
	counter, err := usageCounterOf(schemaName)
	if err != nil {
		return err
	}
	taken, err := counter.TryConsumeUsage(ctx, id)
	if err != nil {
		return err
	}
	if !taken {
		return &promotionCommitRefusal{cErrs: *promotionRefusal("", key, message, map[string]any{"promotion_id": id})}
	}
	return nil
}

func insertApplication(
	ctx corectx.Context, order dmodel.DynamicFields, quote *PromotionQuoteResult,
	program promotion.ProgramResult, at time.Time,
) (string, error) {
	repo, err := repoFor(models.SalesPromotionApplicationSchemaName)
	if err != nil {
		return "", err
	}
	id, err := model.NewId()
	if err != nil {
		return "", err
	}
	candidate := quote.pc.byPromotionId(program.PromotionId)
	fields := dmodel.DynamicFields{
		basemodel.FieldId:                                          string(*id),
		basemodel.FieldOrgId:                                       stringOf(order, basemodel.FieldOrgId),
		models.SalesPromotionApplicationFieldOrderId:               stringOf(order, models.SalesOrderFieldId),
		models.SalesPromotionApplicationFieldPromotionId:           program.PromotionId,
		models.SalesPromotionApplicationFieldRewardId:              program.RewardId,
		models.SalesPromotionApplicationFieldStatus:                string(models.PromotionApplicationConfirmed),
		models.SalesPromotionApplicationFieldIsConfirmed:           true,
		models.SalesPromotionApplicationFieldConfirmationRevision:  confirmationRevision,
		models.SalesPromotionApplicationFieldConfigVersion:         program.ConfigVersion,
		models.SalesPromotionApplicationFieldQualificationSnapshot: qualificationSnapshot(program.Qualification),
		models.SalesPromotionApplicationFieldRewardSnapshot: rewardSnapshot(
			rewardRecordOf(candidate, program.RewardId)),
		models.SalesPromotionApplicationFieldDiscountUntaxed: program.DiscountUntaxedAmount,
		models.SalesPromotionApplicationFieldTaxReduction:    program.TaxReductionAmount,
		models.SalesPromotionApplicationFieldDiscountTotal:   program.DiscountTotalAmount(),
		models.SalesPromotionApplicationFieldAppliedAt:       model.ModelDateTime(at),
	}
	if program.VoucherId != "" {
		fields[models.SalesPromotionApplicationFieldVoucherId] = program.VoucherId
	}
	if _, err := repo.Insert(ctx, fields); err != nil {
		return "", err
	}
	return string(*id), insertAllocations(ctx, order, quote, program, string(*id))
}

func qualificationSnapshot(qualification promotion.Qualification) map[string]any {
	return map[string]any{
		"rule_id":           qualification.RuleId,
		"eligible_line_ids": qualification.EligibleLineKeys,
		"eligible_quantity": qualification.EligibleQuantity.String(),
		"eligible_amount":   qualification.EligibleAmount.String(),
		"sets":              qualification.Sets,
	}
}

// rewardSnapshot freezes the reward as it was applied, so a later edit of the promotion does not
// change the confirmed order (BR-T06).
func rewardSnapshot(reward dmodel.DynamicFields) map[string]any {
	snapshot := map[string]any{}
	for _, field := range []string{
		basemodel.FieldId, models.SalesPromotionRewardFieldRewardType, models.SalesPromotionRewardFieldOrderDescription,
		models.SalesPromotionRewardFieldDiscountMethod, models.SalesPromotionRewardFieldDiscountScope,
		models.SalesPromotionRewardFieldGiftProductVariantId, models.SalesPromotionRewardFieldGiftUomId,
	} {
		snapshot[field] = stringOf(reward, field)
	}
	for _, field := range []string{
		models.SalesPromotionRewardFieldDiscountValue, models.SalesPromotionRewardFieldMaxDiscountAmount,
		models.SalesPromotionRewardFieldGiftQuantity,
	} {
		snapshot[field] = decimalOf(reward, field).String()
	}
	return snapshot
}

// insertAllocations writes one row per discounted line, plus the gift line's own row, and links
// the gift line to its application (BR §8.9).
func insertAllocations(
	ctx corectx.Context, order dmodel.DynamicFields, quote *PromotionQuoteResult,
	program promotion.ProgramResult, applicationId string,
) error {
	rows := []dmodel.DynamicFields{}
	uoms := lineUoms(quote.engine)
	for _, allocation := range program.Allocations {
		rows = append(rows, promotionAllocationRow(order, applicationId, allocation.LineKey,
			promotion.AllocationMerchandiseDiscount, allocation.EligibleQuantity, uoms[allocation.LineKey],
			allocation.BaseUntaxedAmount, allocation.DiscountUntaxedAmount, allocation.TaxReductionAmount,
			program.CalculationSequence))
	}
	giftLineId := quote.giftLineIds[program.PromotionId]
	if program.Gift != nil && giftLineId != "" {
		rows = append(rows, promotionAllocationRow(order, applicationId, giftLineId, promotion.AllocationFreeProduct,
			program.Gift.Quantity, program.Gift.UomId, decimal.Zero, decimal.Zero, decimal.Zero,
			program.CalculationSequence))
	}
	repo, err := repoFor(models.SalesPromotionAllocationSchemaName)
	if err != nil {
		return err
	}
	for _, row := range rows {
		id, err := model.NewId()
		if err != nil {
			return err
		}
		row[basemodel.FieldId] = string(*id)
		if _, err := repo.Insert(ctx, row); err != nil {
			return err
		}
	}
	if giftLineId == "" {
		return nil
	}
	lineRepo, err := repoFor(models.SalesOrderLineSchemaName)
	if err != nil {
		return err
	}
	_, err = lineRepo.Update(ctx, dmodel.DynamicFields{
		models.SalesOrderLineFieldId: giftLineId, models.SalesOrderLineFieldPromotionApplicationId: applicationId,
	})
	return err
}

func lineUoms(input promotion.Input) map[string]string {
	uoms := map[string]string{}
	for _, line := range input.Lines {
		uoms[line.Key] = line.UomId
	}
	return uoms
}

func promotionAllocationRow(
	order dmodel.DynamicFields, applicationId, lineId, allocationType string, quantity decimal.Decimal,
	uomId string, base, untaxed, taxReduction decimal.Decimal, sequence int32,
) dmodel.DynamicFields {
	row := dmodel.DynamicFields{
		basemodel.FieldOrgId:                                    stringOf(order, basemodel.FieldOrgId),
		models.SalesPromotionAllocationFieldApplicationId:       applicationId,
		models.SalesPromotionAllocationFieldOrderLineId:         lineId,
		models.SalesPromotionAllocationFieldAllocationType:      allocationType,
		models.SalesPromotionAllocationFieldEligibleQuantity:    quantity,
		models.SalesPromotionAllocationFieldBaseUntaxed:         base,
		models.SalesPromotionAllocationFieldDiscountUntaxed:     untaxed,
		models.SalesPromotionAllocationFieldTaxReduction:        taxReduction,
		models.SalesPromotionAllocationFieldDiscountTotal:       untaxed.Add(taxReduction),
		models.SalesPromotionAllocationFieldCalculationSequence: sequence,
	}
	if uomId != "" {
		row[models.SalesPromotionAllocationFieldUomId] = uomId
	}
	return row
}
