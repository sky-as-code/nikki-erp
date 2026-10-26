package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

func copySourceTemplate() dmodel.DynamicFields {
	return dmodel.DynamicFields{
		basemodel.FieldId:                               "t-1",
		basemodel.FieldEtag:                             "9",
		basemodel.FieldIsArchived:                       true,
		basemodel.FieldOrgId:                            "org-1",
		models.ProductTemplateFieldName:                 map[string]any{"en-US": "Tea", "vi-VN": "Trà"},
		models.ProductTemplateFieldStatus:               "active",
		models.ProductTemplateFieldDefaultImageFilePath: "files/tea.png",
		"external_id":                                   "ext-1",
		"source_system":                                 "erp-x",
		models.ProductTemplateFieldMinVariantCost:       "1.5",
	}
}

func TestCopyTemplateFieldsCarriesOnlyTheTemplatesOwnFields(t *testing.T) {
	payload, err := copyTemplateFields(copySourceTemplate(), nil)

	require.NoError(t, err)
	assert.Equal(t, map[string]any{"en-US": "Tea (copy)", "vi-VN": "Trà (copy)"}, payload[models.ProductTemplateFieldName])
	assert.Equal(t, "active", payload[models.ProductTemplateFieldStatus])
	for _, dropped := range []string{
		basemodel.FieldId, basemodel.FieldEtag, basemodel.FieldIsArchived, basemodel.FieldOrgId,
		models.ProductTemplateFieldDefaultImageFilePath, "external_id", "source_system",
		models.ProductTemplateFieldMinVariantCost,
	} {
		assert.NotContains(t, payload, dropped)
	}
}

func TestCopyTemplateFieldsUsesAnExplicitName(t *testing.T) {
	name := map[string]any{"en-US": "Iced tea"}
	payload, err := copyTemplateFields(copySourceTemplate(), name)

	require.NoError(t, err)
	assert.Equal(t, name, payload[models.ProductTemplateFieldName])
}

// Only the current configuration is copied: a disabled line, an unselected value, an archived
// exclusion and an exclusion naming an unselected value all stay behind.
func TestCopyConfigurationTakesTheCurrentConfigurationOnly(t *testing.T) {
	sequence := int32(2)
	color := &configLineRow{id: "l-1", attributeId: "a-color", enabled: true, snapshot: models.VariantCreationModeInstant,
		sequence: &sequence, values: []*configValueRow{
			{id: "tv-red", valueId: "red", selected: true},
			{id: "tv-blue", valueId: "blue", selected: false},
		}}
	size := &configLineRow{id: "l-2", attributeId: "a-size", enabled: true, snapshot: models.VariantCreationModeDynamic,
		values: []*configValueRow{{id: "tv-m", valueId: "m", selected: true}}}
	oldLine := &configLineRow{id: "l-3", attributeId: "a-old", enabled: false,
		values: []*configValueRow{{id: "tv-x", valueId: "x", selected: true}}}
	state := &templateConfigState{
		lines: []*configLineRow{color, size, oldLine},
		exclusions: []*configExclusionRow{
			{id: "e-1", leftId: "tv-red", rightId: "tv-m"},
			{id: "e-2", leftId: "tv-red", rightId: "tv-m", archived: true},
			{id: "e-3", leftId: "tv-blue", rightId: "tv-m"},
		},
		tavIndex: map[string]tavInfo{
			"tv-red": {valueId: "red", line: color}, "tv-blue": {valueId: "blue", line: color},
			"tv-m": {valueId: "m", line: size}, "tv-x": {valueId: "x", line: oldLine},
		},
	}

	lines, exclusions := copyConfiguration(state)

	assert.Equal(t, []itProduct.AttributeLineInput{
		{AttributeId: "a-size", ValueIds: []string{"m"}},
		{AttributeId: "a-color", ValueIds: []string{"red"}, Sequence: &sequence},
	}, lines)
	assert.Equal(t, []itProduct.ExclusionInput{{LeftValueId: "red", RightValueId: "m"}}, exclusions)
}
