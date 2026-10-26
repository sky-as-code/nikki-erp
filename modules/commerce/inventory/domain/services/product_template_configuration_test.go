package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
)

func linesInput(lines ...itProduct.AttributeLineInput) itProduct.TemplateConfigurationInput {
	return itProduct.TemplateConfigurationInput{AttributeLines: &lines}
}

func TestConfigurationShapeRefusesEmptyLineDuplicateAttributeAndValue(t *testing.T) {
	cErrs := validateConfigurationShape(linesInput(
		itProduct.AttributeLineInput{AttributeId: "COLOR", ValueIds: []string{"RED"}},
		itProduct.AttributeLineInput{AttributeId: "SIZE"},
		itProduct.AttributeLineInput{AttributeId: "COLOR", ValueIds: []string{"BLUE", "BLUE"}},
		itProduct.AttributeLineInput{ValueIds: []string{"X"}},
	))
	fields := map[string]string{}
	for _, item := range cErrs {
		fields[item.Field] = item.Key
	}
	assert.Contains(t, fields["attribute_lines[1].value_ids"], "line_without_values")
	assert.Contains(t, fields["attribute_lines[2].attribute_id"], "duplicate_attribute_line")
	assert.Contains(t, fields["attribute_lines[2].value_ids"], "duplicate_line_value")
	assert.Contains(t, fields, "attribute_lines[3].attribute_id")
}

func TestConfigurationShapeAcceptsValidLines(t *testing.T) {
	cErrs := validateConfigurationShape(linesInput(
		itProduct.AttributeLineInput{AttributeId: "COLOR", ValueIds: []string{"RED", "BLUE"}},
		itProduct.AttributeLineInput{AttributeId: "SIZE", ValueIds: []string{"S"}},
	))
	assert.Empty(t, cErrs)
}

func TestDesiredExclusionsNormalizeDedupAndValidate(t *testing.T) {
	lines := []desiredLine{
		{attributeId: "COLOR", valueIds: []string{"RED", "BLUE"}},
		{attributeId: "SIZE", valueIds: []string{"S", "M"}},
	}
	out, cErrs := resolveDesiredExclusions(lines, []itProduct.ExclusionInput{
		{LeftValueId: "S", RightValueId: "RED"},
		{LeftValueId: "RED", RightValueId: "S"},
		{LeftValueId: "RED", RightValueId: "BLUE"},
		{LeftValueId: "RED", RightValueId: "XL"},
		{LeftValueId: "M", RightValueId: "M"},
	})
	assert.Equal(t, []PlanExclusion{{LeftValueId: "RED", RightValueId: "S"}}, out)
	require.Len(t, cErrs, 3)
	assert.Equal(t, "exclusions[2]", cErrs[0].Field)
	assert.Contains(t, cErrs[0].Key, "same_attribute")
	assert.Contains(t, cErrs[1].Key, "exclusion_value_not_selected")
	assert.Contains(t, cErrs[2].Key, "same_value")
}

func TestKeepSelectedExclusionsDropsPairsOfUnselectedValues(t *testing.T) {
	lines := []desiredLine{{attributeId: "COLOR", valueIds: []string{"RED"}}, {attributeId: "SIZE", valueIds: []string{"S"}}}
	kept := keepSelectedExclusions([]PlanExclusion{
		{LeftValueId: "RED", RightValueId: "S"}, {LeftValueId: "BLUE", RightValueId: "S"},
	}, lines)
	assert.Equal(t, []PlanExclusion{{LeftValueId: "RED", RightValueId: "S"}}, kept)
}

