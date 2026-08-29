package services

import (
	"sort"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// BuildEffectiveProduct flattens a variant and its template into the view consumers read.
// valueLabels are the display labels of the variant's attribute values, in combination-key order.
func BuildEffectiveProduct(
	template *models.ProductTemplate,
	variant *models.ProductVariant,
	valueLabels []string,
) itProduct.EffectiveProduct {
	effective := itProduct.EffectiveProduct{
		VariantId:          derefString(variant.GetId()),
		TemplateId:         derefString(template.GetId()),
		Name:               template.GetName(),
		ProductTypeId:      derefString(template.GetProductTypeId()),
		CategoryId:         derefString(template.GetCategoryId()),
		BrandId:            derefString(template.GetBrandId()),
		SaleOk:             derefBool(template.GetSaleOk()),
		PurchaseOk:         derefBool(template.GetPurchaseOk()),
		Sku:                derefString(variant.GetSku()),
		Barcode:            derefString(variant.GetPrimaryBarcode()),
		IsTemplateArchived: derefBool(template.IsArchived()),
		IsVariantArchived:  derefBool(variant.IsArchived()),
	}

	if status := template.GetStatus(); status != nil {
		effective.TemplateStatus = status.String()
	}
	if status := variant.GetStatus(); status != nil {
		effective.VariantStatus = status.String()
	}

	effective.DisplayName = BuildDisplayName(template.GetName(), valueLabels)

	// A nil variant value means "inherit the template's", so these are not simply copied across.
	effective.ImageId = firstNonEmpty(
		derefString(variant.GetVariantImageId()), derefString(template.GetDefaultImageId()))
	effective.Weight = firstDecimal(variant.GetWeight(), template.GetDefaultWeight())
	effective.Length = firstDecimal(variant.GetLength(), template.GetDefaultLength())
	effective.Width = firstDecimal(variant.GetWidth(), template.GetDefaultWidth())
	effective.Height = firstDecimal(variant.GetHeight(), template.GetDefaultHeight())

	return effective
}

// BuildDisplayName composes a variant's user-facing name from its template's name and its
// attribute values, e.g. "Classic T-Shirt / Black / M". A variant with no values renders as the
// template name alone, which is the single-variant case.
func BuildDisplayName(templateName *model.LangJson, valueLabels []string) string {
	parts := make([]string, 0, len(valueLabels)+1)
	if base := langJsonToString(templateName); base != "" {
		parts = append(parts, base)
	}
	for _, label := range valueLabels {
		if label != "" {
			parts = append(parts, label)
		}
	}
	return joinNonEmpty(parts, " / ")
}

// ResolveProductSelection turns a template plus the attribute values a user picked into the
// combination key identifying the concrete variant. A transaction line always references a variant,
// never a template, so consumers call this rather than deciding for themselves.
func ResolveProductSelection(selections []itProduct.AttributeSelection) string {
	return BuildCombinationKey(selections)
}

// VariantValueIds lists the attribute-value ids a variant's combination holds, in key order.
func VariantValueIds(variant *models.ProductVariant) []string {
	key := variant.GetCombinationKey()
	if key == nil {
		return nil
	}
	selections := ParseCombinationKey(*key)
	valueIds := make([]string, 0, len(selections))
	for _, selection := range selections {
		valueIds = append(valueIds, selection.ValueId)
	}
	return valueIds
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefBool(v *bool) bool {
	return v != nil && *v
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstDecimal(values ...*decimal.Decimal) *decimal.Decimal {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func joinNonEmpty(parts []string, separator string) string {
	result := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		if result != "" {
			result += separator
		}
		result += part
	}
	return result
}

// langJsonToString picks a single string out of a localized value. The caller's language is unknown
// here, so it prefers en-US and otherwise the lowest-sorting locale present; iterating the map
// directly would render the same product under a different name on each call.
func langJsonToString(value *model.LangJson) string {
	if value == nil || len(*value) == 0 {
		return ""
	}
	if text, ok := (*value)[model.LanguageCodeEnUs]; ok && text != "" {
		return text
	}

	codes := make([]string, 0, len(*value))
	for code := range *value {
		codes = append(codes, string(code))
	}
	sort.Strings(codes)

	for _, code := range codes {
		if text := (*value)[model.LanguageCode(code)]; text != "" {
			return text
		}
	}
	return ""
}
