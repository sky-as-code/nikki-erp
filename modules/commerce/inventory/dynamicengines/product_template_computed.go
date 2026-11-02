package dynamicengines

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

const (
	// TemplateAttributeConfigurationFn reads the enabled lines with their selected values and the
	// active exclusions, in the shape a configuration save sends.
	TemplateAttributeConfigurationFn = "inventory.template_attribute_configuration"

	// TemplateWeightUomSymbolFn and TemplateDimensionUomSymbolFn give the symbols of the units the
	// template's weight and its length, width and height are expressed in, from the org's inventory
	// settings. The Variant* pair answers the same for a variant's own measurements.
	TemplateWeightUomSymbolFn    = "inventory.template_weight_uom_symbol"
	TemplateDimensionUomSymbolFn = "inventory.template_dimension_uom_symbol"
	VariantWeightUomSymbolFn     = "inventory.variant_weight_uom_symbol"
	VariantDimensionUomSymbolFn  = "inventory.variant_dimension_uom_symbol"
)

// newTemplateComputedFunctions builds the "function"-kind computed fields of product_template.
func newTemplateComputedFunctions(units *services.OrgUnitResolver) map[string]dynamicresource.ComputedFieldFn {
	configuration := func(ctx corectx.Context, req dynamicresource.ComputeFnRequest) ([]any, error) {
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

	return map[string]dynamicresource.ComputedFieldFn{
		TemplateAttributeConfigurationFn: configuration,
		TemplateWeightUomSymbolFn: unitSymbolFn(units, models.ProductTemplateSchemaName,
			invsettings.OrgSettingDefaultWeightUomId),
		TemplateDimensionUomSymbolFn: unitSymbolFn(units, models.ProductTemplateSchemaName,
			invsettings.OrgSettingDefaultLengthUomId),
	}
}

func mergeComputedFunctions(
	sets ...map[string]dynamicresource.ComputedFieldFn,
) map[string]dynamicresource.ComputedFieldFn {
	out := map[string]dynamicresource.ComputedFieldFn{}
	for _, set := range sets {
		maps.Copy(out, set)
	}
	return out
}

// pageUnitsCtxKey memoizes a page's resolved units on the request, so the weight and the dimension
// field share one settings read and one read per unit rather than paying for both.
type pageUnitsCtxKey struct{}

// newVariantUnitFunctions gives a variant's measurements the same unit symbols as its template's.
func newVariantUnitFunctions(units *services.OrgUnitResolver) map[string]dynamicresource.ComputedFieldFn {
	return map[string]dynamicresource.ComputedFieldFn{
		VariantWeightUomSymbolFn: unitSymbolFn(units, models.ProductVariantSchemaName,
			invsettings.OrgSettingDefaultWeightUomId),
		VariantDimensionUomSymbolFn: unitSymbolFn(units, models.ProductVariantSchemaName,
			invsettings.OrgSettingDefaultLengthUomId),
	}
}

// unitSymbolFn gives the symbol of one setting's unit for every row of the page, from the units of
// the rows' organizations: a fixed number of reads per page, never one per row.
func unitSymbolFn(units *services.OrgUnitResolver, schemaName string, setting string) dynamicresource.ComputedFieldFn {
	return func(ctx corectx.Context, req dynamicresource.ComputeFnRequest) ([]any, error) {
		orgOf, err := rowOrgIds(ctx, schemaName, req.Models)
		if err != nil {
			return nil, err
		}
		resolved, err := resolvePageUnits(ctx, units, orgOf)
		if err != nil {
			return nil, errors.Wrapf(err, "computed %q", setting)
		}
		out := make([]any, len(req.Models))
		for i, row := range req.Models {
			if uom := resolved[orgOf[readString(row[basemodel.FieldId])]][setting]; uom != nil {
				out[i] = uom.Symbol
			} else if orgId := readString(row[basemodel.FieldOrgId]); orgId != "" {
				if uom := resolved[orgId][setting]; uom != nil {
					out[i] = uom.Symbol
				}
			}
		}
		return out, nil
	}
}

func resolvePageUnits(
	ctx corectx.Context, units *services.OrgUnitResolver, orgOf map[string]string,
) (map[string]services.OrgUnits, error) {
	cached := map[string]services.OrgUnits{}
	if ctx != nil {
		if found, ok := ctx.Value(pageUnitsCtxKey{}).(map[string]services.OrgUnits); ok {
			cached = found
		}
	}
	missing := []string{}
	for _, orgId := range orgOf {
		if _, ok := cached[orgId]; !ok && orgId != "" && !slices.Contains(missing, orgId) {
			missing = append(missing, orgId)
		}
	}
	if len(missing) == 0 {
		return cached, nil
	}
	resolved, err := units.ResolveUnits(ctx, missing)
	if err != nil {
		return nil, err
	}
	for orgId, orgUnits := range resolved {
		cached[orgId] = orgUnits
	}
	if ctx != nil {
		ctx.WithValue(pageUnitsCtxKey{}, cached)
	}
	return cached, nil
}

// rowOrgIds maps each row's id to its org. A row projected without org_id (a `fields=` list naming
// only the unit) has it read back in one search for the page.
func rowOrgIds(ctx corectx.Context, schemaName string, rows []dmodel.DynamicFields) (map[string]string, error) {
	orgOf := map[string]string{}
	missing := []string{}
	for i, row := range rows {
		id, orgId := readString(row[basemodel.FieldId]), readString(row[basemodel.FieldOrgId])
		switch {
		case orgId != "":
			orgOf[firstNonEmptyKey(id, i)] = orgId
		case id != "":
			missing = append(missing, id)
		}
	}
	found, err := searchByIds(ctx, dynamicresource.LookupSourceOnion, schemaName,
		basemodel.FieldId, missing, []string{basemodel.FieldId, basemodel.FieldOrgId}, len(missing))
	if err != nil {
		return nil, err
	}
	for _, row := range found {
		orgOf[readString(row[basemodel.FieldId])] = readString(row[basemodel.FieldOrgId])
	}
	return orgOf, nil
}

// firstNonEmptyKey keys an unsaved row (a meta/compute call has no id yet) by its position.
func firstNonEmptyKey(id string, index int) string {
	if id != "" {
		return id
	}
	return fmt.Sprintf("#%d", index)
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
