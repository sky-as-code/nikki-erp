package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The audit trail. Every state-changing action writes exactly one event, through this file and
// nowhere else, so a blank field always means the same thing. The write deliberately runs in the
// caller's transaction rather than its own, so the event and the transition it records commit
// together or not at all.

// Audit action names. They record the operation the user asked for, not the transition it
// produced, since one operation can end in either of two statuses.
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

	// AuditActionOverridePrice records a line priced differently from what the vendor quotes. It is
	// a line action where every action above is an order action, hence AuditEntry.EntityType.
	AuditActionOverridePrice = "override_price"

	// AuditActionReprice records one line re-resolved from the vendor's current price list. Also a
	// line action: repricing an order writes one event per line, so a reader sees which line moved.
	AuditActionReprice = "reprice"
)

// AuditEntry is one thing that happened to one record.
type AuditEntry struct {
	// EntityType is the schema name of what changed, so the order's trail and the agreement's can
	// share one table.
	EntityType string
	EntityId   string
	Action     string

	// Empty for an action that changes no status, such as locking; writing the unchanged status
	// into both would suggest a transition that did not occur.
	FromStatus string
	ToStatus   string

	// Reason is required by unlock and optional everywhere else.
	Reason string

	// Metadata is free-form because the interesting detail differs per action.
	Metadata map[string]any

	OrgId string
}

// WriteAuditEvent records one entry in the caller's transaction. The actor comes from the context,
// not the caller, so an operation cannot record somebody else as having performed it; it stays
// empty for a system-initiated transition, which genuinely has no actor.
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

	// Through the repository, not the resource service: the service guard refuses every client
	// write to this resource and must keep doing so while the system writes its own events.
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
