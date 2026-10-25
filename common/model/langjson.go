package model

import (
	"sort"

	"github.com/sky-as-code/nikki-erp/common/array"
)

// Normalize returns a copy in which every supported language holds a translation.
//
// Keys are canonicalized to BCP47 and those outside `enabled` are dropped; a nil `enabled` means
// the supported set is unknown, so nothing is dropped and only the keys present are filled. Codes
// are processed in sorted order, and an empty one copies the nearest filled code before it, else
// the nearest after it. Go maps and jsonb keep no key order, so the sort is a processing order only.
//
// A value referencing a translation key ($ref) or holding no key is returned as is, and one with no
// supported translation is only stripped: there is nothing to copy from, and `{}` must stay `{}` so
// a required check can still refuse it.
func (this LangJson) Normalize(enabled []LanguageCode) LangJson {
	if _, hasRef := this[LanguageCodeRef]; hasRef || len(this) == 0 {
		return this
	}
	allowed := canonicalLanguageCodes(enabled)
	present := this.canonicalTranslations(allowed, enabled != nil)
	if !hasAnyTranslation(present) {
		return present
	}
	codes := make([]LanguageCode, 0, len(present)+len(allowed))
	for code := range present {
		codes = append(codes, code)
	}
	for _, code := range allowed {
		if _, ok := present[code]; !ok {
			codes = append(codes, code)
		}
	}
	sort.Strings(codes)
	return fillFromNeighbours(codes, present)
}

func canonicalLanguageCodes(codes []LanguageCode) []LanguageCode {
	result := make([]LanguageCode, 0, len(codes))
	for _, code := range codes {
		canonical, err := ToBCP47LanguageCode(code)
		if err != nil || canonical == LanguageCodeRef || array.Contains(result, canonical) {
			continue
		}
		result = append(result, canonical)
	}
	return result
}

// canonicalTranslations keeps a key that fails canonicalization under its raw spelling when no
// supported set is known, so the schema validator still reports it instead of it vanishing here.
func (this LangJson) canonicalTranslations(allowed []LanguageCode, shouldStrip bool) LangJson {
	result := make(LangJson, len(this))
	for code, text := range this {
		canonical, err := ToBCP47LanguageCode(code)
		if err != nil {
			canonical = code
		}
		if shouldStrip && !array.Contains(allowed, canonical) {
			continue
		}
		if existing := result[canonical]; existing == "" {
			result[canonical] = text
		}
	}
	return result
}

func hasAnyTranslation(translations LangJson) bool {
	for _, text := range translations {
		if text != "" {
			return true
		}
	}
	return false
}

func fillFromNeighbours(sortedCodes []LanguageCode, translations LangJson) LangJson {
	result := make(LangJson, len(sortedCodes))
	lastFilled := ""
	for _, code := range sortedCodes {
		if text := translations[code]; text != "" {
			lastFilled = text
		}
		result[code] = lastFilled
	}
	firstFilled := ""
	for i := len(sortedCodes) - 1; i >= 0; i-- {
		code := sortedCodes[i]
		if text := translations[code]; text != "" {
			firstFilled = text
		}
		if result[code] == "" {
			result[code] = firstFilled
		}
	}
	return result
}
