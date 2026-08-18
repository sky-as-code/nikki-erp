package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The audit trail (PUR-R6).
//
// Every state-changing action writes exactly one event, through this file and nowhere else. One
// writer is what makes the trail trustworthy: a second path would eventually record a different
// set of fields, and a reader comparing two events could no longer tell whether a blank actor
// means "the system did it" or "that writer forgot to fill it in".
//
// The write is deliberately NOT its own transaction. It runs in the caller's, so that the event and
// the transition it records commit together or not at all. An event that survived a rolled-back
// transition would be a record of something that did not happen.

// Audit action names. They are the operation the user asked for, not the transition it produced:
// "confirm" says what was done, where "rfq -> to_approve" only says what came of it, and the two
// are not the same when one operation can end in either of two statuses.
const (
	AuditActionConfirm     = "confirm"
	AuditActionApprove     = "approve"
	AuditActionCancel      = "cancel"
	AuditActionSend        = "send"
	AuditActionLock        = "lock"
	AuditActionUnlock      = "unlock"
	AuditActionAcknowledge = "acknowledge"
	AuditActionDuplicate   = "duplicate"
	AuditActionMerge       = "merge"
	AuditActionClose       = "close"
	AuditActionCreateRfq   = "create_rfq"
)

// AuditEntry is one thing that happened to one record.
type AuditEntry struct {
	// EntityType is the schema name of what changed, so that the order's trail and the
	// agreement's can share a table without a reader having to guess which is which.
	EntityType string
	EntityId   string
	Action     string

	// FromStatus and ToStatus are empty for an action that changes no status — locking an order
	// is a real event with nothing to record in those two columns, and writing the unchanged
	// status into both would suggest a transition that did not occur.
	FromStatus string
	ToStatus   string

	// Reason is required by unlock (BR §21) and optional everywhere else.
	Reason string

	// Metadata carries whatever else is worth keeping about this particular action. It is
	// free-form because the interesting detail differs per action, and a column per action would
	// be a table of mostly-null columns.
	Metadata map[string]any

	OrgId string
}

// WriteAuditEvent records one entry, in the caller's transaction.
//
// The actor comes from the context rather than from the caller, so that an operation cannot record
// somebody else as having performed it. It is left empty when the context carries no user — a
// system-initiated transition genuinely has no actor, and inventing one would be worse than a blank.
func WriteAuditEvent(ctx corectx.Context, entry AuditEntry) error {
	engine, err := engineFor(models.AuditEventSchemaName)
	if err != nil {
		return err
	}

	id, err := model.NewId()
	if err != nil {
		return errors.Wrap(err, "WriteAuditEvent")
	}

	event := dmodel.DynamicFields{
		models.AuditEventFieldId:         *id,
		models.AuditEventFieldEntityType: entry.EntityType,
		models.AuditEventFieldEntityId:   entry.EntityId,
		models.AuditEventFieldAction:     entry.Action,
	}
	if actorId := actorOf(ctx); actorId != "" {
		event[models.AuditEventFieldActorId] = actorId
	}
	if entry.FromStatus != "" {
		event[models.AuditEventFieldFromStatus] = entry.FromStatus
	}
	if entry.ToStatus != "" {
		event[models.AuditEventFieldToStatus] = entry.ToStatus
	}
	if entry.Reason != "" {
		event[models.AuditEventFieldReason] = entry.Reason
	}
	if len(entry.Metadata) > 0 {
		event[models.AuditEventFieldMetadata] = entry.Metadata
	}
	if entry.OrgId != "" {
		event[basemodel.FieldOrgId] = entry.OrgId
	}

	// The insert goes through the repository rather than the resource service, because the
	// service is where the client-facing guard lives: defineAuditEventGuards refuses every write
	// to this resource, and it must keep refusing them for a client while still letting the
	// system write its own.
	_, err = engine.ResourceRepository().Insert(ctx, event)
	return errors.Wrap(err, "WriteAuditEvent")
}

// actorOf returns the id of the user the request is running as, or "" when there is none.
func actorOf(ctx corectx.Context) string {
	if ctx == nil {
		return ""
	}
	return string(ctx.GetPermissions().UserId)
}
