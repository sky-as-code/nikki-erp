package services

import (
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// NormalizeAttributeValueName is the comparison key of a value name in one locale: trimmed,
// NFC-normalized and case-folded, so "Đỏ", " đỏ" and a decomposed "Đỏ" collide. It is never
// stored; only the trimmed name is.
func NormalizeAttributeValueName(name string) string {
	folded := cases.Fold().String(norm.NFC.String(strings.TrimSpace(name)))
	return norm.NFC.String(folded)
}

// attributeValueName is one sibling as the name rule sees it.
type attributeValueName struct {
	id   model.Id
	name model.LangJson
}

// nameClash is a locale in which a candidate's name equals a sibling's.
type nameClash struct {
	locale model.LanguageCode
	name   string
}

// findNameClash compares the candidate against every sibling in each locale the candidate names.
// The candidate itself (same id) is skipped. Locales are walked in sorted order so the reported
// clash is deterministic.
func findNameClash(candidate attributeValueName, siblings []attributeValueName) *nameClash {
	for _, locale := range sortedLocales(candidate.name) {
		key := NormalizeAttributeValueName(candidate.name[locale])
		if key == "" {
			continue
		}
		for _, sibling := range siblings {
			if candidate.id != "" && sibling.id == candidate.id {
				continue
			}
			if NormalizeAttributeValueName(sibling.name[locale]) == key {
				return &nameClash{locale: locale, name: strings.TrimSpace(candidate.name[locale])}
			}
		}
	}
	return nil
}

// isLangJsonBlank is true when no locale carries text, an empty map included.
func isLangJsonBlank(name model.LangJson) bool {
	for _, text := range name {
		if strings.TrimSpace(text) != "" {
			return false
		}
	}
	return true
}

func newValueNameBlankError(field string) *ft.ClientErrorItem {
	return ft.NewValidationError(field, ft.ErrorKey("product_attribute_value.name_blank", "inventory"),
		"the value name cannot be empty; enter it in at least one language")
}

func sortedLocales(name model.LangJson) []model.LanguageCode {
	locales := make([]model.LanguageCode, 0, len(name))
	for locale := range name {
		locales = append(locales, locale)
	}
	sort.Slice(locales, func(i, j int) bool { return locales[i] < locales[j] })
	return locales
}

// trimLangJsonInPlace trims every locale of a langjson field, keeping whichever shape the field
// arrived in: a JSON body decodes to map[string]any, a Go caller passes model.LangJson.
func trimLangJsonInPlace(fields dmodel.DynamicFields, key string) {
	switch typed := fields[key].(type) {
	case map[string]any:
		for locale, raw := range typed {
			if text, ok := raw.(string); ok {
				typed[locale] = strings.TrimSpace(text)
			}
		}
	case model.LangJson:
		for locale, text := range typed {
			typed[locale] = strings.TrimSpace(text)
		}
	case *model.LangJson:
		if typed != nil {
			trimLangJsonInPlace(dmodel.DynamicFields{key: *typed}, key)
		}
	}
}
