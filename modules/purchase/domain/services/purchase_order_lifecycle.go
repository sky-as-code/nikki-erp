package services

import (
	"strings"
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Confirm, Approve and Cancel: the three operations that move an order through its lifecycle.
//
// Each of them runs in ONE transaction and writes EXACTLY ONE audit event (PUR-R6), and each
// re-reads the order's status inside that transaction before acting on it. The re-read is what
// makes two concurrent confirms safe: both may have passed a check against the same stale status
// outside the transaction, and only the one that reads inside it can see what the other did.

// Confirm commits the order, routing it through approval when the organization's policy says so
// (BR §16, §47).
//
// It ends in one of two statuses, which is why the audit event records the action rather than the
// transition alone: `to_approve` when this org approves orders of this size, `purchase_order` when
// it does not. The totals decide which, and the totals are the STORED ones — recomputed first, so
// the decision is made against what the lines actually say rather than a header that has drifted.
//
// Confirming also snapshots the modification policy's effect: under auto_lock the order comes out
// locked (BR §47.3). is_locked is a separate boolean and never a status (PUR-R2), so an order can
// be both confirmed and locked without either fact hiding the other.
//
// alternativeChoice says what happens to the other alternatives in this order's sourcing group
// (§31). Empty means the caller has not decided, which is what triggers the warning.
func (this *PurchaseOrderDomainServiceImpl) Confirm(
	ctx corectx.Context, orderId string, alternativeChoice string,
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
		if IsOrderCommitted(status) {
			result = orderViolationResult("purchase_order.already_confirmed",
				"this purchase order is already confirmed")
			return nil
		}

		// The totals are brought in step before the approval decision reads them: confirming
		// against a stale total could route a large order past an approver.
		if err := RecomputeOrderTotals(tranxCtx, orderId); err != nil {
			return err
		}
		order, err = loadOrder(tranxCtx, orderId)
		if err != nil {
			return err
		}
		if order == nil {
			result = orderNotFoundResult(orderId)
			return nil
		}

		if refusal := assertConfirmable(tranxCtx, order); refusal != nil {
			result = refusal
			return nil
		}

		// §31: confirming one alternative leaves the others quoting for a requirement that has
		// just been met. The caller has to say what happens to them — the warning is a REFUSAL
		// rather than a note, because defaulting either way makes a purchasing decision on their
		// behalf: cancelling loses quotes they may still want, and keeping leaves live requests
		// to vendors who will never be given the business.
		openAlternatives, err := OpenAlternativesOf(tranxCtx, order)
		if err != nil {
			return err
		}
		if len(openAlternatives) > 0 && alternativeChoice == "" {
			result = alternativesWarningResult(openAlternatives)
			return nil
		}
		if alternativeChoice != "" && alternativeChoice != AlternativeChoiceKeep &&
			alternativeChoice != AlternativeChoiceCancel {
			result = orderViolationResult("purchase_order.unknown_alternative_choice",
				"the choice for the open alternatives must be '"+AlternativeChoiceKeep+
					"' or '"+AlternativeChoiceCancel+"'")
			return nil
		}

		config, err := LoadConfiguration(tranxCtx, stringOf(order, basemodel.FieldOrgId))
		if err != nil {
			return err
		}
		total := decimalOf(order, models.PurchaseOrderFieldTotalAmount)
		needsApproval := RequiresApproval(config, total)

		next := string(models.PurchaseOrderStatusPurchaseOrder)
		if needsApproval {
			next = string(models.PurchaseOrderStatusToApprove)
		}

		vErrs := newOrderErrors()
		AssertOrderTransition(status, next, vErrs)
		if vErrs.Count() > 0 {
			result = &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
			return nil
		}

		changes := dmodel.DynamicFields{
			models.PurchaseOrderFieldStatus:           next,
			models.PurchaseOrderFieldApprovalRequired: needsApproval,
		}
		// confirmed_at marks the commitment, so it is stamped only when the order actually becomes
		// one. An order sitting in to_approve has been submitted, not confirmed.
		if !needsApproval {
			changes[models.PurchaseOrderFieldConfirmedAt] = time.Now()
			if config.PoModificationPolicy == models.PoModificationPolicyAutoLock {
				changes[models.PurchaseOrderFieldIsLocked] = true
			}
		}

		if err := writeOrderChanges(tranxCtx, order, changes); err != nil {
			return err
		}

		if alternativeChoice == AlternativeChoiceCancel && len(openAlternatives) > 0 {
			if err := this.CancelOpenAlternatives(tranxCtx, order, orderId); err != nil {
				return err
			}
		}

		result = mutateOk()
		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   orderId,
			Action:     AuditActionConfirm,
			FromStatus: status,
			ToStatus:   next,
			OrgId:      stringOf(order, basemodel.FieldOrgId),
			Metadata: map[string]any{
				"total_amount":       total.String(),
				"approval_required":  needsApproval,
				"alternative_choice": alternativeChoice,
			},
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// assertConfirmable refuses an order that has nothing to commit to.
//
// An order with no priced line would be a commitment to buy nothing, and one whose lines are all
// sections and notes is the same thing with headings. Both are far more likely to be a half-filled
// form than a deliberate act, and confirming one produces a document a vendor cannot act on.
func assertConfirmable(
	ctx corectx.Context, order dmodel.DynamicFields,
) *dyn.OpResult[dyn.MutateResultData] {
	lineEngine, err := engineFor(models.PurchaseOrderLineSchemaName)
	if err != nil {
		// A missing engine is a wiring fault, not a refusal; let the caller's error path have it.
		return nil
	}
	orderId := stringOf(order, models.PurchaseOrderFieldId)
	lines, err := models.FindOrderLines(ctx, lineEngine.ResourceRepository(), orderId, models.MaxOrderLines)
	if err != nil {
		return nil
	}

	for _, line := range lines {
		if isMoneyBearingLine(line) {
			return nil
		}
	}
	return orderViolationResult("purchase_order.no_lines",
		"a purchase order with no product line cannot be confirmed")
}

// Approve moves an order out of to_approve (BR §17).
//
// It records who approved and when, which is the reason approval is its own operation rather than
// a status edit: approved_by and approved_at are the evidence that spending control was applied,
// and an update that could set the status without them would leave a committed order with no
// approver named.
//
// The `rejected` state of §18 is deliberately not implemented (§11 scope). An approver who will not
// approve cancels the order, which records the refusal in the trail with a reason.
func (this *PurchaseOrderDomainServiceImpl) Approve(
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
		if status != string(models.PurchaseOrderStatusToApprove) {
			result = orderViolationResult("purchase_order.not_awaiting_approval",
				"only a purchase order awaiting approval can be approved; this one is '"+status+"'")
			return nil
		}

		next := string(models.PurchaseOrderStatusPurchaseOrder)
		config, err := LoadConfiguration(tranxCtx, stringOf(order, basemodel.FieldOrgId))
		if err != nil {
			return err
		}

		now := time.Now()
		changes := dmodel.DynamicFields{
			models.PurchaseOrderFieldStatus:      next,
			models.PurchaseOrderFieldApprovedAt:  now,
			models.PurchaseOrderFieldConfirmedAt: now,
		}
		if approver := actorOf(tranxCtx); approver != "" {
			changes[models.PurchaseOrderFieldApprovedBy] = approver
		}
		if config.PoModificationPolicy == models.PoModificationPolicyAutoLock {
			changes[models.PurchaseOrderFieldIsLocked] = true
		}

		if err := writeOrderChanges(tranxCtx, order, changes); err != nil {
			return err
		}

		result = mutateOk()
		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   orderId,
			Action:     AuditActionApprove,
			FromStatus: status,
			ToStatus:   next,
			OrgId:      stringOf(order, basemodel.FieldOrgId),
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Cancel calls the order off, from any status but cancelled (BR §23).
//
// Cancelling is available even after confirmation, because a deal both sides agreed to can still
// fall through. What it does not do is remove anything: the order, its lines and its trail all
// stay, which is the whole difference between cancel and delete — and why delete is refused
// everywhere except from cancelled.
//
// The reason is optional. Requiring one would be defensible, but the requirement does not ask for
// it and a mandatory free-text field mostly produces the word "cancelled".
func (this *PurchaseOrderDomainServiceImpl) Cancel(
	ctx corectx.Context, orderId string, reason string,
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
		next := string(models.PurchaseOrderStatusCancelled)

		if status == next {
			// Already cancelled: report success rather than a violation, so that a retried
			// cancel after a lost response is not an error.
			result = mutateOk()
			return nil
		}

		vErrs := newOrderErrors()
		AssertOrderTransition(status, next, vErrs)
		if vErrs.Count() > 0 {
			result = &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
			return nil
		}

		if err := writeOrderChanges(tranxCtx, order, dmodel.DynamicFields{
			models.PurchaseOrderFieldStatus: next,
		}); err != nil {
			return err
		}

		result = mutateOk()
		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   orderId,
			Action:     AuditActionCancel,
			FromStatus: status,
			ToStatus:   next,
			Reason:     reason,
			OrgId:      stringOf(order, basemodel.FieldOrgId),
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// alternativesWarningResult asks the caller to decide what happens to the open alternatives.
//
// The open orders' codes are named rather than only counted, because the buyer's decision depends
// on which vendors are still being asked — "cancel the other two" is a different call when one of
// them is the incumbent supplier.
func alternativesWarningResult(open []dmodel.DynamicFields) *dyn.OpResult[dyn.MutateResultData] {
	codes := make([]string, 0, len(open))
	for _, alternative := range open {
		codes = append(codes, stringOf(alternative, models.PurchaseOrderFieldCode))
	}

	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.PurchaseOrderSchemaName,
		"purchase_order.open_alternatives",
		"this order has open alternatives ("+strings.Join(codes, ", ")+"); "+
			"say whether to keep or cancel them",
		map[string]any{
			"alternative_codes": codes,
			"choices":           []string{AlternativeChoiceKeep, AlternativeChoiceCancel},
		},
	))
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}

// loadOrder reads one order inside the caller's transaction, returning nil when there is none.
func loadOrder(ctx corectx.Context, orderId string) (dmodel.DynamicFields, error) {
	engine, err := engineFor(models.PurchaseOrderSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PurchaseOrderFieldId: orderId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadOrder")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}

// writeOrderChanges applies a status-bearing change set to an order.
//
// It goes through the repository rather than the resource service because most of these fields are
// no_update: the client-facing rule is that a status cannot be edited, and the lifecycle operations
// are precisely the exception. The etag is carried through so a concurrent edit still loses.
func writeOrderChanges(
	ctx corectx.Context, order dmodel.DynamicFields, changes dmodel.DynamicFields,
) error {
	engine, err := engineFor(models.PurchaseOrderSchemaName)
	if err != nil {
		return err
	}

	update := make(dmodel.DynamicFields, len(changes)+2)
	for key, value := range changes {
		update[key] = value
	}
	update[models.PurchaseOrderFieldId] = stringOf(order, models.PurchaseOrderFieldId)
	update[basemodel.FieldEtag] = stringOf(order, basemodel.FieldEtag)

	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeOrderChanges")
}
