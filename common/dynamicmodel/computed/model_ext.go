package computed

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// DefOf is the typed accessor for a field's computed definition. ModelField stores the raw
// expression untyped (this package imports model, so model cannot import this package back);
// DefOf recovers the typed view and derives the definition's kind. It returns (nil, nil) for a
// field that is not computed.
//
// An edge-model field is computed in the sense that its value is derived rather than supplied,
// but it carries no expression — the repository hydrates it from the peer schema. Excluded
// here so the type assertion below cannot trip over its nil expression.
func DefOf(field *dmodel.ModelField) (*Definition, error) {
	if !field.IsComputed() || field.IsEdgeModel() {
		return nil, nil
	}
	raw := field.RawComputedExpr()
	// A chained GoFunction(...) call yields a builder, not the node: unwrap it here so the builder
	// never has to implement Expr and thus can never be passed as an operand.
	if builder, isBuilder := raw.(*GoFunctionBuilder); isBuilder {
		raw = builder.Build()
	}
	expr, ok := raw.(Expr)
	if !ok {
		return nil, errors.Errorf(
			"field %q: computed expression is a %T, not a computed.Expr — pass a value built with "+
				"the computed package's constructors", field.Name(), raw)
	}
	def, err := NewDefinition(field.IsPersisted(), expr)
	return def, errors.Wrapf(err, "field %q", field.Name())
}

func init() {
	// The client-facing summary of a definition. Registered through the same seam as the parser
	// and finalizer: only this package can read its own expression types.
	dmodel.RegisterComputedDescriber(func(expression any) *dmodel.ComputedDescriptor {
		expr, ok := expression.(Expr)
		if !ok {
			if builder, isBuilder := expression.(*GoFunctionBuilder); isBuilder {
				expr = builder.Build()
			} else {
				return nil
			}
		}
		def, err := NewDefinition(false, expr)
		if err != nil || def == nil {
			// A malformed definition is reported at finalize time with full context; a schema
			// response is the wrong place to surface it a second time.
			return nil
		}
		descriptor := &dmodel.ComputedDescriptor{Kind: string(def.Kind)}
		if def.Function != nil {
			descriptor.DependsOn = def.Function.DependsOn
		}
		return descriptor
	})
}
