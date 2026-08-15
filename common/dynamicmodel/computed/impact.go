package computed

import (
	"sort"
	"strings"

	"go.bryk.io/pkg/errors"
)

// Impact analysis over the finalized plans (spec §20). Boot already fails when a computed field
// points at a schema element that no longer exists — resolution reports the unknown name — so
// these helpers serve proactive checks: tooling validating a schema change before it is applied.

// Dependents lists every computed field whose definition depends on the named schema element
// (a physical field, an edge, or another computed field), sorted for stable output.
func Dependents(schemaName string, elementName string) []FieldRef {
	target := FieldRef{Schema: schemaName, Field: elementName}
	plansMu.RLock()
	defer plansMu.RUnlock()

	var result []FieldRef
	for _, schemaPlan := range plans {
		for _, fieldPlan := range schemaPlan.Fields {
			if dependsOn(fieldPlan, target) {
				result = append(result, FieldRef{Schema: schemaPlan.SchemaName, Field: fieldPlan.FieldName})
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}

func dependsOn(plan *FieldPlan, target FieldRef) bool {
	for _, dep := range plan.Dependencies {
		if dep == target {
			return true
		}
	}
	return false
}

// AssertNoDependents returns the spec §20 error when deleting or renaming the element would
// break computed fields, and nil when the change is safe.
func AssertNoDependents(schemaName string, elementName string) error {
	dependents := Dependents(schemaName, elementName)
	if len(dependents) == 0 {
		return nil
	}
	lines := make([]string, len(dependents))
	for i, ref := range dependents {
		lines[i] = "- " + ref.String()
	}
	return errors.Errorf(
		"Cannot delete field %q.\n\nDependent computed fields:\n%s",
		elementName, strings.Join(lines, "\n"))
}
