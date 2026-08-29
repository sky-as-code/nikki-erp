package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Two of this module's resources are not client-writable. The action stays defined and refuses
// rather than being removed, so the caller gets a 400 naming the reason instead of a 404 that reads
// as a wrong URL.

// defineAuditEventGuards closes the audit trail to clients. An audit event is written only by the
// audit helper, inside the same transaction as the transition it records; a client-written one
// would be indistinguishable from a real event. Update and delete are refused because the record is
// immutable — the schema extends the readonly auditable mixin, so there is no updated_at.
func defineAuditEventGuards(engine drif.DynamicResourceEngine) error {
	return attachWriteGuards(engine, models.AuditEventSchemaName, rejectAuditEventWrite)
}

func rejectAuditEventWrite(
	_ corectx.Context, _ *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	vErrs.Append(*ft.NewBusinessViolation(
		models.AuditEventSchemaName,
		"purchase_audit_event.not_client_writable",
		"audit events are written by the system when a purchase order or agreement changes state; "+
			"they cannot be created, edited or removed",
	))
	return nil
}

// defineSourcingGroupGuards closes direct creation and deletion of a sourcing group. The group is
// created by create_alternative and reaped when it drops to one order or fewer, so a hand-made one
// would be an empty container nothing reaps and a hand-deleted one would strand its orders. Update
// stays open: the group carries no fields of its own beyond the base ones.
func defineSourcingGroupGuards(engine drif.DynamicResourceEngine) error {
	return attachWriteGuards(engine, models.SourcingGroupSchemaName, rejectSourcingGroupWrite,
		drif.ActionCreate, drif.ActionDelete)
}

func rejectSourcingGroupWrite(
	_ corectx.Context, _ *drif.DynamicEntity, _ *drif.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	vErrs.Append(*ft.NewBusinessViolation(
		models.SourcingGroupSchemaName,
		"purchase_sourcing_group.not_client_writable",
		"sourcing groups are created by adding an alternative to a purchase order and removed when "+
			"fewer than two alternatives remain",
	))
	return nil
}

// attachWriteGuards hangs the same refusal on each named action, defaulting to all three write
// verbs when none is named.
func attachWriteGuards(
	engine drif.DynamicResourceEngine,
	schemaName string,
	guard drif.ActionValidateExtraFn,
	actions ...string,
) error {
	if len(actions) == 0 {
		actions = []string{drif.ActionCreate, drif.ActionUpdate, drif.ActionDelete}
	}

	for _, action := range actions {
		err := engine.ModifyAction(drif.DynamicActionDelta{
			ActionName:    action,
			ValidateExtra: guard,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to attach the '%s' %s guard", schemaName, action)
		}
	}
	return nil
}
