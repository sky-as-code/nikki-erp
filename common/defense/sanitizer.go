package defense

import (
	"github.com/microcosm-cc/bluemonday"
)

var ugcPolicy = bluemonday.UGCPolicy()
var strictPolicy = bluemonday.StrictPolicy()

// SanitizeRichText removes any dangerous HTML/JS from input
func SanitizeRichText(input string) string {
	return ugcPolicy.Sanitize(input)
}

// SanitizePlainText strips all HTML tags
func SanitizePlainText(input string) string {
	return strictPolicy.Sanitize(input)
}
