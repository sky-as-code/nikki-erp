package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
)

// SelectPromotionReward records the draft's reward choice for one promotion (BR-R12, D6) and
// answers the fresh quote. The choice is only a wish: the engine still decides whether the promotion
// qualifies and what the reward is worth (BR-T01).
func SelectPromotionReward(
	ctx corectx.Context, orderId, promotionId, rewardId string, decline bool, deps PromotionActionDeps,
) (map[string]any, *ft.ClientErrors, error) {
	if promotionId == "" {
		return nil, promotionRefusal("promotion_id", "promotion.promotion_required", "promotion_id is required"), nil
	}
	if !decline && rewardId == "" {
		return nil, promotionRefusal("reward_id", PromotionErrSelectionRequired,
			"choose a reward, or decline the promotion"), nil
	}
	order, vErrs, err := loadDraftOrder(ctx, orderId)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	if !decline {
		if vErrs, err := assertRewardOfPromotion(ctx, order, promotionId, rewardId); err != nil || vErrs != nil {
			return nil, vErrs, err
		}
	}
	err = withTransaction(ctx, models.SalesOrderPromotionSelectionSchemaName, func(tranxCtx corectx.Context) error {
		return upsertSelection(tranxCtx, order, promotionId, rewardId, decline)
	})
	if err != nil {
		return nil, nil, err
	}
	return PreviewPromotions(ctx, orderId, deps)
}

func assertRewardOfPromotion(
	ctx corectx.Context, order dmodel.DynamicFields, promotionId, rewardId string,
) (*ft.ClientErrors, error) {
	reward, err := loadRecord(ctx, models.SalesPromotionRewardSchemaName, basemodel.FieldId, rewardId)
	if err != nil {
		return nil, err
	}
	if reward == nil || stringOf(reward, models.SalesPromotionRewardFieldPromotionId) != promotionId ||
		stringOf(reward, basemodel.FieldOrgId) != stringOf(order, basemodel.FieldOrgId) {
		return promotionRefusal("reward_id", PromotionErrNotAvailable, "this reward does not belong to the promotion"), nil
	}
	return nil, nil
}

func upsertSelection(ctx corectx.Context, order dmodel.DynamicFields, promotionId, rewardId string, decline bool) error {
	repo, err := repoFor(models.SalesOrderPromotionSelectionSchemaName)
	if err != nil {
		return err
	}
	orderId := stringOf(order, models.SalesOrderFieldId)
	existing, err := searchRows(ctx, models.SalesOrderPromotionSelectionSchemaName, false,
		equalsCondition(models.SalesOrderPromotionSelectionFieldOrderId, orderId),
		equalsCondition(models.SalesOrderPromotionSelectionFieldPromotionId, promotionId))
	if err != nil {
		return err
	}
	fields := dmodel.DynamicFields{
		models.SalesOrderPromotionSelectionFieldDecision: PromotionDecisionAccepted,
		models.SalesOrderPromotionSelectionFieldRewardId: nil,
	}
	if decline {
		fields[models.SalesOrderPromotionSelectionFieldDecision] = PromotionDecisionDeclined
	} else {
		fields[models.SalesOrderPromotionSelectionFieldRewardId] = rewardId
	}
	if len(existing) > 0 {
		fields[basemodel.FieldId] = stringOf(existing[0], basemodel.FieldId)
		_, err = repo.Update(ctx, fields)
		return err
	}
	id, err := model.NewId()
	if err != nil {
		return err
	}
	fields[basemodel.FieldId] = string(*id)
	fields[basemodel.FieldOrgId] = stringOf(order, basemodel.FieldOrgId)
	fields[models.SalesOrderPromotionSelectionFieldOrderId] = orderId
	fields[models.SalesOrderPromotionSelectionFieldPromotionId] = promotionId
	if decline {
		delete(fields, models.SalesOrderPromotionSelectionFieldRewardId)
	}
	_, err = repo.Insert(ctx, fields)
	return err
}
