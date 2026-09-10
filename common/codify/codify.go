// Package codify derives a machine code from a human label: "Nước ngọt có ga" → "nuoc_ngot_co_ga".
// It is the rule behind fields such as a category's code when only the name is given.
package codify

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// DefaultMaxLength keeps a generated code within the widest column the callers declare.
const DefaultMaxLength = 50

// vietnameseAscii maps the letters NFD decomposition does not reduce to ASCII.
var vietnameseAscii = map[rune]string{
	'đ': "d", 'Đ': "d",
}

// Codify lowercases, strips diacritics, turns every run of non-alphanumerics into one "_" and
// trims it from both ends. The result is at most maxLength runes, cut on a "_" when one is near
// the limit so a word is not sliced mid-way. Empty input gives an empty code.
func Codify(text string, maxLength ...int) string {
	limit := DefaultMaxLength
	if len(maxLength) > 0 && maxLength[0] > 0 {
		limit = maxLength[0]
	}
	ascii := toAscii(text)
	var builder strings.Builder
	pendingSeparator := false
	for _, r := range ascii {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if pendingSeparator && builder.Len() > 0 {
				builder.WriteByte('_')
			}
			pendingSeparator = false
			builder.WriteRune(unicode.ToLower(r))
			continue
		}
		pendingSeparator = true
	}
	return truncate(builder.String(), limit)
}

func toAscii(text string) string {
	var mapped strings.Builder
	for _, r := range text {
		if replacement, ok := vietnameseAscii[r]; ok {
			mapped.WriteString(replacement)
			continue
		}
		mapped.WriteRune(r)
	}
	stripMarks := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(stripMarks, mapped.String())
	if err != nil {
		return mapped.String()
	}
	return result
}

func truncate(code string, limit int) string {
	if len(code) <= limit {
		return code
	}
	cut := code[:limit]
	if idx := strings.LastIndexByte(cut, '_'); idx >= limit/2 {
		cut = cut[:idx]
	}
	return strings.TrimRight(cut, "_")
}
