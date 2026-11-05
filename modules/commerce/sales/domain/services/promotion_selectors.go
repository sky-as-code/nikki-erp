package services

import (
	"strconv"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
)

var hundredPercent = decimal.NewFromInt(100)

// The rule and reward domain services, and the selectors they manage (BR §8.6). A selector row is
// written only here: POST {rule|reward}/:id/manage_products takes
// {add: [{product_template_id?, product_variant_id?}], remove: [row ids]}, and manage_categories
// takes {add: [category ids], remove: [row ids]}.

var (
	_ itPromotion.SalesPromotionRuleDomainService   = (*SalesPromotionRuleDomainServiceImpl)(nil)
	_ itPromotion.SalesPromotionRewardDomainService = (*SalesPromotionRewardDomainServiceImpl)(nil)
)

type SalesPromotionRuleDomainServiceImpl struct {
	promotionChildService
}

func NewSalesPromotionRuleDomainService(base dynamicresource.CrudDomainService) *SalesPromotionRuleDomainServiceImpl {
	return &SalesPromotionRuleDomainServiceImpl{promotionChildService{
		CrudDomainService: base, schemaName: models.SalesPromotionRuleSchemaName, check: checkRule, clear: clearRule,
	}}
}

func (this *SalesPromotionRuleDomainServiceImpl) ManageProducts(
	ctx corectx.Context, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	return manageSelectors(ctx, ruleProductSelectors, cmd)
}

func (this *SalesPromotionRuleDomainServiceImpl) ManageCategories(
	ctx corectx.Context, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	return manageSelectors(ctx, ruleCategorySelectors, cmd)
}

type SalesPromotionRewardDomainServiceImpl struct {
	promotionChildService
}

func NewSalesPromotionRewardDomainService(base dynamicresource.CrudDomainService) *SalesPromotionRewardDomainServiceImpl {
	return &SalesPromotionRewardDomainServiceImpl{promotionChildService{
		CrudDomainService: base, schemaName: models.SalesPromotionRewardSchemaName,
		check: checkReward, clear: clearReward,
	}}
}

func (this *SalesPromotionRewardDomainServiceImpl) ManageProducts(
	ctx corectx.Context, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	return manageSelectors(ctx, rewardProductSelectors, cmd)
}

func (this *SalesPromotionRewardDomainServiceImpl) ManageCategories(
	ctx corectx.Context, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	return manageSelectors(ctx, rewardCategorySelectors, cmd)
}

// selectorKind names one selector table and the owner it hangs off.
type selectorKind struct {
	ownerSchema    string
	selectorSchema string
	ownerField     string
	byCategory     bool
}

var (
	ruleProductSelectors = selectorKind{models.SalesPromotionRuleSchemaName,
		models.SalesPromotionRuleProductSchemaName, models.SalesPromotionSelectorFieldRuleId, false}
	ruleCategorySelectors = selectorKind{models.SalesPromotionRuleSchemaName,
		models.SalesPromotionRuleCategorySchemaName, models.SalesPromotionSelectorFieldRuleId, true}
	rewardProductSelectors = selectorKind{models.SalesPromotionRewardSchemaName,
		models.SalesPromotionRewardProductSchemaName, models.SalesPromotionSelectorFieldRewardId, false}
	rewardCategorySelectors = selectorKind{models.SalesPromotionRewardSchemaName,
		models.SalesPromotionRewardCategorySchemaName, models.SalesPromotionSelectorFieldRewardId, true}
)

