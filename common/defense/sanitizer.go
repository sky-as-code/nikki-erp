package defense

import (
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/sky-as-code/nikki-erp/common/util"
)

var ugcPolicy = bluemonday.UGCPolicy()
var strictPolicy = bluemonday.StrictPolicy()

// SanitizeRichText removes any dangerous HTML/JS from input
func SanitizeRichText(input string) string {
	return ugcPolicy.Sanitize(input)
}

// SanitizeRichTextPtr removes any dangerous HTML/JS from input
func SanitizeRichTextPtr(input *string) *string {
	if input == nil || len(*input) == 0 {
		return input
	}
	return util.ToPtr(ugcPolicy.Sanitize(*input))
}

// SanitizePlainText strips all HTML tags
func SanitizePlainText(input string, trimSpaces ...bool) string {
	if len(trimSpaces) > 0 && trimSpaces[0] {
		input = strings.TrimSpace(input)
	}
	return strictPolicy.Sanitize(input)
}

// SanitizePlainTextPtr strips all HTML tags
func SanitizePlainTextPtr(input *string, trimSpaces ...bool) *string {
	if input == nil || len(*input) == 0 {
		return input
	}
	if len(trimSpaces) > 0 && trimSpaces[0] {
		return util.ToPtr(strings.TrimSpace(*input))
	}
	return util.ToPtr(strictPolicy.Sanitize(*input))
}
