package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
)

// The promotion's scope: which channels, points and pricelists it may run on (BR §4, §8.3). Each
// link is explicit; "select all current" is the client adding every current id in one call, and a
// channel created later is never added on its own (BR-S02).

var promotionScopeTargets = map[itPromotion.PromotionScope]string{
	itPromotion.PromotionScopeChannels:   models.SalesChannelSchemaName,
	itPromotion.PromotionScopePoints:     models.SalesPointSchemaName,
	itPromotion.PromotionScopePricelists: models.SalesPricelistSchemaName,
}

// ManageScope links and unlinks one kind of scope target. It is refused while the promotion is
// active, refuses a target of another org, and bumps config_version in the same transaction.
func (this *SalesPromotionDomainServiceImpl) ManageScope(
	ctx corectx.Context, scope itPromotion.PromotionScope, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	targetSchema := promotionScopeTargets[scope]
	return dynamicresource.ManageM2m(ctx, this.Repository(), cmd, dynamicresource.ManageM2mParam{
		Action:          "manage promotion " + string(scope),
		DestSchemaName:  targetSchema,
		JunctionIdField: basemodel.FieldId,
		Validate: func(tranxCtx corectx.Context, change dynamicresource.M2mChange, vErrs *ft.ClientErrors) error {
			return assertScopeChange(tranxCtx, targetSchema, change, vErrs)
		},
		AfterChange: func(tranxCtx corectx.Context, change dynamicresource.M2mChange) error {
			refused, err := bumpPromotionConfig(tranxCtx, string(change.SrcId))
			if err != nil {
				return err
			}
			if refused != nil {
				return refused.ToError()
			}
			return nil
		},
	})
}

func assertScopeChange(
	ctx corectx.Context, targetSchema string, change dynamicresource.M2mChange, vErrs *ft.ClientErrors,
) error {
	promotion, err := loadRecord(ctx, models.SalesPromotionSchemaName, basemodel.FieldId, string(change.SrcId))
	if err != nil {
		return err
	}
	if promotion == nil {
		vErrs.Append(*ft.NewNotFoundError(basemodel.FieldId))
		return nil
	}
	if locked := assertPromotionEditable(promotion); locked != nil {
		vErrs.Append(*locked...)
		return nil
	}
	return assertTargetsInOrg(ctx, targetSchema, stringOf(promotion, basemodel.FieldOrgId), change.Added, vErrs)
}

// assertTargetsInOrg refuses an id that does not exist or belongs to another org, with the same
// answer for both so another org's records cannot be probed (BR §8.1).
func assertTargetsInOrg(
	ctx corectx.Context, schemaName, orgId string, ids []model.Id, vErrs *ft.ClientErrors,
) error {
	for _, id := range ids {
		target, err := loadRecord(ctx, schemaName, basemodel.FieldId, string(id))
		if err != nil {
			return err
		}
		if target == nil || stringOf(target, basemodel.FieldOrgId) != orgId {
			vErrs.Append(*ft.NewBusinessViolation(dynamicresource.FieldM2mAdd,
				ft.ErrorKey("promotion.scope_target_invalid", "sales"),
				"'"+string(id)+"' is not a "+schemaName+" of this organization"))
		}
	}
	return nil
}
