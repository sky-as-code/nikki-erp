package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type fakeUnitSettings struct {
	values map[string]any
	calls  int
}

func (this *fakeUnitSettings) GetEffectiveSettings(
	_ corectx.Context, _ itExt.GetEffectiveSettingsQuery,
) (*itExt.GetEffectiveSettingsResult, error) {
	this.calls++
	result := &itExt.GetEffectiveSettingsResult{HasData: true}
	result.Data.Values = this.values
	return result, nil
}

type fakeUomDirectory struct {
	byId     map[string]itExt.GetUomResultData
	bySymbol map[string]itExt.GetUomResultData
}

func (this *fakeUomDirectory) GetUom(_ corectx.Context, query itExt.GetUomQuery) (*itExt.GetUomResult, error) {
	uom, ok := this.byId[string(query.Id)]
	return &itExt.GetUomResult{Data: uom, HasData: ok}, nil
}

func (this *fakeUomDirectory) FindUomBySymbol(
	_ corectx.Context, query itExt.FindUomBySymbolQuery,
) (*itExt.GetUomResult, error) {
	uom, ok := this.bySymbol[string(query.OrgId)+"/"+query.Symbol]
	return &itExt.GetUomResult{Data: uom, HasData: ok}, nil
}

func unitFixture(stored map[string]any) (*OrgUnitResolver, *fakeUnitSettings) {
	kg := itExt.GetUomResultData{Id: "kg-1", Symbol: "kg", CategoryId: "weight-1", OrgId: "org-1",
		Name: model.LangJson{"en-US": "Kilogram"}}
	gram := itExt.GetUomResultData{Id: "g-1", Symbol: "g", CategoryId: "weight-1", OrgId: "org-1",
		Name: model.LangJson{"en-US": "Gram"}}
	cm := itExt.GetUomResultData{Id: "cm-1", Symbol: "cm", CategoryId: "length-1", OrgId: "org-1",
		Name: model.LangJson{"en-US": "Centimeter"}}
	kg2 := itExt.GetUomResultData{Id: "kg-2", Symbol: "kg", CategoryId: "weight-2", OrgId: "org-2"}
	settings := &fakeUnitSettings{values: stored}
	uoms := &fakeUomDirectory{
		byId:     map[string]itExt.GetUomResultData{"kg-1": kg, "g-1": gram, "cm-1": cm, "kg-2": kg2},
		bySymbol: map[string]itExt.GetUomResultData{"org-1/kg": kg, "org-1/cm": cm, "org-2/kg": kg2},
	}
	return NewOrgUnitResolver(settings, uoms), settings
}

func unitCtx() corectx.Context {
	return corectx.NewRequestContext(context.Background())
}

// A stored unit wins for its own organization; another organization falls back to its own default.
func TestResolveUnitsPrefersTheStoredUnitOfTheSameOrg(t *testing.T) {
	resolver, settings := unitFixture(map[string]any{"inventory.default_weight_uom_id": "g-1"})

	units, err := resolver.ResolveUnits(unitCtx(), []string{"org-1", "org-2"})

	require.NoError(t, err)
	assert.Equal(t, "g", units["org-1"][invsettings.OrgSettingDefaultWeightUomId].Symbol)
	assert.Equal(t, "cm", units["org-1"][invsettings.OrgSettingDefaultLengthUomId].Symbol)
	assert.Equal(t, "kg-2", string(units["org-2"][invsettings.OrgSettingDefaultWeightUomId].Id))
	assert.Nil(t, units["org-2"][invsettings.OrgSettingDefaultLengthUomId], "org-2 has no cm")
	assert.Equal(t, 1, settings.calls, "one settings read for every organization")
}

func TestDefaultUnitIdsResolvesTheOrgsOwnUnits(t *testing.T) {
	resolver, _ := unitFixture(nil)

	ids, err := resolver.DefaultUnitIds(unitCtx(), "org-1")

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		invsettings.OrgSettingDefaultWeightUomId: "kg-1",
		invsettings.OrgSettingDefaultLengthUomId: "cm-1",
	}, ids)
}

// A weight setting accepts only a weight unit of the same organization.
func TestAssertUnitSettingKeepsTheCategory(t *testing.T) {
	resolver, _ := unitFixture(nil)
	ctx := unitCtx()

	cErr, err := resolver.AssertUnitSetting(ctx, invsettings.OrgSettingDefaultWeightUomId, "g-1")
	require.NoError(t, err)
	assert.Nil(t, cErr)

	cErr, err = resolver.AssertUnitSetting(ctx, invsettings.OrgSettingDefaultWeightUomId, "cm-1")
	require.NoError(t, err)
	require.NotNil(t, cErr)
	assert.Equal(t, invsettings.OrgSettingDefaultWeightUomId, cErr.Field)

	cErr, err = resolver.AssertUnitSetting(ctx, invsettings.OrgSettingDefaultLengthUomId, "missing")
	require.NoError(t, err)
	require.NotNil(t, cErr)
}
