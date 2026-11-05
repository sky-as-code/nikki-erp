package services

import (
	"sort"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing/promotion"
)

// collectVouchers adds the promotions unlocked by the order's attached vouchers. A voucher that can
// no longer unlock anything keeps its promotion in the quote as rejected, with the reason, so a
// draft changed after the code was entered explains itself (BR-S05).
func (this *promotionContext) collectVouchers(
	ctx corectx.Context, active map[string]dmodel.DynamicFields, inScope map[string]bool, contextOk bool,
) error {
	attached, err := searchBy(ctx, models.SalesOrderVoucherSchemaName, models.SalesOrderVoucherFieldOrderId, this.orderId)
	if err != nil || len(attached) == 0 {
		return err
	}
	vouchers, err := searchRows(ctx, models.SalesVoucherSchemaName, true,
		inCondition(basemodel.FieldId, idsOf(attached, models.SalesOrderVoucherFieldVoucherId)))
	if err != nil {
		return err
	}
	byId := map[string]dmodel.DynamicFields{}
	for _, voucher := range vouchers {
		byId[stringOf(voucher, basemodel.FieldId)] = voucher
	}
	sort.SliceStable(attached, func(i, j int) bool {
		return stringOf(attached[i], models.SalesOrderVoucherFieldPromotionId) <
			stringOf(attached[j], models.SalesOrderVoucherFieldPromotionId)
	})
	for _, link := range attached {
		promotionId := stringOf(link, models.SalesOrderVoucherFieldPromotionId)
		record := active[promotionId]
		if record == nil {
			if record, err = loadRecord(ctx, models.SalesPromotionSchemaName, basemodel.FieldId, promotionId); err != nil {
				return err
			}
		}
		voucher := byId[stringOf(link, models.SalesOrderVoucherFieldVoucherId)]
		this.addVoucherCandidate(record, voucher, active[promotionId] != nil, inScope[promotionId], contextOk)
	}
	return nil
}

func (this *promotionContext) addVoucherCandidate(
	record, voucher dmodel.DynamicFields, live bool, scoped bool, contextOk bool,
) {
	if record == nil {
		return
	}
	candidate := newCandidate(record)
	candidate.voucher = voucher
	candidate.voucherCode = stringOf(voucher, models.SalesVoucherFieldCode)
	candidate.rejectReason = this.voucherRejectReason(record, voucher, live, scoped, contextOk)
	if candidate.rejectReason != "" {
		this.rejected = append(this.rejected, candidate)
		return
	}
	candidate.program.VoucherId = stringOf(voucher, basemodel.FieldId)
	candidate.program.VoucherSourceOrderId = stringOf(voucher, models.SalesVoucherFieldSourceOrderId)
	this.candidates = append(this.candidates, candidate)
}

// voucherRejectReason checks the base conditions a voucher can never bypass (BR-C01, BR-V04, BR-V05).
func (this *promotionContext) voucherRejectReason(
	record, voucher dmodel.DynamicFields, live bool, scoped bool, contextOk bool,
) string {
	switch {
	case !contextOk:
		return PromotionReasonContextRequired
	case !VoucherUsableAt(voucher, this.now):
		return PromotionReasonNotAvailable
	case !models.PromotionProgramType(stringOf(record, models.SalesPromotionFieldProgramType)).TakesVouchers():
		return PromotionReasonNotAvailable
	case stringOf(voucher, models.SalesVoucherFieldSourceOrderId) == this.orderId:
		return PromotionReasonNotAvailable
	case !voucherCustomerMatches(voucher, this.customerId):
		return PromotionReasonCustomerMismatch
	case !live:
		return PromotionReasonNotAvailable
	case !scoped:
		return PromotionReasonScopeMismatch
	}
	return ""
}

// VoucherUsableAt is the voucher's own half of BR-V05: present, not archived, active, inside its
// window and with use left. The promotion's half is checked by candidate loading.
func VoucherUsableAt(voucher dmodel.DynamicFields, now time.Time) bool {
	if voucher == nil || boolOf(voucher, basemodel.FieldIsArchived) {
		return false
	}
	if stringOf(voucher, models.SalesVoucherFieldStatus) != string(models.VoucherStatusActive) {
		return false
	}
	if !withinWindow(voucher, models.SalesVoucherFieldValidFrom, models.SalesVoucherFieldExpiresAt, now) {
		return false
	}
	limit := int32Of(voucher, models.SalesVoucherFieldUsageLimit)
	return limit <= 0 || int32Of(voucher, models.SalesVoucherFieldUsedCount) < limit
}

// voucherCustomerMatches is BR-V04: a named voucher needs exactly that customer on the order.
func voucherCustomerMatches(voucher dmodel.DynamicFields, customerId string) bool {
	owner := stringOf(voucher, models.SalesVoucherFieldCustomerId)
	return owner == "" || owner == customerId
}

