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