func TestRemapPlanErrorsPointsAtPayloadRow(t *testing.T) {
	desired := &desiredConfig{lines: []desiredLine{{attributeId: "COLOR", row: 0}, {attributeId: "SIZE", row: 1}}}
	plan := &VariantPlan{}
	plan.violate(ReasonPlanDuplicateLine, "dup", map[string]any{"attribute_id": "SIZE"})
	plan.violate(ReasonPlanNoValidCombination, "none", nil)
	remapped := remapPlanErrors(plan.ClientErrors, desired)
	assert.Equal(t, "attribute_lines[1].attribute_id", remapped[0].Field)
	assert.Equal(t, PlanFieldAttributeLines, remapped[1].Field)
}

// newStateFixture is a template with Color (instant, Red selected, Blue unselected), a disabled
// Size revision 1 and an enabled Size revision 2 (dynamic), and one active and one archived
// exclusion.
func newStateFixture() *templateConfigState {
	color := &configLineRow{id: "L-COLOR", attributeId: "COLOR", enabled: true, revision: 1,
		snapshot: models.VariantCreationModeInstant, values: []*configValueRow{
			{id: "TV-RED", valueId: "RED", selected: true}, {id: "TV-BLUE", valueId: "BLUE", selected: false},
		}}
	oldSize := &configLineRow{id: "L-SIZE-1", attributeId: "SIZE", enabled: false, revision: 1,
		snapshot: models.VariantCreationModeInstant, values: []*configValueRow{{id: "TV-S-OLD", valueId: "S", selected: true}}}
	size := &configLineRow{id: "L-SIZE-2", attributeId: "SIZE", enabled: true, revision: 2,
		snapshot: models.VariantCreationModeDynamic, sequence: util.ToPtr(int32(2)),
		values: []*configValueRow{{id: "TV-S", valueId: "S", selected: true}}}
	state := &templateConfigState{templateId: "T1", lines: []*configLineRow{color, oldSize, size}, tavIndex: map[string]tavInfo{}}
	for _, line := range state.lines {
		for _, value := range line.values {
			state.tavIndex[value.id] = tavInfo{valueId: value.valueId, line: line}
		}
	}
	state.exclusions = []*configExclusionRow{
		{id: "X1", leftId: "TV-RED", rightId: "TV-S"},
		{id: "X2", leftId: "TV-BLUE", rightId: "TV-S", archived: true},
	}
	return state
}

func TestStatePlanConfigUsesEnabledLinesSnapshotsAndActiveExclusions(t *testing.T) {
	state := newStateFixture()
	config := state.planConfig()
	require.Len(t, config.Lines, 2)
	assert.Equal(t, PlanLine{AttributeId: "COLOR", Mode: models.VariantCreationModeInstant, ValueIds: []string{"RED"}}, config.Lines[0])
	assert.Equal(t, models.VariantCreationModeDynamic, config.Lines[1].Mode)
	assert.Equal(t, []PlanExclusion{{LeftValueId: "RED", RightValueId: "S"}}, config.Exclusions)
	assert.Equal(t, int32(2), state.maxRevision("SIZE"))
	assert.Equal(t, "L-SIZE-2", state.enabledLine("SIZE").id)
}

func TestStateValueRefsOnlyCurrentSelections(t *testing.T) {
	refs := newStateFixture().valueRefs()
	assert.Equal(t, "TV-S", refs["S"].TemplateValueId)
	assert.Equal(t, "TV-RED", refs["RED"].TemplateValueId)
	assert.NotContains(t, refs, "BLUE")
}

func TestPreviousPlanConfigIsNilWithoutGeneratingLine(t *testing.T) {
	state := &templateConfigState{lines: []*configLineRow{{attributeId: "NOTE", enabled: true,
		snapshot: models.VariantCreationModeNever, values: []*configValueRow{{valueId: "A", selected: true}}}}}
	assert.Nil(t, state.previousPlanConfig())
	assert.NotNil(t, newStateFixture().previousPlanConfig())
}

func TestDecideArchivesEverythingWithoutDeletePermission(t *testing.T) {
	decider := usageDecider{deleteAllowed: false}
	decisions, cErrs, err := decider.decide(nil, []PlannedVariant{{VariantId: "V1"}, {VariantId: "V2"}})
	require.NoError(t, err)
	assert.Empty(t, cErrs)
	assert.Equal(t, invalidDecision{reason: ArchiveReasonNoDeletePermission}, decisions["V1"])
	assert.False(t, decisions["V2"].delete)
}

