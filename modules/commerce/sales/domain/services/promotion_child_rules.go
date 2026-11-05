package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
)

// Rules and rewards are configuration of their promotion: a write is refused while the promotion
// is active and bumps its config_version in the same transaction (BR-L02). Each kind adds its own
// checks against the promotion's type.

// childRuleFn checks a child as it will be against its promotion, appending what it refuses.
type childRuleFn func(promotion, merged dmodel.DynamicFields, vErrs *ft.ClientErrors)

// childClearFn returns the fields that no longer apply to the merged child, to be written as null.
type childClearFn func(promotion, merged dmodel.DynamicFields) []string

// promotionChildService is the write path shared by rules and rewards.
type promotionChildService struct {
	dynamicresource.CrudDomainService
	schemaName string
	check      childRuleFn
	clear      childClearFn
}

func (this *promotionChildService) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand, options ...dynamicresource.CreateOptions,
) (*dynamicresource.CreateResult, error) {
	var result *dynamicresource.CreateResult
	refused, err := dynamicresource.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			promotionId := stringOf(cmd, models.SalesPromotionRuleFieldPromotionId)
			if vErrs, err := this.prepare(tranxCtx, promotionId, nil, cmd); err != nil || vErrs != nil {
				return vErrs, err
			}
			var err error
			result, err = this.CrudDomainService.Create(tranxCtx, cmd, options...)
			if vErrs, err := clientErrorsOf(result, err); err != nil || vErrs != nil {
				return vErrs, err
			}
			return bumpPromotionConfig(tranxCtx, promotionId)
		})
	if err != nil || refused != nil {
		return createRefusal(refused, err)
	}
	return result, nil
}

func (this *promotionChildService) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand, options ...dynamicresource.UpdateOptions,
) (*dynamicresource.MutateResult, error) {
	var result *dynamicresource.MutateResult
	refused, err := dynamicresource.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			stored, vErrs, err := this.loadChild(tranxCtx, stringOf(cmd, basemodel.FieldId))
			if err != nil || vErrs != nil {
				return vErrs, err
			}
			promotionId := stringOf(stored, models.SalesPromotionRuleFieldPromotionId)
			if vErrs, err := this.prepare(tranxCtx, promotionId, stored, cmd); err != nil || vErrs != nil {
				return vErrs, err
			}
			result, err = this.CrudDomainService.Update(tranxCtx, cmd, options...)
			if vErrs, err := mutateErrorsOf(result, err); err != nil || vErrs != nil {
				return vErrs, err
			}
			return bumpPromotionConfig(tranxCtx, promotionId)
		})
	if err != nil || refused != nil {
		return mutateRefusal(refused, err)
	}
	return result, nil
}

func (this *promotionChildService) Delete(
	ctx corectx.Context, cmd dynamicresource.DeleteCommand, options ...dynamicresource.DeleteOptions,
) (*dynamicresource.MutateResult, error) {
	var result *dynamicresource.MutateResult
	refused, err := dynamicresource.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			stored, vErrs, err := this.loadChild(tranxCtx, stringOf(cmd, basemodel.FieldId))
			if err != nil || vErrs != nil {
				return vErrs, err
			}
			if vErrs, err := assertChildDeletable(tranxCtx, this.schemaName, stored); err != nil || vErrs != nil {
				return vErrs, err
			}
			result, err = this.CrudDomainService.Delete(tranxCtx, cmd, options...)
			if vErrs, err := mutateErrorsOf(result, err); err != nil || vErrs != nil {
				return vErrs, err
			}
			return bumpPromotionConfig(tranxCtx, stringOf(stored, models.SalesPromotionRuleFieldPromotionId))
		})
	if err != nil || refused != nil {
		return mutateRefusal(refused, err)
	}
	return result, nil
}

func (this *promotionChildService) loadChild(
	ctx corectx.Context, childId string,
) (dmodel.DynamicFields, *ft.ClientErrors, error) {
	stored, err := loadRecord(ctx, this.schemaName, basemodel.FieldId, childId)
	if err != nil {
		return nil, nil, err
	}
	if stored == nil {
		return nil, &notFoundResult(this.schemaName, childId).ClientErrors, nil
	}
	return stored, nil, nil
}

// prepare checks the parent promotion, then clears the fields that no longer apply and runs the
// kind's own rule against the child as it will be.
func (this *promotionChildService) prepare(
	ctx corectx.Context, promotionId string, stored, cmd dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	promotion, err := loadRecord(ctx, models.SalesPromotionSchemaName, basemodel.FieldId, promotionId)
	if err != nil {
		return nil, err
	}
	if promotion == nil {
		return violationsOf(models.SalesPromotionRuleFieldPromotionId, "promotion.not_found",
			"no promotion with id '"+promotionId+"'"), nil
	}
	if vErrs := assertPromotionEditable(promotion); vErrs != nil {
		return vErrs, nil
	}
	if orgId := stringOf(cmd, basemodel.FieldOrgId); orgId != "" &&
		orgId != stringOf(promotion, basemodel.FieldOrgId) {
		return violationsOf(models.SalesPromotionRuleFieldPromotionId, "promotion.not_found",
			"no promotion with id '"+promotionId+"'"), nil
	}
	merged := mergedFields(stored, cmd)
	for _, field := range this.clear(promotion, merged) {
		if stored == nil {
			delete(cmd, field)
		} else if cmd[field] != nil || stored[field] != nil {
			cmd[field] = nil
		}
		delete(merged, field)
	}
	vErrs := ft.NewClientErrors()
	this.check(promotion, merged, vErrs)
	if vErrs.Count() > 0 {
		return vErrs, nil
	}
	return nil, nil
}

