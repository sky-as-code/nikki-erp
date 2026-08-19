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
// acknowledge (BR §13-§15, §20-§22).
//
// Each is its own permission, for the reason the IAM seed gives: they are materially different
// powers. Sending an RFQ puts the company's name in front of a vendor; unlocking reopens a document
// that was deliberately closed. Folding them into `update` would let a role that may fix a typo do
// both.
//
// Print is deliberately NOT here. It produces a document from data the caller can already read, so
// it needs no server-side state change and no permission of its own beyond `read` — which is why
// the IAM seed has no `print` action either.

// Send marks a request for quotation as sent to the vendor (BR §13).
//
// The status change is the whole of it: this module does not send email. What it records is that
// the RFQ went out, which is what makes "waiting on a quote" distinguishable from "not yet asked",
// and what stops a second send from being the first.
//
// Only an RFQ can be sent. Sending a confirmed order would be asking for a quotation on something
// already bought.
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

// Lock closes an order to further editing (BR §20).
//
// Locking is a boolean and never a status (PUR-R2): an order is locked *and* confirmed, not locked
// *instead of* confirmed. Keeping them apart is what lets auto_lock work without inventing a
// status that the transition table would then have to route around.
//
// Only a committed order can be locked. Locking a draft would freeze a document nobody has agreed
// to, which is a state with no way out except unlocking it again.
func (this *PurchaseOrderDomainServiceImpl) Lock(
	ctx corectx.Context, orderId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.setLocked(ctx, orderId, true, "", AuditActionLock)
}

// Unlock reopens a locked order for editing, and requires a reason (BR §21).
//
// The reason is mandatory here and optional on cancel, which is the requirement's judgement and a
// defensible one: unlocking undoes a control that was deliberately applied, so the trail needs to
// say why. Refusing the operation without one is better than accepting a blank, because a trail of
// unexplained unlocks is the same as no trail.
func (this *PurchaseOrderDomainServiceImpl) Unlock(
	ctx corectx.Context, orderId string, reason string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if reason == "" {
		return orderViolationResult("purchase_order.unlock_reason_required",
			"unlocking a purchase order requires a reason"), nil
	}
	return this.setLocked(ctx, orderId, false, reason, AuditActionUnlock)
}

// Acknowledge records that the vendor confirmed receipt of the order (BR §22).
//
// Like is_locked this is a flag rather than a status, and for the same reason: acknowledgement is
// something the vendor does, and an order is confirmed whether or not they have got round to it.
//
// Only a committed order can be acknowledged — there is nothing to acknowledge before then.
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

// Duplicate copies an order and its lines into a new draft RFQ (BR §15).
//
// The copy starts at rfq with a fresh code and none of the original's history: no confirmed_at, no
// approval, no acknowledgement, not locked. Carrying any of those across would produce a document
// claiming to have been approved by someone who never saw it.
//
// This is also the answer to "can I revive a cancelled order" — duplicate it. The new order is
// visibly a new one, which a revived original would not be.
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

// copyableOrderFields is the terms of the order without any of its history.
//
// It is an allowlist rather than "everything except a few", deliberately: a field added to the
// schema later defaults to NOT being copied, which is the safe direction. The alternative would
// silently carry a new history-bearing field into every duplicate from the day it was added.
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
	// order_deadline is deliberately NOT copied: it is a date in the original's past by the time
	// anyone duplicates it, and a deadline that has already gone is worse than none.
	//
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

// transition is the shared body of the simple status changes. Confirm, Approve and Cancel do NOT
// use it: each of them reads configuration, stamps evidence or decides its own target, and folding
// those into a table-driven helper would hide the rules rather than share them.
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
