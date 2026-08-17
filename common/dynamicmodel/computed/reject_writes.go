package computed

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// ErrKeyComputedFieldNotWritable identifies an explicit write to a computed field.
const ErrKeyComputedFieldNotWritable = "err_computed_field_not_writable"

// RejectWrites reports a client error for every computed field the input tries to write. The
// generic write path already silently strips column-less fields, so a stray computed value could
// never reach a column — this check exists to tell the client explicitly instead of letting the
// value quietly vanish (spec §25's recommendation).
//
// Edge-model fields are computed too, so echoing a hydrated edge back from a GET into a write is
// reported rather than ignored, on the same reasoning: a value the server will discard is better
// named than silently dropped.
func RejectWrites(schema *dmodel.ModelSchema, input dmodel.DynamicFields) ft.ClientErrors {
	var errs ft.ClientErrors
	for name := range input {
		field, ok := schema.Field(name)
		if !ok || !field.IsComputed() {
			continue
		}
		errs.Append(*ft.NewValidationError(name,
			ft.ErrorKey(ErrKeyComputedFieldNotWritable),
			fmt.Sprintf("Field %q is computed and cannot be written", name)))
	}
	return errs
}
