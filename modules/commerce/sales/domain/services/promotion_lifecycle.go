package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
)

// Activate validates the whole configuration and moves the promotion to active. A promotion may
// be activated before starts_at; it simply applies to no order yet (BR-L01). Activating an active
// promotion is a no-op that answers the record.
func (this *SalesPromotionDomainServiceImpl) Activate(
	ctx corectx.Context, promotionId string,
) (*itPromotion.RecordResult, error) {
	refused, err := dynamicresource.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			stored, err := loadRecord(tranxCtx, models.SalesPromotionSchemaName, basemodel.FieldId, promotionId)
			if err != nil || stored == nil {
				return promotionMissing(promotionId, err)
			}
			if stringOf(stored, models.SalesPromotionFieldStatus) == string(models.PromotionStatusActive) {
				return nil, nil
			}
			vErrs, err := assertPromotionActivatable(tranxCtx, stored)
			if err != nil || vErrs != nil {
				return vErrs, err
			}
			return nil, writeChanges(tranxCtx, models.SalesPromotionSchemaName, stored, dmodel.DynamicFields{
				models.SalesPromotionFieldStatus: string(models.PromotionStatusActive),
			})
		})
	return lifecycleAnswer(ctx, promotionId, refused, err)
}

// Suspend stops new applications and issuance while keeping every confirmed order as it is; it is
// what makes the configuration editable again (BR-L02). Only an active promotion is suspended; a
// suspended one answers as it is.
func (this *SalesPromotionDomainServiceImpl) Suspend(
	ctx corectx.Context, promotionId string,
) (*itPromotion.RecordResult, error) {
	refused, err := dynamicresource.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			stored, err := loadRecord(tranxCtx, models.SalesPromotionSchemaName, basemodel.FieldId, promotionId)
			if err != nil || stored == nil {
				return promotionMissing(promotionId, err)
			}
			switch models.PromotionStatus(stringOf(stored, models.SalesPromotionFieldStatus)) {
			case models.PromotionStatusSuspended:
				return nil, nil
			case models.PromotionStatusDraft:
				return violationsOf(models.SalesPromotionFieldStatus, "promotion.invalid_transition",
					"only an active promotion can be suspended"), nil
			}
			return nil, writeChanges(tranxCtx, models.SalesPromotionSchemaName, stored, dmodel.DynamicFields{
				models.SalesPromotionFieldStatus: string(models.PromotionStatusSuspended),
			})
		})
	return lifecycleAnswer(ctx, promotionId, refused, err)
}

func promotionMissing(promotionId string, err error) (*ft.ClientErrors, error) {
	if err != nil {
		return nil, err
	}
	return &notFoundResult(models.SalesPromotionSchemaName, promotionId).ClientErrors, nil
}

func lifecycleAnswer(
	ctx corectx.Context, promotionId string, refused *ft.ClientErrors, err error,
) (*itPromotion.RecordResult, error) {
	if err != nil {
		return nil, err
	}
	if refused != nil {
		return recordRefusal(refused), nil
	}
	return recordResult(ctx, models.SalesPromotionSchemaName, promotionId)
}

// promotionCopiedFields are the configuration a duplicate takes over; the reference, status, usage
// and history start afresh (BR §9.1).
var promotionCopiedFields = []string{
	basemodel.FieldOrgId, models.SalesPromotionFieldName, models.SalesPromotionFieldDescription,
	models.SalesPromotionFieldProgramType, models.SalesPromotionFieldCurrencyId,
	models.SalesPromotionFieldStartsAt, models.SalesPromotionFieldEndsAt, models.SalesPromotionFieldUsageLimit,
	models.SalesPromotionFieldSequence, models.SalesPromotionFieldNextVoucherValidityDays,
}

// Duplicate creates a draft with a new reference and copies the rules, rewards, selectors and
// scope. Vouchers, usage and history are never copied.
func (this *SalesPromotionDomainServiceImpl) Duplicate(
	ctx corectx.Context, sourceId string,
) (*itPromotion.RecordResult, error) {
	var newId string
	refused, err := dynamicresource.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			source, err := loadRecord(tranxCtx, models.SalesPromotionSchemaName, basemodel.FieldId, sourceId)
			if err != nil || source == nil {
				return promotionMissing(sourceId, err)
			}
			created, err := this.createStamped(tranxCtx, pickFields(source, promotionCopiedFields),
				[]string{"duplicate", sourceId})
			if vErrs, err := clientErrorsOf(created, err); err != nil || vErrs != nil {
				return vErrs, err
			}
			newId = stringOf(created.Data, basemodel.FieldId)
			return copyPromotionChildren(tranxCtx, sourceId, newId)
		})
	if err != nil || refused != nil {
		return lifecycleAnswer(ctx, sourceId, refused, err)
	}
	return recordResult(ctx, models.SalesPromotionSchemaName, newId)
}

func pickFields(source dmodel.DynamicFields, fields []string) dmodel.DynamicFields {
	picked := dmodel.DynamicFields{}
	for _, field := range fields {
		if value, ok := source[field]; ok && value != nil {
			picked[field] = value
		}
	}
	return picked
}

