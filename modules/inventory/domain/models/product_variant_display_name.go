package models

import (
	"fmt"
	"strings"
)

// FormatVariantDisplayName renders "Coca Cola (No sugar, Big)": the template name, then the
// variant's value names in the template's line order. It is the one spelling shared by the
// variant's computed display_name and the configuration preview, so a preview row reads exactly as
// the variant it describes will.
//
// A variant with no values is the ordinary single-variant case, not a defect, so it is named by
// its template alone rather than gaining an empty pair of brackets.
func FormatVariantDisplayName(name string, valueNames []string) string {
	if len(valueNames) == 0 {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, strings.Join(valueNames, ", "))
}
