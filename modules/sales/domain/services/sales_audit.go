package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The document audit trail.
//
// Every state-changing action writes exactly one event, through this file and nowhere else; a
// second writer would eventually record a different set of fields. The write deliberately runs in
// the caller's transaction rather than its own, so the event and the transition it records commit
// together. This is the document trail only: why a number is what it is lives in
// sales_order_adjustments, and the two are separate because an adjustment is replaced when the
// order is repriced while an event never is.

// SalesAuditEntry is one thing that happened to one record.
type SalesAuditEntry struct {
	// SalesOrderId is present even when the event is about a child record such as a bill or a
	// return, so the whole history of a sale is one query.
	SalesOrderId string

	// EntityType is the schema name of what actually changed, so an order's trail and its bills'
	// and returns' can share a table.
	EntityType string
	EntityId   string
	Action     string

	// Empty for an action that changes no status; writing the unchanged status into both would
	// suggest a transition that did not occur. Which of the order's four status dimensions they
	// refer to is implied by the action, because an action moves exactly one of them.
	FromStatus string
	ToStatus   string

	// Required by the operations that demand one (cancelling a confirmed order, overriding a
	// price) and enforced by those operations rather than here, since most actions need none.
	Reason string

	// Metadata carries whatever else is worth keeping about this action. Free-form because the
	// interesting detail differs per action.
	Metadata map[string]any

	OrgId string
}

// WriteSalesAuditEvent records one entry, in the caller's transaction.
//
// The actor comes from the context rather than the caller, so an operation cannot record somebody
// else as having performed it. It is left empty when the context carries no user: a kiosk sale, a
// gateway callback and a scheduled expiry genuinely have no actor.
func WriteSalesAuditEvent(ctx corectx.Context, entry SalesAuditEntry) error {
	engine, err := engineFor(models.SalesOrderEventSchemaName)
	if err != nil {
		return err
	}

	id, err := model.NewId()
	if err != nil {
		return errors.Wrap(err, "WriteSalesAuditEvent")
	}

	event := dmodel.DynamicFields{
		models.SalesOrderEventFieldId:           string(*id),
		models.SalesOrderEventFieldSalesOrderId: entry.SalesOrderId,
		models.SalesOrderEventFieldEntityType:   entry.EntityType,
		models.SalesOrderEventFieldEntityId:     entry.EntityId,
		models.SalesOrderEventFieldAction:       entry.Action,
	}
	if actorId := salesActorOf(ctx); actorId != "" {
		event[models.SalesOrderEventFieldActorId] = actorId
	}
	if entry.FromStatus != "" {
		event[models.SalesOrderEventFieldFromStatus] = entry.FromStatus
	}
	if entry.ToStatus != "" {
		event[models.SalesOrderEventFieldToStatus] = entry.ToStatus
	}
	if entry.Reason != "" {
		event[models.SalesOrderEventFieldReason] = entry.Reason
	}
	if len(entry.Metadata) > 0 {
		event[models.SalesOrderEventFieldMetadata] = entry.Metadata
	}
	if entry.OrgId != "" {
		event[basemodel.FieldOrgId] = entry.OrgId
	}

	// The insert goes through the repository rather than the resource service: the resource is
	// read-only to clients (its IAM seed grants read alone) and must stay that way.
	_, err = engine.ResourceRepository().Insert(ctx, event)
	return errors.Wrap(err, "WriteSalesAuditEvent")
}

// WriteOrderStatusEvent is the common case: one of a sales order's own statuses moved. It exists
// so the dozen callers recording an order transition do not each repeat the entity type and id,
// which is where a copy-paste error would file an event against the wrong record.
func WriteOrderStatusEvent(
	ctx corectx.Context, orderId, action, fromStatus, toStatus, orgId string,
) error {
	return WriteSalesAuditEvent(ctx, SalesAuditEntry{
		SalesOrderId: orderId,
		EntityType:   models.SalesOrderSchemaName,
		EntityId:     orderId,
		Action:       action,
		FromStatus:   fromStatus,
		ToStatus:     toStatus,
		OrgId:        orgId,
	})
}

// salesActorOf returns the id of the user the request is running as, or "" when there is none.
func salesActorOf(ctx corectx.Context) string {
	if ctx == nil {
		return ""
	}
	return string(ctx.GetPermissions().UserId)
}
