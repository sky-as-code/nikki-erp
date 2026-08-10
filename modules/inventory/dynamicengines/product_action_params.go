package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// Shared plumbing for the Products custom actions: reaching the derived service, and turning an
// untyped request body into a typed query.

// asProductService recovers the derived Products service the module installed on the engine.
//
// A failed assertion means SetResourceService was never called for this engine, which is a
// wiring mistake rather than anything a caller did — so it surfaces as an error (500), not as
// client errors (400).
func asProductService(input drif.ProcessInput) (itProduct.ProductService, error) {
	productSvc, ok := input.ResourceService.(itProduct.ProductService)
	if !ok {
		return nil, errors.New(
			"the product engine is not serving the derived Products service; SetResourceService was not called")
	}
	return productSvc, nil
}

// buildResolveSelectionQuery validates and converts the resolve_selection body.
//
// Without a ParamSchema the params arrive untyped, and a body whose selections are malformed
// would otherwise decode to an empty list — resolving to the empty combination, which is a real
// and different variant, answered 200. Every shape problem below is therefore a client error.
func buildResolveSelectionQuery(
	params dmodel.DynamicFields,
) (itProduct.ResolveProductSelectionQuery, *ft.ClientErrors) {
	vErrs := &ft.ClientErrors{}
	query := itProduct.ResolveProductSelectionQuery{}

	templateId, ok := params["template_id"].(string)
	if !ok || templateId == "" {
		vErrs.Append(*ft.NewBusinessViolation("template_id",
			"product_template.template_id_required",
			"a product template must be identified"))
		return query, vErrs
	}
	query.TemplateId = templateId

	if materialize, ok := params["materialize_if_missing"].(bool); ok {
		query.MaterializeIfMissing = materialize
	}

	raw, present := params["selections"]
	if !present {
		// A template with no variant-generating attributes resolves on an empty selection, so
		// absence is legitimate rather than an error.
		return query, vErrs
	}

	items, ok := raw.([]any)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation("selections",
			"product_template.selections_malformed",
			"selections must be a list of attribute-value choices"))
		return query, vErrs
	}

	for _, item := range items {
		selection, itemErr := toAttributeSelection(item)
		if itemErr != nil {
			vErrs.Append(*itemErr)
			continue
		}
		query.Selections = append(query.Selections, selection)
	}
	return query, vErrs
}

func toAttributeSelection(item any) (itProduct.AttributeSelection, *ft.ClientErrorItem) {
	fields, ok := item.(map[string]any)
	if !ok {
		return itProduct.AttributeSelection{}, ft.NewBusinessViolation("selections",
			"product_template.selection_malformed",
			"each selection must be an object with attribute_id and value_id")
	}

	input := itProduct.AttributeSelectionInput{}
	input.AttributeId, _ = fields["attribute_id"].(string)
	input.ValueId, _ = fields["value_id"].(string)
	input.Mode, _ = fields["mode"].(string)

	if input.AttributeId == "" || input.ValueId == "" {
		return itProduct.AttributeSelection{}, ft.NewBusinessViolation("selections",
			"product_template.selection_incomplete",
			"each selection must carry both attribute_id and value_id")
	}
	return input.ToSelection(), nil
}
