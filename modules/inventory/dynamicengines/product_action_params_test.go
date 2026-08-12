package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// resolve_selection declares no ParamSchema, so its body arrives untyped. Every shape problem
// below would otherwise decode to an empty selection list — which resolves to the empty
// combination, a real and different variant, answered 200 OK.

func TestBuildResolveSelectionQueryAcceptsAWellFormedBody(t *testing.T) {
	query, vErrs := buildResolveSelectionQuery(dmodel.DynamicFields{
		"template_id":            "01TEMPLATE",
		"materialize_if_missing": true,
		"selections": []any{
			map[string]any{"attribute_id": "01COLOR", "value_id": "01BLACK", "mode": "instant"},
			map[string]any{"attribute_id": "01SIZE", "value_id": "01M"},
		},
	})

	require.Zero(t, vErrs.Count())
	assert.Equal(t, "01TEMPLATE", query.TemplateId)
	assert.True(t, query.MaterializeIfMissing)
	require.Len(t, query.Selections, 2)
	assert.Equal(t, "01COLOR", query.Selections[0].AttributeId)
	// An omitted mode keeps the attribute in the combination rather than silently dropping it.
	assert.Equal(t, models.VariantCreationModeInstant, query.Selections[1].Mode)
}

// A template with no variant-generating attributes resolves on an empty selection, so an absent
// list is legitimate — it is a malformed one that must be rejected.
func TestBuildResolveSelectionQueryAllowsAbsentSelections(t *testing.T) {
	query, vErrs := buildResolveSelectionQuery(dmodel.DynamicFields{"template_id": "01TEMPLATE"})

	assert.Zero(t, vErrs.Count())
	assert.Empty(t, query.Selections)
}

func TestBuildResolveSelectionQueryRejectsMalformedBodies(t *testing.T) {
	testCases := []struct {
		name    string
		params  dmodel.DynamicFields
		wantKey string
	}{
		{
			name:    "a missing template is rejected",
			params:  dmodel.DynamicFields{"selections": []any{}},
			wantKey: "product_template.template_id_required",
		},
		{
			name:    "an empty template id is rejected",
			params:  dmodel.DynamicFields{"template_id": ""},
			wantKey: "product_template.template_id_required",
		},
		{
			name:    "selections that are not a list are rejected",
			params:  dmodel.DynamicFields{"template_id": "01TEMPLATE", "selections": "black"},
			wantKey: "product_template.selections_malformed",
		},
		{
			name: "a selection that is not an object is rejected",
			params: dmodel.DynamicFields{
				"template_id": "01TEMPLATE",
				"selections":  []any{"black"},
			},
			wantKey: "product_template.selection_malformed",
		},
		{
			name: "a selection missing its value is rejected",
			params: dmodel.DynamicFields{
				"template_id": "01TEMPLATE",
				"selections":  []any{map[string]any{"attribute_id": "01COLOR"}},
			},
			wantKey: "product_template.selection_incomplete",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, vErrs := buildResolveSelectionQuery(testCase.params)

			require.NotZero(t, vErrs.Count(), "a malformed body must not resolve silently")
			found := false
			for _, item := range *vErrs {
				if item.Key == testCase.wantKey {
					found = true
				}
			}
			assert.Truef(t, found, "expected %q, got %v", testCase.wantKey, *vErrs)
		})
	}
}
