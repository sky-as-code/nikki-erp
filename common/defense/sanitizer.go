package defense

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var ugcPolicy = bluemonday.UGCPolicy()
var strictPolicy = bluemonday.StrictPolicy()

// SanitizeRichText removes any dangerous HTML/JS from input
func SanitizeRichText(input string) string {
	return ugcPolicy.Sanitize(input)
}

// SanitizePlainText strips all HTML tags
func SanitizePlainText(input string, trimSpaces ...bool) string {
	if len(trimSpaces) > 0 && trimSpaces[0] {
		input = strings.TrimSpace(input)
	}
	return strictPolicy.Sanitize(input)
}
