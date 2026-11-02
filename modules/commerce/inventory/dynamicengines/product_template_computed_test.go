package dynamicengines

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type countingSettings struct{ calls int }

func (this *countingSettings) GetEffectiveSettings(
	_ corectx.Context, _ itExt.GetEffectiveSettingsQuery,
) (*itExt.GetEffectiveSettingsResult, error) {
	this.calls++
	return &itExt.GetEffectiveSettingsResult{HasData: true}, nil
}

type symbolUoms struct{ lookups int }

func (this *symbolUoms) GetUom(_ corectx.Context, _ itExt.GetUomQuery) (*itExt.GetUomResult, error) {
	return &itExt.GetUomResult{}, nil
}

func (this *symbolUoms) FindUomBySymbol(
	_ corectx.Context, query itExt.FindUomBySymbolQuery,
) (*itExt.GetUomResult, error) {
	this.lookups++
	return &itExt.GetUomResult{HasData: true, Data: itExt.GetUomResultData{
		Id: model.Id(query.Symbol), Symbol: query.Symbol, OrgId: query.OrgId,
		Name: model.LangJson{"en-US": query.Symbol + "@" + string(query.OrgId)},
	}}, nil
}

// Both unit fields of a page share one settings read and one lookup per organization and unit,
// whatever the number of rows.
func TestTemplateUnitSymbolsAreResolvedPerPage(t *testing.T) {
	settings, uoms := &countingSettings{}, &symbolUoms{}
	fns := newTemplateComputedFunctions(services.NewOrgUnitResolver(settings, uoms))
	ctx := corectx.NewRequestContext(context.Background())
	rows := []dmodel.DynamicFields{
		{basemodel.FieldId: "t1", basemodel.FieldOrgId: "org-1"},
		{basemodel.FieldId: "t2", basemodel.FieldOrgId: "org-1"},
		{basemodel.FieldId: "t3", basemodel.FieldOrgId: "org-2"},
	}

	weights, err := fns[TemplateWeightUomSymbolFn](ctx, dynamicresource.ComputeFnRequest{Models: rows})
	require.NoError(t, err)
	lengths, err := fns[TemplateDimensionUomSymbolFn](ctx, dynamicresource.ComputeFnRequest{Models: rows})
	require.NoError(t, err)

	assert.Equal(t, []any{"kg", "kg", "kg"}, weights)
	assert.Equal(t, "cm", lengths[2])
	assert.Equal(t, 1, settings.calls)
	assert.Equal(t, 4, uoms.lookups, "two organizations, two units each")
}
