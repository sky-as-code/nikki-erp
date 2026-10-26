package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
)

// Request fields of resolve_selection. template_id is the name the endpoint was first published
// with and is still accepted.
const (
	paramResolveTemplateId       = "product_template_id"
	paramResolveLegacyTemplateId = "template_id"
	paramResolveSelections       = "selections"
	paramResolveExpectedEtag     = "expected_etag"
	paramResolveMaterialize      = "materialize_if_missing"
)

// buildResolveCompleteQuery validates and converts the resolve_selection body. The params arrive
// untyped, so malformed selections would otherwise decode to an empty list and resolve to the
// empty combination, a real and different variant, answered 200. Every shape problem is therefore
// a client error, and so is a second value for the same attribute.
func buildResolveCompleteQuery(
	params dmodel.DynamicFields,
) (itProduct.ResolveCompleteVariantQuery, *ft.ClientErrors) {
	vErrs := &ft.ClientErrors{}
	query := itProduct.ResolveCompleteVariantQuery{Selections: map[string]string{}}

	query.TemplateId = readResolveTemplateId(params)
	if query.TemplateId == "" {
		vErrs.Append(*ft.NewBusinessViolation(paramResolveTemplateId,
			ft.ErrorKey("product_template.template_id_required", "inventory"),
			"a product template must be identified"))
		return query, vErrs
	}
	query.MaterializeIfMissing = readResolveMaterialize(params)
	if etag, ok := params[paramResolveExpectedEtag].(string); ok && etag != "" {
		query.ExpectedEtag = &etag
	}

	raw, present := params[paramResolveSelections]
	if !present || raw == nil {
		// A template with no variant-generating attributes resolves on an empty selection, so
		// absence is legitimate rather than an error.
		return query, vErrs
	}
	items, ok := raw.([]any)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(paramResolveSelections,
			ft.ErrorKey("product_template.selections_malformed", "inventory"),
			"selections must be a list of attribute-value choices"))
		return query, vErrs
	}
	for _, item := range items {
		if itemErr := addResolveSelection(query.Selections, item); itemErr != nil {
			vErrs.Append(*itemErr)
		}
	}
	return query, vErrs
}

func readResolveTemplateId(params dmodel.DynamicFields) string {
	if templateId, ok := params[paramResolveTemplateId].(string); ok && templateId != "" {
		return templateId
	}
	templateId, _ := params[paramResolveLegacyTemplateId].(string)
	return templateId
}

// readResolveMaterialize reads the write flag. False is a pure lookup that never writes.
func readResolveMaterialize(params dmodel.DynamicFields) bool {
	materialize, _ := params[paramResolveMaterialize].(bool)
	return materialize
}

func addResolveSelection(selections map[string]string, item any) *ft.ClientErrorItem {
	fields, ok := item.(map[string]any)
	if !ok {
		return ft.NewBusinessViolation(paramResolveSelections,
			ft.ErrorKey("product_template.selection_malformed", "inventory"),
			"each selection must be an object with attribute_id and value_id")
	}
	attributeId, _ := fields["attribute_id"].(string)
	valueId, _ := fields["value_id"].(string)
	if attributeId == "" || valueId == "" {
		return ft.NewBusinessViolation(paramResolveSelections,
			ft.ErrorKey("product_template.selection_incomplete", "inventory"),
			"each selection must carry both attribute_id and value_id")
	}
	if _, taken := selections[attributeId]; taken {
		return ft.NewBusinessViolation(paramResolveSelections,
			ft.ErrorKey("product_template.selection_duplicate_attribute", "inventory"),
			"only one value may be chosen for attribute {{attribute_id}}",
			map[string]any{"attribute_id": attributeId})
	}
	selections[attributeId] = valueId
	return nil
}
