package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Two of this module's resources are not client-writable, and both close the write surface the same
// way: the action stays defined and refuses, rather than being removed.
//
// That distinction matters to whoever hits it. A removed action answers 404, which reads as a wrong
// URL and sends someone looking for a typo; a refused one answers 400 naming the reason and what to
// do instead. The inventory module settled on the same approach for its derived resources.

// defineAuditEventGuards closes the audit trail to clients.
//
// An audit event is written by the audit helper inside the same transaction as the transition it
// records, and by nothing else (PUR-R6). A client-written event would be a claim that something
// happened, sitting in the same table as the events that did happen, with no way for a reader to
// tell them apart — which destroys the value of the trail rather than adding to it.
//
// Update and delete are refused for the stronger reason that the record is immutable: an audit
// trail someone can edit is not an audit trail. The schema says the same thing by extending the
// readonly auditable mixin, so there is no updated_at for a change to be recorded in.
func defineAuditEventGuards(engine drif.DynamicResourceEngine) error {
	return attachWriteGuards(engine, models.AuditEventSchemaName, rejectAuditEventWrite)
}

func rejectAuditEventWrite(
	_ corectx.Context, _ dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	vErrs.Append(*ft.NewBusinessViolation(
		models.AuditEventSchemaName,
		"purchase_audit_event.not_client_writable",
		"audit events are written by the system when a purchase order or agreement changes state; "+
			"they cannot be created, edited or removed",
	))
	return nil
}

// defineSourcingGroupGuards closes direct creation and deletion of a sourcing group.
//
// The group is a technical record with no meaning of its own (BR 28): it exists to say that several
// orders are alternatives for the same requirement. It is created by the create_alternative action
// and reaped when it drops to one order or fewer, so a hand-made group would be an empty container
// that nothing reaps, and a hand-deleted one would strand the orders that pointed at it.
//
// Update is left open: the group carries no fields of its own beyond the base ones, so there is
// nothing meaningful to forbid, and refusing an update that changes nothing would only be noise.
func defineSourcingGroupGuards(engine drif.DynamicResourceEngine) error {
	return attachWriteGuards(engine, models.SourcingGroupSchemaName, rejectSourcingGroupWrite,
		drif.ActionCreate, drif.ActionDelete)
}

func rejectSourcingGroupWrite(
	_ corectx.Context, _ dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
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
