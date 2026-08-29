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

// Confirm, Approve and Cancel move an order through its lifecycle. Each runs in one transaction,
// writes exactly one audit event, and re-reads the order's status inside that transaction before
// acting: only the read inside the transaction can see what a concurrent operation did.

// Confirm commits the order, ending in to_approve when the org's policy requires approval for an
// order of this size and purchase_order when it does not. The decision reads the stored totals,
// recomputed first so it is made against what the lines say rather than a drifted header. Under the
// auto_lock modification policy the order also comes out locked; is_locked is a separate boolean and
// never a status. alternativeChoice says what happens to the other alternatives in this order's
// sourcing group, and an empty value means the caller has not decided, which triggers the warning.
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

		// Totals must be current before the approval decision reads them, or a stale total could
		// route a large order past an approver.
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

		// Confirming one alternative leaves the others quoting for a requirement already met, so
		// the caller must say what happens to them. The warning is a refusal, not a note:
		// defaulting either way would make a purchasing decision on their behalf.
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
		// confirmed_at marks the commitment, so it is stamped only on becoming one; an order in
		// to_approve has been submitted, not confirmed.
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

// assertConfirmable refuses an order with no money-bearing line: it would commit to buying nothing
// and produce a document a vendor cannot act on.
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

// Approve moves an order out of to_approve, recording approved_by and approved_at as the evidence
// that spending control was applied — which is why it is its own operation and not a status edit.
// There is deliberately no `rejected` state: an approver who will not approve cancels the order,
// recording the refusal with a reason.
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

// Cancel calls the order off from any status but cancelled, including after confirmation. It
// removes nothing — the order, its lines and its trail all stay, which is the difference between
// cancel and delete, and why delete is refused except from cancelled. The reason is optional.
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
			// Already cancelled: a retry after a lost response is not an error.
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

// alternativesWarningResult asks the caller to decide what happens to the open alternatives. It
// names their codes rather than counting them, because which vendors are still being asked changes
// the buyer's decision.
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

// writeOrderChanges applies a status-bearing change set to an order. It goes through the repository
// rather than the resource service because most of these fields are no_update and the lifecycle
// operations are the exception to that. The etag is carried through so a concurrent edit still loses.
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
