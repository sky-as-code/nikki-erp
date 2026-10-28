package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// RecordAudit is for a custom write that goes neither through the default domain service nor
// through Repository().Insert/Update/DeleteOne, such as a corecrud.UpdateRegardless caller. Call
// it inside the write's transaction with the row before and the fields written: a nil before is a
// create, a nil after a delete. An update that changed no audited field records nothing.
func RecordAudit(
	ctx corectx.Context, schema *dmodel.ModelSchema, action string, before, after dmodel.DynamicFields,
) error {
	if schema == nil || !IsAuditedSchema(schema.Name()) || !sharedAuditSink.Enabled() {
		return nil
	}
	entry := newAuditEntry(schema, action, before, after)
	if before != nil && after != nil && entry.IsEmpty() {
		return nil
	}
	err := sharedAuditSink.Record(ctx, []audit.Entry{entry})
	return errors.Wrapf(err, "record %s %s", action, schema.Name())
}
