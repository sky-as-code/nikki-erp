package template

import (
	"fmt"
	"strings"
)

const (
	openBrace  = "{{"
	closeBrace = "}}"

	// interpolateBuilderSlack is extra Builder capacity beyond len(message). Values substituted for
	// "{{key}}" are often longer than the placeholder, so a small fixed slack reduces reallocations.
	interpolateBuilderSlack = 32
)

// Interpolate replaces "{{key}}" segments in message using vars; keys are trimmed.
// Unknown keys leave the original "{{...}}" segment unchanged.
func Interpolate(message string, vars map[string]any) string {
	if len(vars) == 0 {
		return message
	}
	openLen := len(openBrace)
	closeLen := len(closeBrace)
	var b strings.Builder
	b.Grow(len(message) + interpolateBuilderSlack)
	pos := 0
	for pos < len(message) {
		openRel := strings.Index(message[pos:], openBrace)
		if openRel < 0 {
			b.WriteString(message[pos:])
			break
		}
		absOpen := pos + openRel
		b.WriteString(message[pos:absOpen])
		closeRel := strings.Index(message[absOpen+openLen:], closeBrace)
		if closeRel < 0 {
			b.WriteString(message[absOpen:])
			break
		}
		key := strings.TrimSpace(message[absOpen+openLen : absOpen+openLen+closeRel])
		if val, ok := vars[key]; ok {
			b.WriteString(fmt.Sprint(val))
		} else {
			b.WriteString(message[absOpen : absOpen+openLen+closeRel+closeLen])
		}
		pos = absOpen + openLen + closeRel + closeLen
	}
	return b.String()
}