// loadChildren reads the rules, rewards, selectors and pricelist links of every candidate in one
// query per table.
func (this *promotionContext) loadChildren(ctx corectx.Context) error {
	programs := append(append([]*promotionCandidate{}, this.candidates...), this.issuers...)
	programs = append(programs, this.rejected...)
	if len(programs) == 0 {
		return nil
	}
	ids := make([]string, 0, len(programs))
	for _, candidate := range programs {
		ids = append(ids, candidate.program.Id)
	}
	rules, err := loadOwnedRows(ctx, models.SalesPromotionRuleSchemaName, ids)
	if err != nil {
		return err
	}
	rewards, err := loadOwnedRows(ctx, models.SalesPromotionRewardSchemaName, ids)
	if err != nil {
		return err
	}
	ruleSelectors, err := loadSelectors(ctx, models.SalesPromotionSelectorFieldRuleId,
		models.SalesPromotionRuleProductSchemaName, models.SalesPromotionRuleCategorySchemaName, rules)
	if err != nil {
		return err
	}
	rewardSelectors, err := loadSelectors(ctx, models.SalesPromotionSelectorFieldRewardId,
		models.SalesPromotionRewardProductSchemaName, models.SalesPromotionRewardCategorySchemaName, rewards)
	if err != nil {
		return err
	}
	pricelists, err := searchRows(ctx, models.SalesPromotionPricelistSchemaName, false,
		inCondition(models.SalesPromotionPricelistFieldPromotionId, ids))
	if err != nil {
		return err
	}
	rulesBy, rewardsBy := groupBy(rules, models.SalesPromotionRuleFieldPromotionId),
		groupBy(rewards, models.SalesPromotionRewardFieldPromotionId)
	pricelistsBy := groupBy(pricelists, models.SalesPromotionPricelistFieldPromotionId)
	for _, candidate := range programs {
		id := candidate.program.Id
		candidate.rewards = rewardsBy[id]
		candidate.program.Rules = toEngineRules(rulesBy[id], ruleSelectors)
		candidate.program.Rewards = toEngineRewards(rewardsBy[id], rewardSelectors)
		candidate.program.PricelistIds = idsOf(pricelistsBy[id], models.SalesPromotionPricelistFieldPricelistId)
	}
	return nil
}

func loadOwnedRows(ctx corectx.Context, schemaName string, promotionIds []string) ([]dmodel.DynamicFields, error) {
	return searchRows(ctx, schemaName, false, inCondition(models.SalesPromotionRuleFieldPromotionId, promotionIds))
}

// loadSelectors reads the product and category selectors of a set of owners into one Selector per
// owner id.
func loadSelectors(
	ctx corectx.Context, ownerField, productSchema, categorySchema string, owners []dmodel.DynamicFields,
) (map[string]promotion.Selector, error) {
	selectors := map[string]promotion.Selector{}
	ownerIds := idsOf(owners, basemodel.FieldId)
	if len(ownerIds) == 0 {
		return selectors, nil
	}
	products, err := searchRows(ctx, productSchema, false, inCondition(ownerField, ownerIds))
	if err != nil {
		return nil, err
	}
	categories, err := searchRows(ctx, categorySchema, false, inCondition(ownerField, ownerIds))
	if err != nil {
		return nil, err
	}
	for _, row := range products {
		owner := stringOf(row, ownerField)
		selector := selectors[owner]
		if id := stringOf(row, models.SalesPromotionSelectorFieldTemplateId); id != "" {
			selector.ProductTemplateIds = append(selector.ProductTemplateIds, id)
		}
		if id := stringOf(row, models.SalesPromotionSelectorFieldVariantId); id != "" {
			selector.ProductVariantIds = append(selector.ProductVariantIds, id)
		}
		selectors[owner] = selector
	}
	for _, row := range categories {
		owner := stringOf(row, ownerField)
		selector := selectors[owner]
		selector.CategoryIds = append(selector.CategoryIds, stringOf(row, models.SalesPromotionSelectorFieldCategoryId))
		selectors[owner] = selector
	}
	return selectors, nil
}

func toEngineRules(rows []dmodel.DynamicFields, selectors map[string]promotion.Selector) []promotion.Rule {
	rules := make([]promotion.Rule, 0, len(rows))
	for _, row := range rows {
		id := stringOf(row, basemodel.FieldId)
		rules = append(rules, promotion.Rule{
			Id: id, Phase: stringOf(row, models.SalesPromotionRuleFieldPhase),
			MinQuantity:        decimalOf(row, models.SalesPromotionRuleFieldMinQuantity),
			MinAmount:          decimalOf(row, models.SalesPromotionRuleFieldMinAmount),
			AmountTaxBasis:     stringOf(row, models.SalesPromotionRuleFieldAmountTaxBasis),
			QualificationUomId: qualificationUomOf(row),
			Selector:           selectors[id],
		})
	}
	return rules
}

// qualificationUomOf is set only on a Buy X Get Y rule (BR §8.4); elsewhere it must not switch the
// engine to qualification-unit counting.
func qualificationUomOf(row dmodel.DynamicFields) string {
	return stringOf(row, models.SalesPromotionRuleFieldQualificationUomId)
}

func toEngineRewards(rows []dmodel.DynamicFields, selectors map[string]promotion.Selector) []promotion.Reward {
	rewards := make([]promotion.Reward, 0, len(rows))
	for _, row := range rows {
		id := stringOf(row, basemodel.FieldId)
		rewards = append(rewards, promotion.Reward{
			Id: id, RewardType: stringOf(row, models.SalesPromotionRewardFieldRewardType),
			DiscountMethod:       stringOf(row, models.SalesPromotionRewardFieldDiscountMethod),
			DiscountValue:        decimalOf(row, models.SalesPromotionRewardFieldDiscountValue),
			DiscountScope:        stringOf(row, models.SalesPromotionRewardFieldDiscountScope),
			MaxDiscountAmount:    decimalOf(row, models.SalesPromotionRewardFieldMaxDiscountAmount),
			Selector:             selectors[id],
			GiftProductVariantId: stringOf(row, models.SalesPromotionRewardFieldGiftProductVariantId),
			GiftQuantity:         decimalOf(row, models.SalesPromotionRewardFieldGiftQuantity),
			GiftUomId:            stringOf(row, models.SalesPromotionRewardFieldGiftUomId),
		})
	}
	return rewards
}
