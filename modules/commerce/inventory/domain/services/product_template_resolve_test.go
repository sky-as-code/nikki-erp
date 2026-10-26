package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func resolveVariant(id string, archived bool, source models.ArchiveSource) *models.ProductVariant {
	fields := dmodel.DynamicFields{basemodel.FieldId: id, basemodel.FieldIsArchived: archived}
	if source != "" {
		fields[models.ProductVariantFieldArchiveSource] = source.String()
	}
	return models.NewProductVariantFrom(fields)
}

func TestDecideResolveOutcomes(t *testing.T) {
	active := resolveVariant("v-1", false, "")
	synced := resolveVariant("v-2", true, models.ArchiveSourceSystemSync)
	testCases := []struct {
		name      string
		facts     resolveFacts
		want      resolveAction
		wantError string
	}{
		{"active complete is returned", resolveFacts{Complete: active}, resolveActionExisting, ""},
		{"missing complete is created on demand", resolveFacts{Materialize: true}, resolveActionCreate, ""},
		{"lookup only never creates", resolveFacts{}, resolveActionMissing, ""},
		{"configuration-archived complete is restored", resolveFacts{Complete: synced, Materialize: true}, resolveActionRestore, ""},
		{"lookup only never restores", resolveFacts{Complete: synced}, resolveActionMissing, ""},
		{"user-archived complete is refused, never duplicated",
			resolveFacts{Complete: resolveVariant("v-3", true, models.ArchiveSourceUser), Materialize: true},
			"", "inventory:" + ReasonResolveManuallyArchived},
		{"archived complete without a source counts as the user's",
			resolveFacts{Complete: resolveVariant("v-4", true, ""), Materialize: true},
			"", "inventory:" + ReasonResolveManuallyArchived},
		{"template-cascade-archived complete is refused",
			resolveFacts{Complete: resolveVariant("v-5", true, models.ArchiveSourceTemplateCascade), Materialize: true},
			"", "inventory:" + ReasonResolveManuallyArchived},
		{"mixed without a base is refused", resolveFacts{Mixed: true, Materialize: true},
			"", "inventory:" + ReasonResolveBaseMissing},
		{"mixed with an archived base is refused",
			resolveFacts{Mixed: true, Base: resolveVariant("b-1", true, models.ArchiveSourceUser), Materialize: true},
			"", "inventory:" + ReasonResolveBaseArchived},
		{"mixed with an active base creates the complete",
			resolveFacts{Mixed: true, Base: resolveVariant("b-2", false, ""), Materialize: true},
			resolveActionCreate, ""},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			action, refusal := decideResolve(testCase.facts)
			if testCase.wantError != "" {
				require.NotNil(t, refusal)
				assert.Equal(t, testCase.wantError, refusal.Key)
				return
			}
			require.Nil(t, refusal)
			assert.Equal(t, testCase.want, action)
		})
	}
}

func TestAssertTemplateResolvable(t *testing.T) {
	etag := model.Etag("7")
	active := models.NewProductTemplateFrom(dmodel.DynamicFields{basemodel.FieldIsArchived: false})
	active.SetEtag(&etag)
	archived := models.NewProductTemplateFrom(dmodel.DynamicFields{basemodel.FieldIsArchived: true})
	stale, current := "6", "7"

	assert.Nil(t, assertTemplateResolvable(active, nil))
	assert.Nil(t, assertTemplateResolvable(active, &current))
	refused := assertTemplateResolvable(active, &stale)
	require.NotNil(t, refused)
	assert.Equal(t, "common:err_etag_mismatched", (*refused)[0].Key)
	refused = assertTemplateResolvable(archived, nil)
	require.NotNil(t, refused)
	assert.Equal(t, "inventory:"+ReasonResolveTemplateArchived, (*refused)[0].Key)
}

// Never values are validated by ValidateSelection, kept out of the key, and echoed back here.
func TestResolveKeyExcludesNeverValuesAndEchoesThem(t *testing.T) {
	config := PlanConfig{Lines: []PlanLine{
		{AttributeId: "a-color", Mode: models.VariantCreationModeInstant, ValueIds: []string{"red", "blue"}},
		{AttributeId: "a-size", Mode: models.VariantCreationModeDynamic, ValueIds: []string{"m", "l"}},
		{AttributeId: "a-note", Mode: models.VariantCreationModeNever, ValueIds: []string{"gift"}},
	}}
	selection := map[string]string{"a-color": "red", "a-size": "m", "a-note": "gift"}

	completeKey, baseKey, violations := ValidateSelection(config, selection)

	require.Zero(t, violations.Count())
	assert.Equal(t, "a-color:red|a-size:m", completeKey)
	assert.Equal(t, "a-color:red", baseKey)
	echoed := neverSelections(config, selection)
	require.Len(t, echoed, 1)
	assert.Equal(t, "a-note", echoed[0].AttributeId)
	assert.Equal(t, "gift", echoed[0].ValueId)
}
