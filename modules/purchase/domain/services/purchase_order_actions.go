package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The order operations that are not lifecycle transitions: send, duplicate, lock, unlock and
// acknowledge. Each carries its own permission because they are materially different powers;
// folding them into `update` would let a role that may fix a typo also send an RFQ to a vendor.
// Print is deliberately absent: it changes no server state and needs nothing beyond `read`.

// Send records that a request for quotation went out to the vendor. It only changes status — this
// module does not send email. Only an RFQ can be sent; a confirmed order has already been bought.
func (this *PurchaseOrderDomainServiceImpl) Send(
	ctx corectx.Context, orderId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.transition(ctx, orderId, transitionRequest{
		Action: AuditActionSend,
		To:     string(models.PurchaseOrderStatusRfqSent),
		AllowedFrom: map[string]bool{
			string(models.PurchaseOrderStatusRfq): true,
		},
		RefusalKey: "purchase_order.not_sendable",
		RefusalMessage: "only a draft request for quotation can be sent to a vendor; " +
			"this one is '%s'",
	})
}

// Lock closes an order to further editing. Locking is a boolean and never a status: an order is
// locked and confirmed, not locked instead of confirmed, which is what lets auto_lock work without
// a status the transition table would have to route around. Only a committed order can be locked.
func (this *PurchaseOrderDomainServiceImpl) Lock(
	ctx corectx.Context, orderId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setLocked(ctx, orderId, true, "", AuditActionLock)
}

// Unlock reopens a locked order for editing. The reason is mandatory here, unlike on cancel:
// unlocking undoes a deliberately applied control, so the audit trail has to say why.
func (this *PurchaseOrderDomainServiceImpl) Unlock(
	ctx corectx.Context, orderId string, reason string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if reason == "" {
		return orderViolationResult("purchase_order.unlock_reason_required",
			"unlocking a purchase order requires a reason"), nil
	}
	return this.setLocked(ctx, orderId, false, reason, AuditActionUnlock)
}

