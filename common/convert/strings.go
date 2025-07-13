package convert

import (
	"regexp"
	"strings"
)

var (
	// Matches anything not a letter, number, or underscore
	invalidCharRe = regexp.MustCompile(`[^\p{L}\p{N}_]+`)
	// Matches multiple underscores
	multiUnderscoreRe = regexp.MustCompile(`_+`)
)

// SanitizeToSnakeIdentifier replaces non visible Unicode chars with "_",
// collapses multiple underscores, trims leading/trailing "_"
//
// Examples:
//
// "123abcDEF" → "123abcDEF"
//
// "Đây là tên!" → "Đây_là_tên"
//
// "user@domain.com" → "user_domain_com"
//
// "hello---world!" → "hello_world"
//
// "  😊emoji__test😡  " → "emoji_test"
//
// "__underscores__at__ends__" → "underscores_at_ends"
func ToUnicodeSnakeCase(input string) string {
	s := invalidCharRe.ReplaceAllString(input, "_")
	s = multiUnderscoreRe.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")

	return s
}
