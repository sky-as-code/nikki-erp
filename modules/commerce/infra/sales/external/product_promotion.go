package external

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	invModels "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"

	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
)

const variantTemplateUomField = "product_template_uom_id"

// productPromotionAdapter answers the promotion engine's product facts from three published
// readers: the variant search (template, base unit), the pricing basis (category ancestry, walked
// in one batch by Inventory) and Essential's unit lookup (unit category).
type productPromotionAdapter struct {
	variants itProduct.ProductVariantDomainService
	basis    itProduct.ProductPricingBasisService
	uoms     itUom.UomConversionAppService
}

func (this *productPromotionAdapter) ResolveForPromotion(
	ctx corectx.Context, variantIds []string,
) (map[string]itExt.PromotionProductFacts, error) {
	facts := map[string]itExt.PromotionProductFacts{}
	rows, err := this.searchVariants(ctx, invModels.ProductVariantFieldId, variantIds)
	if err != nil || len(rows) == 0 {
		return facts, err
	}
	categories := this.categoryPaths(ctx, variantIds)
	uomCategories := map[string]string{}
	for _, row := range rows {
		uomId := stringField(row, variantTemplateUomField)
		if _, known := uomCategories[uomId]; !known {
			uomCategories[uomId] = this.uomCategoryOf(ctx, uomId)
		}
		variantId := stringField(row, invModels.ProductVariantFieldId)
		facts[variantId] = itExt.PromotionProductFacts{
			ProductVariantId:  variantId,
			ProductTemplateId: stringField(row, invModels.ProductVariantFieldProductTemplateId),
			CategoryIds:       categories[variantId],
			BaseUomId:         uomId,
			BaseUomCategoryId: uomCategories[uomId],
			Code:              stringField(row, invModels.ProductVariantFieldSku),
			Name:              displayNameOf(row),
		}
	}
	return facts, nil
}

func (this *productPromotionAdapter) ResolveUomCategories(
	ctx corectx.Context, templateIds []string, variantIds []string,
) (map[string]string, error) {
	answer := map[string]string{}
	byTemplate, err := this.searchVariants(ctx, invModels.ProductVariantFieldProductTemplateId, templateIds)
	if err != nil {
		return nil, err
	}
	byVariant, err := this.searchVariants(ctx, invModels.ProductVariantFieldId, variantIds)
	if err != nil {
		return nil, err
	}
	uomCategories := map[string]string{}
	record := func(key, uomId string) {
		if _, known := uomCategories[uomId]; !known {
			uomCategories[uomId] = this.uomCategoryOf(ctx, uomId)
		}
		answer[key] = uomCategories[uomId]
	}
	for _, row := range byTemplate {
		record(stringField(row, invModels.ProductVariantFieldProductTemplateId), stringField(row, variantTemplateUomField))
	}
	for _, row := range byVariant {
		record(stringField(row, invModels.ProductVariantFieldId), stringField(row, variantTemplateUomField))
	}
	return answer, nil
}

func (this *productPromotionAdapter) searchVariants(
	ctx corectx.Context, field string, ids []string,
) ([]dmodel.DynamicFields, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(field, dmodel.In, values...))
	found, err := this.variants.SearchProductVariants(ctx, dyn.SearchQuery{
		Graph: graph, Page: 0, Size: model.MODEL_RULE_PAGE_MAX_SIZE,
		Fields: []string{
			invModels.ProductVariantFieldId, invModels.ProductVariantFieldProductTemplateId,
			invModels.ProductVariantFieldSku, invModels.ProductVariantFieldDisplayName, variantTemplateUomField,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "searching product variants for promotion")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	rows := make([]dmodel.DynamicFields, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		rows = append(rows, item.GetFieldData())
	}
	return rows, nil
}

// categoryPaths is best effort: a variant whose ancestry cannot be read simply matches no category
// selector, which errs towards granting no benefit.
func (this *productPromotionAdapter) categoryPaths(ctx corectx.Context, variantIds []string) map[string][]string {
	paths := map[string][]string{}
	if this.basis == nil {
		return paths
	}
	result, err := this.basis.GetPricingBasis(ctx, itProduct.GetPricingBasisQuery{ProductVariantIds: variantIds})
	if err != nil || result == nil || !result.HasData {
		return paths
	}
	for variantId, basis := range result.Data.Bases {
		paths[variantId] = basis.CategoryPath
	}
	return paths
}

func (this *productPromotionAdapter) uomCategoryOf(ctx corectx.Context, uomId string) string {
	if uomId == "" || this.uoms == nil {
		return ""
	}
	found, err := this.uoms.GetUom(ctx, itUom.GetUomQuery{Id: model.Id(uomId)})
	if err != nil || found == nil || !found.HasData {
		return ""
	}
	return string(found.Data.CategoryId)
}

func stringField(row dmodel.DynamicFields, field string) string {
	if value := row.GetString(field); value != nil {
		return *value
	}
	if value := row.GetModelId(field); value != nil {
		return string(*value)
	}
	return ""
}

func displayNameOf(row dmodel.DynamicFields) string {
	name := row.GetLangJson(invModels.ProductVariantFieldDisplayName)
	if name == nil {
		return ""
	}
	return name.TextIn(model.DefaultLanguageCode)
}
