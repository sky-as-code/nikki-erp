package model

// ComputedJsonParser turns a field's raw "computed" JSON block into the expression value that
// FieldBuilder.Computed accepts, plus the declared is_stored flag.
//
// The parser lives in common/dynamicmodel/computed, which imports this package — so this package
// cannot import it back. Instead the computed package registers its parser here at init time,
// and buildFieldFromDto calls through this seam. A schema that declares "computed" without the
// computed package linked in fails loudly rather than silently dropping the definition.
type ComputedJsonParser func(raw []byte, fieldName string) (expression any, isStored bool, err error)

var computedJsonParser ComputedJsonParser

// RegisterComputedJsonParser installs the parser. Called once, from the computed package's init.
func RegisterComputedJsonParser(parser ComputedJsonParser) {
	computedJsonParser = parser
}

// ComputedFinalizer validates every computed field once all schemas are registered and relations
// are resolved. FinalizeRelations invokes it AFTER releasing the registry lock, so the finalizer
// may use the registry's ordinary (read-locking) accessors.
type ComputedFinalizer func(reg *SchemaRegistry) error

var computedFinalizer ComputedFinalizer

// RegisterComputedFinalizer installs the finalizer. Called once, from the computed package's init.
func RegisterComputedFinalizer(finalizer ComputedFinalizer) {
	computedFinalizer = finalizer
}

// ComputedDescriptor is the client-facing summary of a computed field: enough for a form to know
// how the value arrives and what to watch, never the expression tree itself.
type ComputedDescriptor struct {
	// Kind is the computed kind, e.g. "expression", "related", "function".
	Kind string `json:"kind,omitempty"`
	// DependsOn names the same-schema field a function-kind computation reads, when it declares
	// one. A form recomputes the field through meta/compute the moment that field changes.
	DependsOn string `json:"depends_on,omitempty"`
}

// ComputedDescriberFn summarizes a field's raw computed expression for ToSimplized. Registered
// through the same seam as the parser and finalizer, and for the same reason: only the computed
// package can read its own expression types.
type ComputedDescriberFn func(expression any) *ComputedDescriptor

var computedDescriber ComputedDescriberFn

// RegisterComputedDescriber installs the describer. Called once, from the computed package's init.
func RegisterComputedDescriber(describer ComputedDescriberFn) {
	computedDescriber = describer
}

// computedDescriptor summarizes this field's computed definition, or nil when it has none or the
// computed package is not linked in.
func (this ModelField) computedDescriptor() *ComputedDescriptor {
	// An edge-model field is computed in the sense that its value is derived, but it carries no
	// expression to describe.
	if !this.IsComputed() || this.IsEdgeModel() || computedDescriber == nil {
		return nil
	}
	return computedDescriber(this.RawComputedExpr())
}
