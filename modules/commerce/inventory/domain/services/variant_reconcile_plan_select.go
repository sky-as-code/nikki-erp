package services

import (
	"sort"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// PlanFieldSelection is the resolve request field selection violations point at.
const PlanFieldSelection = "selections"

// ValidateSelection checks a resolve selection (catalogue attribute id -> catalogue value id)
// against a configuration: exactly one selected value for every instant and dynamic line, no
// attribute outside the configuration, no value outside its line, and no active exclusion. Never
// values are validated when present but stay out of both keys. baseKey is set only for a mixed
// configuration.
func ValidateSelection(
	config PlanConfig, selection map[string]string,
) (completeKey string, baseKey string, violations ft.ClientErrors) {
	state := newPlanState(config)
	violations = state.checkSelectedValues(selection)
	values := make([]string, 0, len(state.instant)+len(state.dynamic))
	for _, dim := range state.generating() {
		valueId, found := selection[dim.attributeId]
		if !found {
			violations.Append(*selectionViolation(ReasonSelectionMissingAttr,
				"a value is required for attribute {{attribute_id}}",
				map[string]any{"attribute_id": dim.attributeId}))
			continue
		}
		values = append(values, valueId)
	}
	if violations.Count() > 0 {
		return "", "", violations
	}
	if pair := state.firstExcludedPair(values); pair != nil {
		violations.Append(*selectionViolation(ReasonSelectionExcluded,
			"values {{left_value_id}} and {{right_value_id}} cannot be combined",
			map[string]any{"left_value_id": pair.LeftValueId, "right_value_id": pair.RightValueId}))
		return "", "", violations
	}
	completeKey = state.keyOfValues(values)
	if state.isMixed() {
		baseKey = state.instantSubKey(completeKey)
	}
	return completeKey, baseKey, nil
}

func (this *planState) checkSelectedValues(selection map[string]string) ft.ClientErrors {
	attributeIds := make([]string, 0, len(selection))
	for attributeId := range selection {
		attributeIds = append(attributeIds, attributeId)
	}
	sort.Strings(attributeIds)

	violations := ft.ClientErrors{}
	for _, attributeId := range attributeIds {
		vars := map[string]any{"attribute_id": attributeId, "value_id": selection[attributeId]}
		line, found := this.lines[attributeId]
		if !found {
			violations.Append(*selectionViolation(ReasonSelectionUnknownAttr,
				"attribute {{attribute_id}} is not configured on the template", vars))
			continue
		}
		if !line.valueSet[selection[attributeId]] {
			violations.Append(*selectionViolation(ReasonSelectionValueNotAllow,
				"value {{value_id}} is not selectable for attribute {{attribute_id}}", vars))
		}
	}
	return violations
}

func (this *planState) firstExcludedPair(values []string) *PlanExclusion {
	for i, left := range values {
		for _, right := range values[i+1:] {
			if this.exclusions[exclusionPairKey(left, right)] {
				return &PlanExclusion{LeftValueId: left, RightValueId: right}
			}
		}
	}
	return nil
}

func selectionViolation(key string, message string, vars map[string]any) *ft.ClientErrorItem {
	return ft.NewBusinessViolation(PlanFieldSelection, ft.ErrorKey(key, "inventory"), message, vars)
}
