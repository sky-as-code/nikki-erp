// Package services holds the Products logic that is not CRUD over a single record: building a
// variant's identity out of an attribute combination, and resolving the effective product a
// consumer module sees.
package services

import (
	"sort"
	"strings"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// combinationSeparator joins the pairs of a combination key; combinationPairSeparator joins an
// attribute to its value within a pair. Neither may appear in a ULID, so a key always parses back
// apart.
const (
	combinationSeparator     = "|"
	combinationPairSeparator = ":"
)

// BuildCombinationKey turns a set of attribute-value choices into the normalized string that
// identifies a variant within its template. The normalization is what makes it an identity: NEVER
// attributes drop out, duplicates collapse to the last choice, and the rest sort by attribute id,
// so the same values picked in a different order produce the same key.
//
// A template with no variant-generating attributes yields the empty key, which is a real key
// identifying that template's single concrete variant.
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

	return BuildKeyFromPairs(byAttribute)
}

// BuildKeyFromPairs builds a combination key from catalogue attribute id -> catalogue value id
// pairs. The caller has already dropped NEVER attributes; empty ids are skipped. Pairs sort by
// attribute id, not display sequence: reordering attributes in the UI must never change the
// identity of an existing variant.
func BuildKeyFromPairs(pairs map[string]string) string {
	attributeIds := make([]string, 0, len(pairs))
	for attributeId, valueId := range pairs {
		if attributeId == "" || valueId == "" {
			continue
		}
		attributeIds = append(attributeIds, attributeId)
	}
	if len(attributeIds) == 0 {
		return models.EmptyCombinationKey
	}
	sort.Strings(attributeIds)

	parts := make([]string, 0, len(attributeIds))
	for _, attributeId := range attributeIds {
		parts = append(parts, attributeId+combinationPairSeparator+pairs[attributeId])
	}
	return strings.Join(parts, combinationSeparator)
}

// ParseKeyToPairs is the inverse of BuildKeyFromPairs.
func ParseKeyToPairs(key string) map[string]string {
	pairs := map[string]string{}
	for _, selection := range ParseCombinationKey(key) {
		pairs[selection.AttributeId] = selection.ValueId
	}
	return pairs
}

// ParseCombinationKey splits a key back into its attribute-value pairs, in key order. The empty key
// parses to no pairs rather than to an error.
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
