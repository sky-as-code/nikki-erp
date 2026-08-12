// Package services holds the Products logic that is not CRUD over a single record, and so
// cannot be expressed by the dynamic resource engine: building a variant's identity out of an
// attribute combination, and resolving the effective product a consumer module sees.
package services

import (
	"sort"
	"strings"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// combinationSeparator joins the pairs of a combination key; combinationPairSeparator joins an
// attribute to its value within one pair. Neither may appear in a ULID, so a key can always be
// parsed back apart.
const (
	combinationSeparator     = "|"
	combinationPairSeparator = ":"
)

// BuildCombinationKey turns a set of attribute-value choices into the normalized string that
// identifies a variant within its template.
//
// The normalization is what makes the key an identity rather than a rendering: NEVER-mode
// attributes drop out, duplicates of the same attribute collapse to the last choice, and the
// remainder sort by attribute id. Two requests that pick the same values in a different order
// therefore produce the same key and resolve to the same variant. See BR §14.3 and §4.8.
//
// A template with no variant-generating attributes yields the empty key, which is a real key:
// it identifies that template's single concrete variant. See BR §4.5 and AC-PROD-008.
func BuildCombinationKey(selections []itProduct.AttributeSelection) string {
	byAttribute := map[string]string{}
	for _, selection := range selections {
		if !selection.Mode.CreatesVariants() {
			continue
		}
		if selection.AttributeId == "" || selection.ValueId == "" {
			continue
		}
		byAttribute[selection.AttributeId] = selection.ValueId
	}

	if len(byAttribute) == 0 {
		return models.EmptyCombinationKey
	}

	attributeIds := make([]string, 0, len(byAttribute))
	for attributeId := range byAttribute {
		attributeIds = append(attributeIds, attributeId)
	}
	// Sorting by attribute id, not by the attribute's display sequence: reordering attributes
	// in the UI must never change the identity of an existing variant.
	sort.Strings(attributeIds)

	pairs := make([]string, 0, len(attributeIds))
	for _, attributeId := range attributeIds {
		pairs = append(pairs, attributeId+combinationPairSeparator+byAttribute[attributeId])
	}
	return strings.Join(pairs, combinationSeparator)
}

// ParseCombinationKey splits a key back into its attribute-value pairs, in key order. The empty
// key parses to no pairs rather than to an error.
func ParseCombinationKey(key string) []itProduct.AttributeSelection {
	if key == models.EmptyCombinationKey {
		return nil
	}

	parts := strings.Split(key, combinationSeparator)
	selections := make([]itProduct.AttributeSelection, 0, len(parts))
	for _, part := range parts {
		attributeId, valueId, found := strings.Cut(part, combinationPairSeparator)
		if !found || attributeId == "" || valueId == "" {
			continue
		}
		selections = append(selections, itProduct.AttributeSelection{
			AttributeId: attributeId,
			ValueId:     valueId,
		})
	}
	return selections
}

// BuildInstantCombinations returns every combination an INSTANT-mode configuration implies: the
// cartesian product of the allowed values, one value per attribute.
//
// DYNAMIC and NEVER attributes are excluded. A DYNAMIC attribute deliberately has no combination
// generated ahead of use; a NEVER attribute is not part of identity at all. A template whose
// attributes are all excluded yields exactly one combination, the empty one, which is the
// single-variant case rather than an absence of variants. See BR §4.7 and §8.2.
func BuildInstantCombinations(attributes []itProduct.AttributeOptions) []string {
	generating := make([]itProduct.AttributeOptions, 0, len(attributes))
	for _, attribute := range attributes {
		if attribute.Mode != models.VariantCreationModeInstant {
			continue
		}
		if attribute.AttributeId == "" || len(attribute.ValueIds) == 0 {
			// An attribute with no allowed values would multiply the product by zero and
			// wipe out every combination, which is not what an unconfigured attribute means.
			continue
		}
		generating = append(generating, attribute)
	}

	if len(generating) == 0 {
		return []string{models.EmptyCombinationKey}
	}

	sort.Slice(generating, func(i, j int) bool {
		return generating[i].AttributeId < generating[j].AttributeId
	})

	combinations := [][]itProduct.AttributeSelection{{}}
	for _, attribute := range generating {
		expanded := make([][]itProduct.AttributeSelection, 0, len(combinations)*len(attribute.ValueIds))
		for _, combination := range combinations {
			for _, valueId := range attribute.ValueIds {
				next := make([]itProduct.AttributeSelection, len(combination), len(combination)+1)
				copy(next, combination)
				next = append(next, itProduct.AttributeSelection{
					AttributeId: attribute.AttributeId,
					ValueId:     valueId,
					Mode:        attribute.Mode,
				})
				expanded = append(expanded, next)
			}
		}
		combinations = expanded
	}

	keys := make([]string, 0, len(combinations))
	for _, combination := range combinations {
		keys = append(keys, BuildCombinationKey(combination))
	}
	return keys
}

// PlanVariantSync compares the combinations a template's attributes imply against the
// combination keys its non-archived variants currently hold.
func PlanVariantSync(wanted []string, existing []string) itProduct.VariantSyncPlan {
	wantedSet := map[string]bool{}
	for _, key := range wanted {
		wantedSet[key] = true
	}
	existingSet := map[string]bool{}
	for _, key := range existing {
		existingSet[key] = true
	}

	plan := itProduct.VariantSyncPlan{}
	for _, key := range wanted {
		if existingSet[key] {
			plan.Unchanged = append(plan.Unchanged, key)
			continue
		}
		// A wanted key repeated in the input must not be created twice.
		if !containsKey(plan.ToCreate, key) {
			plan.ToCreate = append(plan.ToCreate, key)
		}
	}
	for _, key := range existing {
		if !wantedSet[key] && !containsKey(plan.Obsolete, key) {
			plan.Obsolete = append(plan.Obsolete, key)
		}
	}
	return plan
}

func containsKey(keys []string, key string) bool {
	for _, item := range keys {
		if item == key {
			return true
		}
	}
	return false
}