func TestKeepReferencedBasesArchivesBaseWithSurvivingChild(t *testing.T) {
	decider := usageDecider{children: map[string][]string{"B1": {"C1", "C2"}, "B2": {"C3"}}}
	decisions := map[string]invalidDecision{
		"B1": {delete: true}, "C1": {delete: true}, "C2": {reason: ArchiveReasonEverUsed},
		"B2": {delete: true}, "C3": {delete: true},
	}
	decider.keepReferencedBases(decisions)
	assert.Equal(t, ArchiveReasonHasDependents, decisions["B1"].reason)
	assert.False(t, decisions["B1"].delete)
	assert.True(t, decisions["B2"].delete)
}

func TestKeepReferencedBasesKeepsBaseWithUnplannedChild(t *testing.T) {
	decider := usageDecider{children: map[string][]string{"B1": {"C-ARCHIVED-BY-USER"}}}
	decisions := map[string]invalidDecision{"B1": {delete: true}}
	decider.keepReferencedBases(decisions)
	assert.False(t, decisions["B1"].delete)
}

func TestCollectBaseIdsUsesExtendedKey(t *testing.T) {
	executor := &planExecutor{}
	executor.collectBaseIds(&VariantPlan{
		Keep:   []PlannedVariant{{VariantId: "B1", Kind: models.VariantKindBase, Key: "C:R"}},
		Extend: []PlannedVariant{{VariantId: "B2", Kind: models.VariantKindBase, Key: "C:B", NewKey: "C:B|F:X"}},
		// A complete never lands in the base index.
		Restore: []PlannedVariant{{VariantId: "C1", Kind: models.VariantKindComplete, Key: "C:R|S:M"}},
	})
	assert.Equal(t, map[string]string{"C:R": "B1", "C:B|F:X": "B2"}, executor.baseIds)
}

func TestLiftTemplateConfigurationSeparatesPayload(t *testing.T) {
	cmd := map[string]any{
		"name":                    "Shirt",
		"attribute_lines":         []any{map[string]any{"attribute_id": "COLOR", "value_ids": []any{"RED"}, "sequence": 1}},
		"variant_delete_allowed":  true,
		"attribute_configuration": map[string]any{"ignored": true},
	}
	input, deleteAllowed, malformed := itProduct.LiftTemplateConfiguration(cmd)
	require.NotNil(t, input)
	assert.Empty(t, malformed)
	assert.True(t, deleteAllowed)
	assert.Nil(t, input.Exclusions)
	require.Len(t, *input.AttributeLines, 1)
	assert.Equal(t, []string{"RED"}, (*input.AttributeLines)[0].ValueIds)
	assert.Equal(t, map[string]any{"name": "Shirt"}, map[string]any(cmd))
}

func TestLiftTemplateConfigurationReportsMalformedAndAbsent(t *testing.T) {
	input, _, malformed := itProduct.LiftTemplateConfiguration(map[string]any{"exclusions": "nope"})
	assert.Nil(t, input)
	assert.Equal(t, itProduct.ProductTemplateFieldExclusions, malformed)

	input, deleteAllowed, malformed := itProduct.LiftTemplateConfiguration(map[string]any{"name": "x"})
	assert.Nil(t, input)
	assert.False(t, deleteAllowed)
	assert.Empty(t, malformed)
}

func TestLiftTemplateConfigurationEmptyListMeansRemoveAll(t *testing.T) {
	input, _, _ := itProduct.LiftTemplateConfiguration(map[string]any{"attribute_lines": []any{}})
	require.NotNil(t, input)
	require.NotNil(t, input.AttributeLines)
	assert.Empty(t, *input.AttributeLines)
}
