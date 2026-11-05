package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
)

// FunctionRuleHasMixedUom names the Go half of sales_promotion_rule.has_mixed_uom.
const FunctionRuleHasMixedUom = "sales.promotion_rule_has_mixed_uom"

// PromotionUomCategoryResolver answers the UoM category of each product a selector names, keyed by
// the template or variant id. It is Inventory's answer, reached through the product port; nil
// until that port is wired, in which case no rule reports mixed units.
type PromotionUomCategoryResolver func(
	ctx corectx.Context, templateIds []string, variantIds []string,
) (map[string]string, error)

var promotionUomCategories PromotionUomCategoryResolver

// SetPromotionUomCategoryResolver installs the resolver, from the module's OnAppStarted.
func SetPromotionUomCategoryResolver(resolver PromotionUomCategoryResolver) {
	promotionUomCategories = resolver
}

// RuleHasMixedUomFn computes has_mixed_uom for a page of rules: true when the products a rule's
// selectors name fall in more than one UoM category (BR-C07). One selector read per rule.
func RuleHasMixedUomFn(ctx corectx.Context, req dynamicresource.ComputeFnRequest) ([]any, error) {
	values := make([]any, len(req.Models))
	for index, rule := range req.Models {
		values[index] = false
		if promotionUomCategories == nil {
			continue
		}
		mixed, err := ruleHasMixedUom(ctx, stringOf(rule, basemodel.FieldId))
		if err != nil {
			return nil, err
		}
		values[index] = mixed
	}
	return values, nil
}

func ruleHasMixedUom(ctx corectx.Context, ruleId string) (bool, error) {
	if ruleId == "" {
		return false, nil
	}
	selectors, err := searchBy(ctx, models.SalesPromotionRuleProductSchemaName,
		models.SalesPromotionSelectorFieldRuleId, ruleId)
	if err != nil || len(selectors) < 2 {
		return false, err
	}
	templateIds, variantIds := make([]string, 0), make([]string, 0)
	for _, selector := range selectors {
		if id := stringOf(selector, models.SalesPromotionSelectorFieldTemplateId); id != "" {
			templateIds = append(templateIds, id)
		}
		if id := stringOf(selector, models.SalesPromotionSelectorFieldVariantId); id != "" {
			variantIds = append(variantIds, id)
		}
	}
	categories, err := promotionUomCategories(ctx, templateIds, variantIds)
	if err != nil {
		return false, err
	}
	distinct := map[string]bool{}
	for _, category := range categories {
		if category != "" {
			distinct[category] = true
		}
	}
	return len(distinct) > 1, nil
}
