package dynamicengines

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

const (
	// TemplateAttributeConfigurationFn reads the enabled lines with their selected values and the
	// active exclusions, in the shape a configuration save sends.
	TemplateAttributeConfigurationFn = "inventory.template_attribute_configuration"
)

// newTemplateComputedFunctions builds the "function"-kind computed fields of product_template.
func newTemplateComputedFunctions() map[string]composable.ComputedFieldFn {
	configuration := func(ctx corectx.Context, req composable.ComputeFnRequest) ([]any, error) {
		views, err := services.ReadTemplateConfigurations(ctx, pageTemplateIds(req.Models))
		if err != nil {
			return nil, errors.Wrapf(err, "computed %q", TemplateAttributeConfigurationFn)
		}
		out := make([]any, len(req.Models))
		for i, row := range req.Models {
			if view, found := views[readString(row[basemodel.FieldId])]; found {
				out[i] = toJsonMap(view)
			}
		}
		return out, nil
	}

	return map[string]composable.ComputedFieldFn{
		TemplateAttributeConfigurationFn: configuration,
	}
}

// pageTemplateIds lists the ids of a page of template rows.
func pageTemplateIds(rows []dmodel.DynamicFields) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if id := readString(row[basemodel.FieldId]); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// toJsonMap renders a view as the map a jsonmap field carries.
func toJsonMap(view any) map[string]any {
	encoded, err := json.Marshal(view)
	if err != nil {
		return nil
	}
	out := map[string]any{}
	if json.Unmarshal(encoded, &out) != nil {
		return nil
	}
	return out
}
