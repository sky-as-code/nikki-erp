package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// valueNameRepository answers the attribute value read with canned names and counts the reads.
type valueNameRepository struct {
	composable.CrudRepository
	searches int
	rows     []dmodel.DynamicFields
}

func (this *valueNameRepository) Search(
	_ corectx.Context, _ dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	this.searches++
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data: dyn.PagedResultData[dmodel.DynamicFields]{Items: this.rows}, HasData: len(this.rows) > 0,
	}, nil
}

func useValueNames(t *testing.T) *valueNameRepository {
	t.Helper()
	repo := &valueNameRepository{rows: []dmodel.DynamicFields{
		{models.ProductAttributeValueFieldId: "val_nosugar", models.ProductAttributeValueFieldName: model.LangJson{"en-US": "No sugar"}},
		{models.ProductAttributeValueFieldId: "val_big", models.ProductAttributeValueFieldName: model.LangJson{"en-US": "Big"}},
		{models.ProductAttributeValueFieldId: "val_can", models.ProductAttributeValueFieldName: model.LangJson{"vi-VN": "Lon"}},
	}}
	original := repoFor
	t.Cleanup(func() { repoFor = original })
	repoFor = func(schemaName string) (composable.CrudRepository, error) {
		require.Equal(t, models.ProductAttributeValueSchemaName, schemaName)
		return repo, nil
	}
	return repo
}

// The stored lines put sugar before size; the attribute ids sort the other way, so a label that
// followed the combination key would read "Big, No sugar".
func labelOutcome(plan *VariantPlan) *reconcileOutcome {
	template := models.NewProductTemplate()
	template.SetName(&model.LangJson{"en-US": "Coca Cola", "vi-VN": "Cô ca"})
	existing := models.NewProductVariantFrom(dmodel.DynamicFields{
		models.ProductVariantFieldId: "var_1", models.ProductVariantFieldSku: "CC-NS-BIG",
	})
	return &reconcileOutcome{
		template: template,
		state: &templateConfigState{lines: []*configLineRow{
			{attributeId: "b_sugar", enabled: true, sequence: util.ToPtr(int32(1))},
			{attributeId: "a_size", enabled: true, sequence: util.ToPtr(int32(2))},
		}},
		desired: &desiredConfig{lines: []desiredLine{
			{attributeId: "b_sugar", sequence: util.ToPtr(int32(1))},
			{attributeId: "a_size"},
			{attributeId: "c_pack", sequence: util.ToPtr(int32(3))},
		}},
		plan:      plan,
		variants:  &templateVariants{byId: map[string]*models.ProductVariant{"var_1": existing}},
		decisions: map[string]invalidDecision{},
	}
}

func TestPreviewLabelsFollowLineOrderAndNameCreatesFromTheirValues(t *testing.T) {
	repo := useValueNames(t)
	plan := &VariantPlan{
		Keep: []PlannedVariant{{VariantId: "var_1", Key: "a_size:val_big|b_sugar:val_nosugar"}},
		Create: []PlannedVariant{
			{Key: "a_size:val_big|b_sugar:val_nosugar|c_pack:val_can"},
			{Key: ""},
		},
	}
	outcome := labelOutcome(plan)

	labeler, err := newPreviewLabeler(callerContext(), outcome)
	require.NoError(t, err)

	keep := previewItems(outcome, labeler, plan.Keep)
	created := previewItems(outcome, labeler, plan.Create)
	assert.Equal(t, "Coca Cola (No sugar, Big)", keep[0].DisplayName)
	assert.Empty(t, keep[0].NewDisplayName)
	assert.Equal(t, "Coca Cola (No sugar, Big, Lon)", created[0].DisplayName)
	assert.Equal(t, "Coca Cola", created[1].DisplayName, "an empty key is named by the template alone")
	assert.Equal(t, 1, repo.searches, "value names are read once for the whole preview")
}

func TestPreviewExtendCarriesItsCurrentAndItsNewLabel(t *testing.T) {
	useValueNames(t)
	plan := &VariantPlan{Extend: []PlannedVariant{{
		VariantId: "var_1", Key: "b_sugar:val_nosugar", NewKey: "b_sugar:val_nosugar|c_pack:val_can",
	}}}
	outcome := labelOutcome(plan)

	labeler, err := newPreviewLabeler(callerContext(), outcome)
	require.NoError(t, err)

	extend := previewItems(outcome, labeler, plan.Extend)
	assert.Equal(t, "Coca Cola (No sugar)", extend[0].DisplayName)
	assert.Equal(t, "Coca Cola (No sugar, Lon)", extend[0].NewDisplayName)
}

// A value the catalogue no longer holds is left out, as display_name leaves it out.
func TestPreviewLabelSkipsAValueWithNoName(t *testing.T) {
	useValueNames(t)
	plan := &VariantPlan{Invalid: []PlannedVariant{{VariantId: "var_1", Key: "a_size:val_gone|b_sugar:val_nosugar"}}}
	outcome := labelOutcome(plan)

	labeler, err := newPreviewLabeler(callerContext(), outcome)
	require.NoError(t, err)

	archive, remove := splitInvalid(outcome, labeler)
	require.Len(t, archive, 1)
	assert.Empty(t, remove)
	assert.Equal(t, "Coca Cola (No sugar)", archive[0].DisplayName)
}