// Acknowledge records that the vendor confirmed receipt. Like is_locked it is a flag, not a status:
// an order is confirmed whether or not the vendor has acknowledged it. Only a committed order can be
// acknowledged.
func (this *PurchaseOrderDomainServiceImpl) Acknowledge(
	ctx corectx.Context, orderId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		order, err := loadOrder(tranxCtx, orderId)
		if err != nil {
			return err
		}
		if order == nil {
			result = orderNotFoundResult(orderId)
			return nil
		}

		status := stringOf(order, models.PurchaseOrderFieldStatus)
		if !IsOrderCommitted(status) {
			result = orderViolationResult("purchase_order.not_acknowledgeable",
				"only a confirmed purchase order can be acknowledged; this one is '"+status+"'")
			return nil
		}
		if boolOf(order, models.PurchaseOrderFieldVendorAcknowledged) {
			// Already acknowledged: a retry is not an error, and writing a second audit event
			// would suggest the vendor acknowledged twice.
			result = mutateOk()
			return nil
		}

		if err := writeOrderChanges(tranxCtx, order, dmodel.DynamicFields{
			models.PurchaseOrderFieldVendorAcknowledged: true,
		}); err != nil {
			return err
		}

		result = mutateOk()
		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   orderId,
			Action:     AuditActionAcknowledge,
			OrgId:      stringOf(order, basemodel.FieldOrgId),
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Duplicate copies an order and its lines into a new draft RFQ with a fresh code and none of the
// original's history — no confirmed_at, approval, acknowledgement or lock — which would otherwise
// claim approval by someone who never saw the document. It is also how a cancelled order is revived.
func (this *PurchaseOrderDomainServiceImpl) Duplicate(
	ctx corectx.Context, orderId string,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	var result *dyn.OpResult[dmodel.DynamicFields]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		order, err := loadOrder(tranxCtx, orderId)
		if err != nil {
			return err
		}
		if order == nil {
			notFound := orderNotFoundResult(orderId)
			result = &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: notFound.ClientErrors}
			return nil
		}

		created, err := this.Create(tranxCtx, copyableOrderFields(order))
		if err != nil || created.ClientErrors.Count() > 0 {
			result = created
			return err
		}
		result = created

		newOrderId := stringOf(created.Data, models.PurchaseOrderFieldId)
		if err := this.copyOrderLines(tranxCtx, orderId, newOrderId); err != nil {
			return err
		}
		if err := RecomputeOrderTotals(tranxCtx, newOrderId); err != nil {
			return err
		}

		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   newOrderId,
			Action:     AuditActionDuplicate,
			ToStatus:   string(models.PurchaseOrderStatusRfq),
			OrgId:      stringOf(order, basemodel.FieldOrgId),
			Metadata:   map[string]any{"duplicated_from": orderId},
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// copyableOrderFields is the order's terms without its history. It is deliberately an allowlist, so
// a field added to the schema later defaults to not being copied.
func copyableOrderFields(order dmodel.DynamicFields) dmodel.DynamicFields {
	copied := dmodel.DynamicFields{}
	for _, field := range []string{
		models.PurchaseOrderFieldVendorId,
		models.PurchaseOrderFieldBuyerId,
		models.PurchaseOrderFieldCurrencyId,
		models.PurchaseOrderFieldVendorReference,
		models.PurchaseOrderFieldSourceReference,
		models.PurchaseOrderFieldExpectedArrival,
		models.PurchaseOrderFieldPriority,
		models.PurchaseOrderFieldTermsConditions,
		models.PurchaseOrderFieldAgreementId,
		basemodel.FieldOrgId,
	} {
		if value, ok := order[field]; ok && value != nil {
			copied[field] = value
		}
	}
	// order_deadline is deliberately not copied: it is already past by the time anyone duplicates.
	// sourcing_group_id is not copied either — a duplicate is a new requirement, not another
	// alternative for the one being compared.
	return copied
}

// copyOrderLines copies every line of one order onto another.
func (this *PurchaseOrderDomainServiceImpl) copyOrderLines(
	ctx corectx.Context, fromOrderId string, toOrderId string,
) error {
	lineEngine, err := engineFor(models.PurchaseOrderLineSchemaName)
	if err != nil {
		return err
	}
	lines, err := models.FindOrderLines(
		ctx, lineEngine.ResourceRepository(), fromOrderId, models.MaxOrderLines)
	if err != nil {
		return err
	}

	for _, line := range lines {
		copied := dmodel.DynamicFields{
			models.PurchaseOrderLineFieldPurchaseOrderId: toOrderId,
		}
		for _, field := range []string{
			models.PurchaseOrderLineFieldSequence,
			models.PurchaseOrderLineFieldLineType,
			models.PurchaseOrderLineFieldProductVariantId,
			models.PurchaseOrderLineFieldDescription,
			models.PurchaseOrderLineFieldQuantity,
			models.PurchaseOrderLineFieldUomId,
			models.PurchaseOrderLineFieldInventoryQuantity,
			models.PurchaseOrderLineFieldUnitPrice,
			models.PurchaseOrderLineFieldDiscountPercent,
			models.PurchaseOrderLineFieldTaxAmount,
			basemodel.FieldOrgId,
		} {
			if value, ok := line[field]; ok && value != nil {
				copied[field] = value
			}
		}
		StampLineTotals(copied)

		if _, err := lineEngine.ResourceService().Create(ctx, copied); err != nil {
			return err
		}
	}
	return nil
}

// setLocked is the shared body of Lock and Unlock.
func (this *PurchaseOrderDomainServiceImpl) setLocked(
	ctx corectx.Context, orderId string, locked bool, reason string, action string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		order, err := loadOrder(tranxCtx, orderId)
		if err != nil {
			return err
		}
		if order == nil {
			result = orderNotFoundResult(orderId)
			return nil
		}

		status := stringOf(order, models.PurchaseOrderFieldStatus)
		if !IsOrderCommitted(status) {
			result = orderViolationResult("purchase_order.not_lockable",
				"only a confirmed purchase order can be locked or unlocked; this one is '"+status+"'")
			return nil
		}
		if boolOf(order, models.PurchaseOrderFieldIsLocked) == locked {
			// Already in the requested state: a retry is not an error.
			result = mutateOk()
			return nil
		}

		if err := writeOrderChanges(tranxCtx, order, dmodel.DynamicFields{
			models.PurchaseOrderFieldIsLocked: locked,
		}); err != nil {
			return err
		}

		result = mutateOk()
		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   orderId,
			Action:     action,
			Reason:     reason,
			OrgId:      stringOf(order, basemodel.FieldOrgId),
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// transitionRequest describes a straightforward status change: one target, a set of statuses it may
// be reached from, and what to say when it cannot.
type transitionRequest struct {
	Action         string
	To             string
	AllowedFrom    map[string]bool
	RefusalKey     string
	RefusalMessage string

	// Stamp is applied alongside the status, for a transition that records something else too.
	Stamp func(now time.Time) dmodel.DynamicFields
}

// transition is the shared body of the simple status changes. Confirm, Approve and Cancel do not
// use it: each reads configuration, stamps evidence or picks its own target.
func (this *PurchaseOrderDomainServiceImpl) transition(
	ctx corectx.Context, orderId string, request transitionRequest,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		order, err := loadOrder(tranxCtx, orderId)
		if err != nil {
			return err
		}
		if order == nil {
			result = orderNotFoundResult(orderId)
			return nil
		}

		status := stringOf(order, models.PurchaseOrderFieldStatus)
		if status == request.To {
			result = mutateOk()
			return nil
		}
		if !request.AllowedFrom[status] {
			result = orderViolationResult(request.RefusalKey,
				formatStatus(request.RefusalMessage, status))
			return nil
		}

		changes := dmodel.DynamicFields{models.PurchaseOrderFieldStatus: request.To}
		if request.Stamp != nil {
			for key, value := range request.Stamp(time.Now()) {
				changes[key] = value
			}
		}
		if err := writeOrderChanges(tranxCtx, order, changes); err != nil {
			return err
		}

		result = mutateOk()
		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   orderId,
			Action:     request.Action,
			FromStatus: status,
			ToStatus:   request.To,
			OrgId:      stringOf(order, basemodel.FieldOrgId),
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// formatStatus substitutes the one %s a refusal message may carry, without pulling in fmt for a
// single verb.
func formatStatus(message, status string) string {
	for index := 0; index+1 < len(message); index++ {
		if message[index] == '%' && message[index+1] == 's' {
			return message[:index] + status + message[index+2:]
		}
	}
	return message
}
