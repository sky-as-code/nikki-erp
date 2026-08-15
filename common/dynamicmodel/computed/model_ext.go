package computed

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// DefOf is the typed accessor for a field's computed definition. ModelField stores the raw
// expression untyped (this package imports model, so model cannot import this package back);
// DefOf recovers the typed view and derives the definition's kind. It returns (nil, nil) for a
// field that is not computed.
func DefOf(field *dmodel.ModelField) (*Definition, error) {
	if !field.IsComputed() {
		return nil, nil
	}
	expr, ok := field.RawComputedExpr().(Expr)
	if !ok {
		return nil, errors.Errorf(
			"field %q: computed expression is a %T, not a computed.Expr — pass a value built with "+
				"the computed package's constructors", field.Name(), field.RawComputedExpr())
	}
	def, err := NewDefinition(field.ComputedIsStored(), expr)
	return def, errors.Wrapf(err, "field %q", field.Name())
}
