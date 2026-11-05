package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
)

// The Activate validator (BR-L01, BR-S02, BR-V08). Each missing piece is its own violation of
// sales:promotion.activation_invalid, named by the field the operator must fix, so the form can
// point at every problem at once rather than one per attempt. Nothing is filled in automatically:
// an empty scope is reported, never widened to "every point" (BR-S02).

const promotionActivationInvalid = "promotion.activation_invalid"

// The pseudo-fields the scope violations name, matching the frontend's assignment tabs.
const (
	activationFieldChannels = "sales_channel_ids"
	activationFieldPoints   = "sales_point_ids"
	activationFieldRules    = "rules"
	activationFieldRewards  = "rewards"
)

func assertPromotionActivatable(
	ctx corectx.Context, promotion dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	promotionId := stringOf(promotion, basemodel.FieldId)
	programType := models.PromotionProgramType(stringOf(promotion, models.SalesPromotionFieldProgramType))

	if boolOf(promotion, basemodel.FieldIsArchived) {
		appendActivation(vErrs, basemodel.FieldIsArchived, "an archived promotion cannot be activated")
	}
	if err := checkActivationScope(ctx, promotionId, vErrs); err != nil {
		return nil, err
	}
	checkActivationWindow(promotion, programType, vErrs)

	rules, err := searchBy(ctx, models.SalesPromotionRuleSchemaName, models.SalesPromotionRuleFieldPromotionId, promotionId)
	if err != nil {
		return nil, err
	}
	checkActivationRules(programType, rules, vErrs)

	rewards, err := searchBy(ctx, models.SalesPromotionRewardSchemaName,
		models.SalesPromotionRewardFieldPromotionId, promotionId)
	if err != nil {
		return nil, err
	}
	if err := checkActivationRewards(ctx, programType, rewards, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() == 0 {
		return nil, nil
	}
	return vErrs, nil
}

func appendActivation(vErrs *ft.ClientErrors, field, message string) {
	vErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey(promotionActivationInvalid, "sales"), message))
}

// checkActivationScope requires at least one channel AND one point (BR-S01, BR-S02).
func checkActivationScope(ctx corectx.Context, promotionId string, vErrs *ft.ClientErrors) error {
	channels, err := searchBy(ctx, models.SalesPromotionChannelSchemaName,
		models.SalesPromotionChannelFieldPromotionId, promotionId)
	if err != nil {
		return err
	}
	if len(channels) == 0 {
		appendActivation(vErrs, activationFieldChannels, "link at least one sales channel")
	}
	points, err := searchBy(ctx, models.SalesPromotionPointSchemaName,
		models.SalesPromotionPointFieldPromotionId, promotionId)
	if err != nil {
		return err
	}
	if len(points) == 0 {
		appendActivation(vErrs, activationFieldPoints, "link at least one sales point")
	}
	return nil
}

func checkActivationWindow(
	promotion dmodel.DynamicFields, programType models.PromotionProgramType, vErrs *ft.ClientErrors,
) {
	startsAt := dateTimeOf(promotion, models.SalesPromotionFieldStartsAt)
	endsAt := dateTimeOf(promotion, models.SalesPromotionFieldEndsAt)
	if startsAt != nil && endsAt != nil && !startsAt.GoTime().Before(endsAt.GoTime()) {
		appendActivation(vErrs, models.SalesPromotionFieldEndsAt, "ends_at must be after starts_at")
	}
	if stringOf(promotion, models.SalesPromotionFieldCurrencyId) == "" {
		appendActivation(vErrs, models.SalesPromotionFieldCurrencyId, "a currency is required")
	}
	if programType == models.PromotionTypeNextOrderCoupon &&
		int32Of(promotion, models.SalesPromotionFieldNextVoucherValidityDays) <= 0 {
		appendActivation(vErrs, models.SalesPromotionFieldNextVoucherValidityDays,
			"a next-order coupon promotion needs the coupon's validity in days")
	}
}

// checkActivationRules: every type needs an apply rule; next_order_coupon also needs an issue rule
// and is the only type allowed one (BR-V08); buy_x_get_y counts sets in a qualification unit with
// a positive set size (BR-C07, BR-C08).
func checkActivationRules(
	programType models.PromotionProgramType, rules []dmodel.DynamicFields, vErrs *ft.ClientErrors,
) {
	applyCount, issueCount := 0, 0
	for _, rule := range rules {
		if stringOf(rule, models.SalesPromotionRuleFieldPhase) == string(models.PromotionRulePhaseIssue) {
			issueCount++
			continue
		}
		applyCount++
		if programType == models.PromotionTypeBuyXGetY {
			checkBuyXGetYRule(rule, vErrs)
		}
	}
	if applyCount == 0 {
		appendActivation(vErrs, activationFieldRules, "add at least one rule of the apply phase")
	}
	if programType == models.PromotionTypeNextOrderCoupon && issueCount == 0 {
		appendActivation(vErrs, activationFieldRules, "a next-order coupon promotion needs an issue rule")
	}
	if programType != models.PromotionTypeNextOrderCoupon && issueCount > 0 {
		appendActivation(vErrs, activationFieldRules, "only a next-order coupon promotion has issue rules")
	}
}

func checkBuyXGetYRule(rule dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	if stringOf(rule, models.SalesPromotionRuleFieldQualificationUomId) == "" {
		appendActivation(vErrs, models.SalesPromotionRuleFieldQualificationUomId,
			"a buy X get Y rule needs the unit its sets are counted in")
	}
	if !decimalOf(rule, models.SalesPromotionRuleFieldMinQuantity).IsPositive() {
		appendActivation(vErrs, models.SalesPromotionRuleFieldMinQuantity,
			"a buy X get Y rule needs a set size above zero")
	}
}

// checkActivationRewards: at least one reward; buy_x_get_y gives products only; a discount on
// specific products must name them (BR §8.5).
func checkActivationRewards(
	ctx corectx.Context, programType models.PromotionProgramType, rewards []dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) error {
	if len(rewards) == 0 {
		appendActivation(vErrs, activationFieldRewards, "add at least one reward")
	}
	for _, reward := range rewards {
		rewardType := stringOf(reward, models.SalesPromotionRewardFieldRewardType)
		if programType == models.PromotionTypeBuyXGetY && rewardType != string(models.PromotionRewardFreeProduct) {
			appendActivation(vErrs, models.SalesPromotionRewardFieldRewardType,
				"a buy X get Y promotion gives free products only")
		}
		if stringOf(reward, models.SalesPromotionRewardFieldDiscountScope) !=
			models.PromotionDiscountScopeSpecificProducts {
			continue
		}
		named, err := rewardHasSelectors(ctx, stringOf(reward, basemodel.FieldId))
		if err != nil {
			return err
		}
		if !named {
			appendActivation(vErrs, models.SalesPromotionRewardFieldDiscountScope,
				"a discount on specific products must name at least one product or category")
		}
	}
	return nil
}

func rewardHasSelectors(ctx corectx.Context, rewardId string) (bool, error) {
	for _, schemaName := range []string{
		models.SalesPromotionRewardProductSchemaName, models.SalesPromotionRewardCategorySchemaName,
	} {
		rows, err := searchBy(ctx, schemaName, models.SalesPromotionSelectorFieldRewardId, rewardId)
		if err != nil {
			return false, err
		}
		if len(rows) > 0 {
			return true, nil
		}
	}
	return false, nil
}
