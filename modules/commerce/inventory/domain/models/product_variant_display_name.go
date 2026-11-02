package models

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sky-as-code/nikki-erp/common/model"
)

// FormatVariantDisplayName renders "Coca Cola (No sugar, Big)": the template name, then the
// variant's value names in the template's line order. It is the one spelling shared by the
// variant's stored display_name, the effective product and the configuration preview, so a preview
// row reads exactly as the variant it describes will.
//
// A variant with no values is the ordinary single-variant case, not a defect, so it is named by
// its template alone rather than gaining an empty pair of brackets.
func FormatVariantDisplayName(name string, valueNames []string) string {
	kept := make([]string, 0, len(valueNames))
	for _, value := range valueNames {
		if value != "" {
			kept = append(kept, value)
		}
	}
	if len(kept) == 0 {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, strings.Join(kept, ", "))
}

// BuildVariantDisplayName renders the stored display_name in every locale. Each name is first
// normalized to the locales, so a value translated into fewer languages than the template borrows
// its neighbour's text the way a submitted LangJson does. locales nil means the enabled set is
// unknown, and the template name's own locales are used. fallback names the variant where the
// template has no name (its SKU).
func BuildVariantDisplayName(
	templateName model.LangJson, valueNames []model.LangJson, locales []model.LanguageCode, fallback string,
) model.LangJson {
	if len(locales) == 0 {
		locales = langJsonLocales(templateName)
	}
	name := templateName.Normalize(locales)
	values := make([]model.LangJson, len(valueNames))
	for i, value := range valueNames {
		values[i] = value.Normalize(locales)
	}
	out := model.LangJson{}
	for _, locale := range locales {
		base := name[locale]
		if base == "" {
			base = fallback
		}
		texts := make([]string, 0, len(values))
		for _, value := range values {
			texts = append(texts, value[locale])
		}
		if text := FormatVariantDisplayName(base, texts); text != "" {
			out[locale] = text
		}
	}
	return out
}

func langJsonLocales(lang model.LangJson) []model.LanguageCode {
	locales := make([]model.LanguageCode, 0, len(lang))
	for code := range lang {
		if code != model.LanguageCodeRef {
			locales = append(locales, code)
		}
	}
	if len(locales) == 0 {
		return []model.LanguageCode{model.DefaultLanguageCode}
	}
	sort.Strings(locales)
	return locales
}
