package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// resolve_selection declares no ParamSchema, so its body arrives untyped. Every shape problem
// below would otherwise decode to an empty selection list — which resolves to the empty
// combination, a real and different variant, answered 200 OK.

func TestBuildResolveCompleteQueryAcceptsAWellFormedBody(t *testing.T) {
	query, vErrs := buildResolveCompleteQuery(dmodel.DynamicFields{
		"product_template_id":    "01TEMPLATE",
		"materialize_if_missing": true,
		"expected_etag":          "42",
		"selections": []any{
			map[string]any{"attribute_id": "01COLOR", "value_id": "01BLACK"},
			map[string]any{"attribute_id": "01SIZE", "value_id": "01M"},
		},
	})

	require.Zero(t, vErrs.Count())
	assert.Equal(t, "01TEMPLATE", query.TemplateId)
	assert.True(t, query.MaterializeIfMissing)
	require.NotNil(t, query.ExpectedEtag)
	assert.Equal(t, "42", *query.ExpectedEtag)
	assert.Equal(t, map[string]string{"01COLOR": "01BLACK", "01SIZE": "01M"}, query.Selections)
}

// template_id is the name the endpoint was first published with.
func TestBuildResolveCompleteQueryAcceptsTheLegacyTemplateField(t *testing.T) {
	query, vErrs := buildResolveCompleteQuery(dmodel.DynamicFields{"template_id": "01TEMPLATE"})

	require.Zero(t, vErrs.Count())
	assert.Equal(t, "01TEMPLATE", query.TemplateId)
	assert.False(t, query.MaterializeIfMissing, "absent flag is a lookup that never writes")
}

// A template with no variant-generating attributes resolves on an empty selection, so an absent
// list is legitimate — it is a malformed one that must be rejected.
func TestBuildResolveCompleteQueryAllowsAbsentSelections(t *testing.T) {
	query, vErrs := buildResolveCompleteQuery(dmodel.DynamicFields{"template_id": "01TEMPLATE"})

	assert.Zero(t, vErrs.Count())
	assert.Empty(t, query.Selections)
}

func TestBuildResolveCompleteQueryRejectsMalformedBodies(t *testing.T) {
	testCases := []struct {
		name    string
		params  dmodel.DynamicFields
		wantKey string
	}{
		{
			name:    "a missing template is rejected",
			params:  dmodel.DynamicFields{"selections": []any{}},
			wantKey: "inventory:product_template.template_id_required",
		},
		{
			name:    "an empty template id is rejected",
			params:  dmodel.DynamicFields{"template_id": ""},
			wantKey: "inventory:product_template.template_id_required",
		},
		{
			name:    "selections that are not a list are rejected",
			params:  dmodel.DynamicFields{"template_id": "01TEMPLATE", "selections": "black"},
			wantKey: "inventory:product_template.selections_malformed",
		},
		{
			name: "a selection that is not an object is rejected",
			params: dmodel.DynamicFields{
				"template_id": "01TEMPLATE",
				"selections":  []any{"black"},
			},
			wantKey: "inventory:product_template.selection_malformed",
		},
		{
			name: "two values for one attribute are rejected",
			params: dmodel.DynamicFields{
				"template_id": "01TEMPLATE",
				"selections": []any{
					map[string]any{"attribute_id": "01COLOR", "value_id": "01BLACK"},
					map[string]any{"attribute_id": "01COLOR", "value_id": "01WHITE"},
				},
			},
			wantKey: "inventory:product_template.selection_duplicate_attribute",
		},
		{
			name: "a selection missing its value is rejected",
			params: dmodel.DynamicFields{
				"template_id": "01TEMPLATE",
				"selections":  []any{map[string]any{"attribute_id": "01COLOR"}},
			},
			wantKey: "inventory:product_template.selection_incomplete",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, vErrs := buildResolveCompleteQuery(testCase.params)

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