// assertChildDeletable refuses deleting a reward a confirmed order already references: history
// keeps it, and the promotion is edited by changing the configuration instead (BR §8.11).
func assertChildDeletable(
	ctx corectx.Context, schemaName string, stored dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	promotion, err := loadRecord(ctx, models.SalesPromotionSchemaName, basemodel.FieldId,
		stringOf(stored, models.SalesPromotionRuleFieldPromotionId))
	if err != nil || promotion == nil {
		return nil, err
	}
	if vErrs := assertPromotionEditable(promotion); vErrs != nil {
		return vErrs, nil
	}
	if schemaName != models.SalesPromotionRewardSchemaName {
		return nil, nil
	}
	used, err := searchBy(ctx, models.SalesPromotionApplicationSchemaName,
		models.SalesPromotionApplicationFieldRewardId, stringOf(stored, basemodel.FieldId))
	if err != nil {
		return nil, err
	}
	if len(used) > 0 {
		return violationsOf(basemodel.FieldId, "promotion.immutable_after_use",
			"a reward a confirmed order received cannot be deleted"), nil
	}
	return nil, nil
}

// checkRule: only next_order_coupon has issue rules (BR-V08).
func checkRule(promotion, merged dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	programType := models.PromotionProgramType(stringOf(promotion, models.SalesPromotionFieldProgramType))
	if stringOf(merged, models.SalesPromotionRuleFieldPhase) == string(models.PromotionRulePhaseIssue) &&
		programType != models.PromotionTypeNextOrderCoupon {
		vErrs.Append(*ft.NewBusinessViolation(models.SalesPromotionRuleFieldPhase,
			ft.ErrorKey("promotion.rule_phase_not_allowed", "sales"),
			"only a next-order coupon promotion has issue rules"))
	}
}

// clearRule: the qualification unit belongs to buy_x_get_y alone (BR §8.4).
func clearRule(promotion, _ dmodel.DynamicFields) []string {
	if stringOf(promotion, models.SalesPromotionFieldProgramType) == string(models.PromotionTypeBuyXGetY) {
		return nil
	}
	return []string{models.SalesPromotionRuleFieldQualificationUomId}
}

// checkReward: buy_x_get_y gives products only, and a percentage is at most 100.
func checkReward(promotion, merged dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	rewardType := stringOf(merged, models.SalesPromotionRewardFieldRewardType)
	if stringOf(promotion, models.SalesPromotionFieldProgramType) == string(models.PromotionTypeBuyXGetY) &&
		rewardType != string(models.PromotionRewardFreeProduct) {
		vErrs.Append(*ft.NewBusinessViolation(models.SalesPromotionRewardFieldRewardType,
			ft.ErrorKey("promotion.reward_type_not_allowed", "sales"),
			"a buy X get Y promotion gives free products only"))
	}
	if stringOf(merged, models.SalesPromotionRewardFieldDiscountMethod) == "percentage" &&
		decimalOf(merged, models.SalesPromotionRewardFieldDiscountValue).GreaterThan(hundredPercent) {
		vErrs.Append(*ft.NewBusinessViolation(models.SalesPromotionRewardFieldDiscountValue,
			ft.ErrorKey("promotion.percentage_out_of_range", "sales"), "a percentage is at most 100"))
	}
}

// rewardFieldsByType are the fields each reward type uses; every other type field is null
// (BR §8.5), so a value left over from an earlier type never takes effect silently.
var rewardFieldsByType = map[string][]string{
	string(models.PromotionRewardDiscount): {
		models.SalesPromotionRewardFieldDiscountMethod, models.SalesPromotionRewardFieldDiscountValue,
		models.SalesPromotionRewardFieldDiscountScope, models.SalesPromotionRewardFieldMaxDiscountAmount,
	},
	string(models.PromotionRewardFreeProduct): {
		models.SalesPromotionRewardFieldGiftProductVariantId, models.SalesPromotionRewardFieldGiftQuantity,
		models.SalesPromotionRewardFieldGiftUomId,
	},
	string(models.PromotionRewardFreeShipping): {models.SalesPromotionRewardFieldMaxDiscountAmount},
}

func clearReward(_ dmodel.DynamicFields, merged dmodel.DynamicFields) []string {
	kept := map[string]bool{}
	for _, field := range rewardFieldsByType[stringOf(merged, models.SalesPromotionRewardFieldRewardType)] {
		kept[field] = true
	}
	cleared := make([]string, 0, 8)
	for _, fields := range rewardFieldsByType {
		for _, field := range fields {
			if !kept[field] && !containsString(cleared, field) {
				cleared = append(cleared, field)
			}
		}
	}
	return cleared
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