// copyPromotionChildren copies the scope rows, then each rule and reward with its selectors.
func copyPromotionChildren(ctx corectx.Context, sourceId, targetId string) (*ft.ClientErrors, error) {
	scopes := map[string]string{
		models.SalesPromotionChannelSchemaName:   models.SalesPromotionChannelFieldChannelId,
		models.SalesPromotionPointSchemaName:     models.SalesPromotionPointFieldPointId,
		models.SalesPromotionPricelistSchemaName: models.SalesPromotionPricelistFieldPricelistId,
	}
	for schemaName, targetField := range scopes {
		if err := copyRows(ctx, schemaName, models.SalesPromotionChannelFieldPromotionId, sourceId,
			dmodel.DynamicFields{models.SalesPromotionChannelFieldPromotionId: targetId},
			models.SalesPromotionChannelFieldPromotionId, targetField); err != nil {
			return nil, err
		}
	}
	if err := copyOwnedRows(ctx, models.SalesPromotionRuleSchemaName, sourceId, targetId,
		models.SalesPromotionSelectorFieldRuleId, ruleCopiedFields,
		models.SalesPromotionRuleProductSchemaName, models.SalesPromotionRuleCategorySchemaName); err != nil {
		return nil, err
	}
	return nil, copyOwnedRows(ctx, models.SalesPromotionRewardSchemaName, sourceId, targetId,
		models.SalesPromotionSelectorFieldRewardId, rewardCopiedFields,
		models.SalesPromotionRewardProductSchemaName, models.SalesPromotionRewardCategorySchemaName)
}

var ruleCopiedFields = []string{
	basemodel.FieldOrgId, models.SalesPromotionRuleFieldPhase, models.SalesPromotionRuleFieldName,
	models.SalesPromotionRuleFieldMinQuantity, models.SalesPromotionRuleFieldQualificationUomId,
	models.SalesPromotionRuleFieldMinAmount, models.SalesPromotionRuleFieldAmountTaxBasis,
}

var rewardCopiedFields = []string{
	basemodel.FieldOrgId, models.SalesPromotionRewardFieldRewardType,
	models.SalesPromotionRewardFieldOrderDescription, models.SalesPromotionRewardFieldDiscountMethod,
	models.SalesPromotionRewardFieldDiscountValue, models.SalesPromotionRewardFieldDiscountScope,
	models.SalesPromotionRewardFieldMaxDiscountAmount, models.SalesPromotionRewardFieldGiftProductVariantId,
	models.SalesPromotionRewardFieldGiftQuantity, models.SalesPromotionRewardFieldGiftUomId,
}

var selectorCopiedFields = []string{
	basemodel.FieldOrgId, models.SalesPromotionSelectorFieldTemplateId,
	models.SalesPromotionSelectorFieldVariantId, models.SalesPromotionSelectorFieldCategoryId,
}

// copyOwnedRows copies every rule (or reward) of the source promotion through the owner's domain
// service, then the selectors of each copy.
func copyOwnedRows(
	ctx corectx.Context, schemaName, sourceId, targetId, ownerField string, fields []string,
	selectorSchemas ...string,
) error {
	rows, err := searchBy(ctx, schemaName, models.SalesPromotionRuleFieldPromotionId, sourceId)
	if err != nil {
		return err
	}
	domSvc, err := domainServiceFor(schemaName)
	if err != nil {
		return err
	}
	for _, row := range rows {
		cmd := pickFields(row, fields)
		cmd[models.SalesPromotionRuleFieldPromotionId] = targetId
		created, err := domSvc.Create(ctx, cmd)
		if vErrs, err := clientErrorsOf(created, err); err != nil || vErrs != nil {
			return errors.Errorf("duplicate %s: %v", schemaName, firstError(vErrs, err))
		}
		copyId := stringOf(created.Data, basemodel.FieldId)
		for _, selectorSchema := range selectorSchemas {
			if err := copyRows(ctx, selectorSchema, ownerField, stringOf(row, basemodel.FieldId),
				dmodel.DynamicFields{ownerField: copyId}, selectorCopiedFields...); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyRows inserts one copy of every row whose field holds value, keeping the named fields and
// applying overrides.
func copyRows(
	ctx corectx.Context, schemaName, field, value string, overrides dmodel.DynamicFields, keep ...string,
) error {
	rows, err := searchBy(ctx, schemaName, field, value)
	if err != nil {
		return err
	}
	repo, err := repoFor(schemaName)
	if err != nil {
		return err
	}
	for _, row := range rows {
		copied := pickFields(row, keep)
		for key, override := range overrides {
			copied[key] = override
		}
		newId, err := model.NewId()
		if err != nil {
			return errors.Wrap(err, "copyRows")
		}
		copied[basemodel.FieldId] = string(*newId)
		if _, err := repo.Insert(ctx, copied); err != nil {
			return errors.Wrapf(err, "copy %s", schemaName)
		}
	}
	return nil
}

func firstError(vErrs *ft.ClientErrors, err error) any {
	if err != nil {
		return err
	}
	if vErrs != nil {
		return vErrs.ToError()
	}
	return nil
}