// manageSelectors adds and removes selector rows of one owner in one transaction, refused while the
// promotion is active, and bumps the promotion's config_version.
func manageSelectors(
	ctx corectx.Context, kind selectorKind, cmd itPromotion.ManageCommand,
) (*itPromotion.ManageResult, error) {
	repo, err := repoFor(kind.selectorSchema)
	if err != nil {
		return nil, err
	}
	changed := 0
	refused, err := dynamicresource.InRefusableTransaction(ctx, repo,
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			owner, err := loadRecord(tranxCtx, kind.ownerSchema, basemodel.FieldId, stringOf(cmd, basemodel.FieldId))
			if err != nil || owner == nil {
				return promotionChildMissing(kind.ownerSchema, stringOf(cmd, basemodel.FieldId), err)
			}
			promotionId := stringOf(owner, models.SalesPromotionRuleFieldPromotionId)
			if refused, err := bumpPromotionConfig(tranxCtx, promotionId); err != nil || refused != nil {
				return refused, err
			}
			added, refused, err := addSelectors(tranxCtx, kind, owner, cmd[dynamicresource.FieldM2mAdd])
			if err != nil || refused != nil {
				return refused, err
			}
			removed, refused, err := removeSelectors(tranxCtx, kind, owner, cmd[dynamicresource.FieldM2mRemove])
			changed = added + removed
			return refused, err
		})
	if err != nil || refused != nil {
		return mutateRefusal(refused, err)
	}
	return &itPromotion.ManageResult{Data: dyn.MutateResultData{AffectedCount: changed}, HasData: true}, nil
}

func promotionChildMissing(schemaName, id string, err error) (*ft.ClientErrors, error) {
	if err != nil {
		return nil, err
	}
	return &notFoundResult(schemaName, id).ClientErrors, nil
}

func addSelectors(
	ctx corectx.Context, kind selectorKind, owner dmodel.DynamicFields, raw any,
) (int, *ft.ClientErrors, error) {
	items, _ := raw.([]any)
	domSvc, err := domainServiceFor(kind.selectorSchema)
	if err != nil {
		return 0, nil, err
	}
	for index, item := range items {
		row, vErrs := selectorRowOf(kind, item, index)
		if vErrs != nil {
			return 0, vErrs, nil
		}
		row[basemodel.FieldOrgId] = stringOf(owner, basemodel.FieldOrgId)
		row[kind.ownerField] = stringOf(owner, basemodel.FieldId)
		created, err := domSvc.Create(ctx, row)
		if vErrs, err := clientErrorsOf(created, err); err != nil || vErrs != nil {
			return 0, vErrs, err
		}
	}
	return len(items), nil, nil
}

// selectorRowOf reads one add entry: a category id, or an object naming exactly one of a template
// and a variant.
func selectorRowOf(kind selectorKind, item any, index int) (dmodel.DynamicFields, *ft.ClientErrors) {
	field := dynamicresource.FieldM2mAdd + "[" + strconv.Itoa(index) + "]"
	if kind.byCategory {
		categoryId, ok := item.(string)
		if !ok || categoryId == "" {
			return nil, violationsOf(field, "promotion.selector_invalid", "a category id is expected")
		}
		return dmodel.DynamicFields{models.SalesPromotionSelectorFieldCategoryId: categoryId}, nil
	}
	entry, _ := item.(map[string]any)
	templateId := stringOf(entry, models.SalesPromotionSelectorFieldTemplateId)
	variantId := stringOf(entry, models.SalesPromotionSelectorFieldVariantId)
	if (templateId == "") == (variantId == "") {
		return nil, violationsOf(field, "promotion.selector_invalid",
			"name exactly one of product_template_id and product_variant_id")
	}
	if templateId != "" {
		return dmodel.DynamicFields{models.SalesPromotionSelectorFieldTemplateId: templateId}, nil
	}
	return dmodel.DynamicFields{models.SalesPromotionSelectorFieldVariantId: variantId}, nil
}

func removeSelectors(
	ctx corectx.Context, kind selectorKind, owner dmodel.DynamicFields, raw any,
) (int, *ft.ClientErrors, error) {
	items, _ := raw.([]any)
	repo, err := repoFor(kind.selectorSchema)
	if err != nil {
		return 0, nil, err
	}
	for index, item := range items {
		rowId, _ := item.(string)
		row, err := loadRecord(ctx, kind.selectorSchema, basemodel.FieldId, rowId)
		if err != nil {
			return 0, nil, err
		}
		if row == nil || stringOf(row, kind.ownerField) != stringOf(owner, basemodel.FieldId) {
			field := dynamicresource.FieldM2mRemove + "[" + strconv.Itoa(index) + "]"
			return 0, violationsOf(field, "promotion.selector_invalid", "no such selector on this record"), nil
		}
		if _, err := repo.DeleteOne(ctx, dmodel.DynamicFields{basemodel.FieldId: rowId}); err != nil {
			return 0, nil, err
		}
	}
	return len(items), nil, nil
}
