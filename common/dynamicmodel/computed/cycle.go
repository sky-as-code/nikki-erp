package computed

import (
	"strings"
)

// FieldRef names one field on one schema, the unit of dependency tracking.
type FieldRef struct {
	Schema string
	Field  string
}

func (this FieldRef) String() string {
	return this.Schema + "." + this.Field
}

// CycleError reports a computed-field dependency cycle. Detection happens during schema
// finalization — a cyclic schema set never boots, so evaluation can rely on the graph being a
// DAG (spec §18).
type CycleError struct {
	// Chain is the dependency path with the repeated field at both ends, e.g. A, B, A.
	Chain []FieldRef
}

func (this *CycleError) Error() string {
	names := make([]string, len(this.Chain))
	for i, ref := range this.Chain {
		names[i] = ref.String()
	}
	return "Computed field dependency cycle detected: " + strings.Join(names, " -> ")
}

// newCycleError builds the error from the in-flight resolution stack plus the field that closed
// the loop. The chain starts at the repeated field so the message reads A -> B -> A.
func newCycleError(stack []FieldRef, repeated FieldRef) *CycleError {
	start := 0
	for i, ref := range stack {
		if ref == repeated {
			start = i
			break
		}
	}
	chain := append([]FieldRef{}, stack[start:]...)
	chain = append(chain, repeated)
	return &CycleError{Chain: chain}
}
