package services

import (
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// Processing a return, and cancelling one.
//
// Goods come back, money goes back, and the tax authority is told. Each side effect fails and retries
// on its own, because a database transaction cannot roll back an external API call. A failed fiscal
// adjustment must not roll back a completed inventory return or refund, and must never take money
// back from the customer: the return reaches completed on the two customer-facing steps alone,
// carrying the fiscal failure as a visible, retryable flag.

const (
	ReasonReturnNotFound        = "sales_return.not_found"
	ReasonReturnNotProcessable  = "sales_return.not_processable"
	ReasonReturnNotCancellable  = "sales_return.not_cancellable"
	ReasonReturnAlreadyComplete = "sales_return.already_complete"
)

type ProcessReturnParams struct {
	SalesReturnId string
}

type ProcessReturnResult struct {
	SalesReturnId string
	Status        string

	InventoryReturnStatus  string
	RefundStatus           string
	FiscalAdjustmentStatus string

	// Pending names what did not complete: a client treating a 200 as "everything done" would tell a
	// customer their refund had been made when it may still be in flight.
	Pending []string

	// FiscalRetryRequired is surfaced rather than inferred from the status: a return can be complete
	// and still need somebody in Accounting to act.
	FiscalRetryRequired bool

	// AdjustmentOrderId names the order restating what the customer kept, when a partial return
	// produced one. Empty for a full return, where nothing was kept, and for a return that has not
	// completed.
	AdjustmentOrderId string
}

// ProcessReturn takes the order's lock, not the return's.
//
// It writes the order's lines — returned_quantity, once the return completes — and re-derives the
// order's payment status from the refunds it settles, so locking the return would let a concurrent
// confirm or cancel interleave with those writes.
func ProcessReturn(
	ctx corectx.Context,
	params ProcessReturnParams,
	dLock lock.DistributedLock,
	fulfillment itExt.FulfillmentExtService,
	invoicing itInvoicing.InvoicingExtService,
	paymentOrders itExt.PaymentOrderExtService,
) (*ProcessReturnResult, *ft.ClientErrors, error) {
	if dLock == nil {
		return nil, nil, errors.New(
			"the distributed lock is not available; a return cannot be processed unguarded")
	}

	salesReturn, err := loadRecord(ctx, models.SalesReturnSchemaName,
		models.SalesReturnFieldId, params.SalesReturnId)
	if err != nil {
		return nil, nil, err
	}
	if salesReturn == nil {
		return nil, refuseReturn(models.SalesReturnFieldId, ReasonReturnNotFound,
			"no such return"), nil
	}

	orderId := stringOf(salesReturn, models.SalesReturnFieldSalesOrderId)
	key := confirmLockKeyOf(orderId)
	acquired, err := dLock.AcquireWithRetry(ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, err
	}
	if !acquired {
		return nil, refuseReturn(models.SalesReturnFieldId, ReasonLockUnavailable,
			"this order is being changed by another request; try again"), nil
	}
	defer func() { _ = dLock.Release(ctx, key) }()

	salesReturn, err = loadRecord(ctx, models.SalesReturnSchemaName,
		models.SalesReturnFieldId, params.SalesReturnId)
	if err != nil {
		return nil, nil, err
	}
	if salesReturn == nil {
		return nil, refuseReturn(models.SalesReturnFieldId, ReasonReturnNotFound,
			"no such return"), nil
	}

	return processReturnUnderLock(ctx, salesReturn, fulfillment, invoicing, paymentOrders)
}

func processReturnUnderLock(
	ctx corectx.Context,
	salesReturn dmodel.DynamicFields,
	fulfillment itExt.FulfillmentExtService,
	invoicing itInvoicing.InvoicingExtService,
	paymentOrders itExt.PaymentOrderExtService,
) (*ProcessReturnResult, *ft.ClientErrors, error) {
	current := stringOf(salesReturn, models.SalesReturnFieldStatus)
	if current == string(models.SalesReturnStatusCompleted) {
		return nil, refuseReturn(models.SalesReturnFieldStatus, ReasonReturnAlreadyComplete,
			"this return has already completed"), nil
	}
	if current != string(models.SalesReturnStatusDraft) &&
		current != string(models.SalesReturnStatusApproved) &&
		current != string(models.SalesReturnStatusProcessing) {
		return nil, refuseReturn(models.SalesReturnFieldStatus, ReasonReturnNotProcessable,
			"a return in status "+current+" cannot be processed"), nil
	}

	result := &ProcessReturnResult{
		SalesReturnId:          stringOf(salesReturn, models.SalesReturnFieldId),
		InventoryReturnStatus:  stringOf(salesReturn, models.SalesReturnFieldInventoryReturnStatus),
		RefundStatus:           stringOf(salesReturn, models.SalesReturnFieldRefundStatus),
		FiscalAdjustmentStatus: stringOf(salesReturn, models.SalesReturnFieldFiscalAdjustmentStatus),
	}

	// Order matters: goods first, because refunding before the goods arrive is a loss the business
	// absorbs; money second; the tax correction last, since it must never block the others.
	inventoryReference, err := runInventoryStep(ctx, salesReturn, fulfillment, result)
	if err != nil {
		return nil, nil, err
	}
	if err := runRefundStep(ctx, salesReturn, result, paymentOrders); err != nil {
		return nil, nil, err
	}

	// The commercial status is decided before the fiscal step runs, from the two customer-facing
	// statuses alone. Moving this below the fiscal step would not change the answer, but would invite
	// somebody to make the fiscal status an input.
	result.Status = DeriveReturnStatus(current, result.InventoryReturnStatus, result.RefundStatus)

	if err := runFiscalStep(ctx, salesReturn, invoicing, result); err != nil {
		return nil, nil, err
	}
	result.FiscalRetryRequired = result.FiscalAdjustmentStatus == string(models.SalesReturnStepFailed)

	result.Pending = pendingReturnSteps(result)

	if err := writeReturnOutcome(ctx, salesReturn, result, inventoryReference); err != nil {
		return nil, nil, err
	}

	// The order is told what the refunds did to it. Without this the money is back with the customer
	// and the order still reads `paid`, because payment status is derived on demand and nothing was
	// asking after a return.
	if _, err := SyncOrderPaymentStatus(ctx,
		stringOf(salesReturn, models.SalesReturnFieldSalesOrderId)); err != nil {
		return nil, nil, err
	}

	// What the customer kept is restated as its own order, once everything else has settled. It runs
	// last because it reads returned_quantity, which writeReturnOutcome has only just written — and
	// only on completion, since a return that failed at the inventory step has not taken anything
	// back yet.
	if result.Status == string(models.SalesReturnStatusCompleted) {
		adjustment, err := CreateAdjustmentOrder(ctx, salesReturn)
		if err != nil {
			return nil, nil, err
		}
		if adjustment.Created {
			result.AdjustmentOrderId = adjustment.AdjustmentOrderId
		}
	}
	return result, nil, nil
}

// runInventoryStep asks Inventory to take the goods back. A not_required step is left alone: raising
// a movement for a service would tell Inventory to receive something that does not exist.
func runInventoryStep(
	ctx corectx.Context,
	salesReturn dmodel.DynamicFields,
	fulfillment itExt.FulfillmentExtService,
	result *ProcessReturnResult,
) (string, error) {
	if result.InventoryReturnStatus == string(models.SalesReturnStepNotRequired) ||
		result.InventoryReturnStatus == string(models.SalesReturnStepCompleted) {
		return "", nil
	}

	lines, err := returnLinesOf(ctx, stringOf(salesReturn, models.SalesReturnFieldId))
	if err != nil {
		return "", err
	}

	fulfillmentLines := make([]itExt.FulfillmentLine, 0, len(lines))
	for _, line := range lines {
		if !boolOrTrue(line, models.SalesReturnLineFieldRequiresInventoryReturn) {
			continue
		}
		orderLine, err := loadRecord(ctx, models.SalesOrderLineSchemaName,
			models.SalesOrderLineFieldId, stringOf(line, models.SalesReturnLineFieldSalesOrderLineId))
		if err != nil {
			return "", err
		}
		if orderLine == nil {
			continue
		}
		fulfillmentLines = append(fulfillmentLines, itExt.FulfillmentLine{
			SalesOrderLineId: stringOf(line, models.SalesReturnLineFieldSalesOrderLineId),
			ProductVariantId: stringOf(orderLine, models.SalesOrderLineFieldProductVariantId),
			UomId:            stringOf(orderLine, models.SalesOrderLineFieldUomId),
			Quantity:         decimalOf(line, models.SalesReturnLineFieldQuantity),
		})
	}

	if len(fulfillmentLines) == 0 {
		result.InventoryReturnStatus = string(models.SalesReturnStepNotRequired)
		return "", nil
	}

	// A nil port is a supported state, not a fault: no adapter ships in every build. The step stays
	// pending, which is honest.
	if fulfillment == nil {
		result.InventoryReturnStatus = string(models.SalesReturnStepPending)
		return "", nil
	}

	response, err := fulfillment.RequestReturnReceipt(ctx, itExt.FulfillmentRequest{
		SalesOrderId:              stringOf(salesReturn, models.SalesReturnFieldSalesOrderId),
		SalesFulfillmentRequestId: stringOf(salesReturn, models.SalesReturnFieldId),
		IdempotencyKey:            stringOf(salesReturn, models.SalesReturnFieldId),
		Lines:                     fulfillmentLines,
	})
	if err != nil {
		// A transport failure is a fault, not a refusal; the step stays where it was so a retry picks
		// it up.
		return "", err
	}

	switch {
	case response == nil:
		result.InventoryReturnStatus = string(models.SalesReturnStepPending)
	case !response.Accepted:
		result.InventoryReturnStatus = string(models.SalesReturnStepFailed)
	case response.Completed:
		result.InventoryReturnStatus = string(models.SalesReturnStepCompleted)
	default:
		result.InventoryReturnStatus = string(models.SalesReturnStepProcessing)
	}

	if response != nil {
		return response.InventoryReference, nil
	}
	return "", nil
}

// runRefundStep allocates the refund across the original payments and records the legs. Money returns
// by the route it arrived: every leg names an original payment and is capped at what that payment
// captured, so a refund can never exceed money that came in or be an unexplained outflow.
func runRefundStep(
	ctx corectx.Context,
	salesReturn dmodel.DynamicFields,
	result *ProcessReturnResult,
	paymentOrders itExt.PaymentOrderExtService,
) error {
	if result.RefundStatus == string(models.SalesReturnStepNotRequired) ||
		result.RefundStatus == string(models.SalesReturnStepCompleted) {
		return nil
	}

	refundTotal := decimalOf(salesReturn, models.SalesReturnFieldRefundTotal)
	if !refundTotal.IsPositive() {
		result.RefundStatus = string(models.SalesReturnStepNotRequired)
		return nil
	}

	existing, err := searchBy(ctx, models.SalesRefundPaymentSchemaName,
		models.SalesRefundPaymentFieldSalesReturnId, stringOf(salesReturn, models.SalesReturnFieldId))
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		// The legs are already written; this is a retry. Writing a second set would refund twice —
		// but the legs written last time may still be pending, so settlement is attempted again
		// before the step's status is decided.
		if _, err := SettleRefundLegs(ctx, salesReturn, paymentOrders); err != nil {
			return err
		}
		return refreshRefundStepStatus(ctx, salesReturn, result, refundTotal)
	}

	payments, err := capturedPaymentsOfOrder(ctx, stringOf(salesReturn, models.SalesReturnFieldSalesOrderId))
	if err != nil {
		return err
	}
	if len(payments) == 0 {
		// Nothing was captured, so there is nothing to give back; not_required lets the return
		// complete rather than waiting on a refund that can never be made.
		result.RefundStatus = string(models.SalesReturnStepNotRequired)
		return nil
	}

	legs := allocateRefund(refundTotal, payments)
	if err := writeRefundLegs(ctx, salesReturn, legs); err != nil {
		return err
	}

	// Written and then settled, rather than written as already settled: the legs are the record that
	// a refund was authorised, and they must exist before any money moves so that a failure halfway
	// through leaves evidence of what was owed.
	if _, err := SettleRefundLegs(ctx, salesReturn, paymentOrders); err != nil {
		return err
	}
	return refreshRefundStepStatus(ctx, salesReturn, result, refundTotal)
}

// refreshRefundStepStatus re-reads the legs and decides where the refund step stands.
//
// Completed only when the money actually back equals what was owed. A leg still pending or failed
// leaves the step `processing`, which is what brings the return round again rather than declaring a
// customer repaid who is still waiting.
func refreshRefundStepStatus(
	ctx corectx.Context,
	salesReturn dmodel.DynamicFields,
	result *ProcessReturnResult,
	refundTotal decimal.Decimal,
) error {
	legs, err := searchBy(ctx, models.SalesRefundPaymentSchemaName,
		models.SalesRefundPaymentFieldSalesReturnId,
		stringOf(salesReturn, models.SalesReturnFieldId))
	if err != nil {
		return err
	}

	if models.SumCompletedRefunds(legs).GreaterThanOrEqual(refundTotal) {
		result.RefundStatus = string(models.SalesReturnStepCompleted)
	} else {
		result.RefundStatus = string(models.SalesReturnStepProcessing)
	}
	return nil
}

// runFiscalStep tells the tax authority, through the invoicing provider. It can fail without
// consequence to the return: result.Status is already decided and nothing here writes to it, so
// there is no code path from a fiscal failure to the return's commercial status.
func runFiscalStep(
	ctx corectx.Context,
	salesReturn dmodel.DynamicFields,
	invoicing itInvoicing.InvoicingExtService,
	result *ProcessReturnResult,
) error {
	if result.FiscalAdjustmentStatus == string(models.SalesReturnStepNotRequired) ||
		result.FiscalAdjustmentStatus == string(models.SalesReturnStepCompleted) {
		return nil
	}

	bill, err := issuedBillOfOrder(ctx, stringOf(salesReturn, models.SalesReturnFieldSalesOrderId))
	if err != nil {
		return err
	}
	if bill == nil {
		// No invoice was issued, so there is nothing to adjust; the obligation never arose.
		result.FiscalAdjustmentStatus = string(models.SalesReturnStepNotRequired)
		return nil
	}

	intent, err := fiscalIntentFor(ctx, salesReturn)
	if err != nil {
		return err
	}

	_, vErrs, err := RequestInvoice(ctx, RequestInvoiceParams{
		SalesBillId: stringOf(bill, models.SalesBillFieldId),
		Intent:      intent,
		Reason:      stringOf(salesReturn, models.SalesReturnFieldReason),

		// The return's own id is the idempotency key, so a retry after a timeout adjusts the same
		// invoice rather than issuing a second correction against it.
		IdempotencyKey: "return-" + stringOf(salesReturn, models.SalesReturnFieldId),
	}, invoicing)

	switch {
	case err != nil:
		// A fault here is recorded, not propagated: propagating would fail the whole operation and
		// roll back goods and money that have already moved.
		result.FiscalAdjustmentStatus = string(models.SalesReturnStepFailed)
	case vErrs != nil:
		result.FiscalAdjustmentStatus = string(models.SalesReturnStepFailed)
	default:
		result.FiscalAdjustmentStatus = string(models.SalesReturnStepProcessing)
	}
	return nil
}

// fiscalIntentFor decides whether this return adjusts the whole invoice or part of it: full when
// every order line comes back in its entirety, partial otherwise. The authority treats the two
// differently.
func fiscalIntentFor(ctx corectx.Context, salesReturn dmodel.DynamicFields) (string, error) {
	orderLines, err := searchBy(ctx, models.SalesOrderLineSchemaName,
		models.SalesOrderLineFieldSalesOrderId, stringOf(salesReturn, models.SalesReturnFieldSalesOrderId))
	if err != nil {
		return "", err
	}
	returnLines, err := returnLinesOf(ctx, stringOf(salesReturn, models.SalesReturnFieldId))
	if err != nil {
		return "", err
	}

	returned := make(map[string]decimal.Decimal, len(returnLines))
	for _, line := range returnLines {
		id := stringOf(line, models.SalesReturnLineFieldSalesOrderLineId)
		returned[id] = returned[id].Add(decimalOf(line, models.SalesReturnLineFieldQuantity))
	}

	for _, line := range orderLines {
		ordered := decimalOf(line, models.SalesOrderLineFieldOrderedQuantity)
		if !returned[stringOf(line, models.SalesOrderLineFieldId)].GreaterThanOrEqual(ordered) {
			return string(models.SalesFiscalIntentAdjustForPartialReturn), nil
		}
	}
	return string(models.SalesFiscalIntentAdjustForFullReturn), nil
}

func pendingReturnSteps(result *ProcessReturnResult) []string {
	pending := make([]string, 0, 3)
	if !stepSettled(result.InventoryReturnStatus) {
		pending = append(pending, "inventory_return")
	}
	if !stepSettled(result.RefundStatus) {
		pending = append(pending, "refund")
	}
	// Reported as pending even when the return is complete: the customer is whole, and Accounting
	// still has work to do.
	if !stepSettled(result.FiscalAdjustmentStatus) {
		pending = append(pending, "fiscal_adjustment")
	}
	return pending
}

func writeReturnOutcome(
	ctx corectx.Context,
	salesReturn dmodel.DynamicFields,
	result *ProcessReturnResult,
	inventoryReference string,
) error {
	return withTransaction(ctx, models.SalesReturnSchemaName, func(tranxCtx corectx.Context) error {
		changes := dmodel.DynamicFields{
			models.SalesReturnFieldStatus:                 result.Status,
			models.SalesReturnFieldInventoryReturnStatus:  result.InventoryReturnStatus,
			models.SalesReturnFieldRefundStatus:           result.RefundStatus,
			models.SalesReturnFieldFiscalAdjustmentStatus: result.FiscalAdjustmentStatus,
		}
		if inventoryReference != "" {
			changes[models.SalesReturnFieldInventoryReference] = inventoryReference
		}
		if result.Status == string(models.SalesReturnStatusCompleted) {
			changes[models.SalesReturnFieldCompletedAt] = model.ModelDateTime(time.Now().UTC())
		}
		if err := writeChanges(
			tranxCtx, models.SalesReturnSchemaName, salesReturn, changes); err != nil {
			return err
		}

		// The order's lines are told what came back, in the same transaction as the return's own
		// outcome. Without this the order says it still holds goods the customer has handed back:
		// returned_quantity was written once at order creation and never again, so
		// DeriveFulfillmentStatus could never reach partially_returned however much came back.
		//
		// It does NOT replace the returnable-quantity guard, which counts non-cancelled return lines
		// and stays the authority — this is the order reflecting reality, not a new source of truth.
		return applyReturnedQuantities(tranxCtx, salesReturn, result.Status)
	})
}

// applyReturnedQuantities adds this return's lines to the order lines they came from.
//
// Only once the return is complete. A return still being processed may yet fail at the inventory
// step, and an order line that had already counted the goods back would understate what the customer
// still holds.
func applyReturnedQuantities(
	ctx corectx.Context, salesReturn dmodel.DynamicFields, status string,
) error {
	if status != string(models.SalesReturnStatusCompleted) {
		return nil
	}

	lines, err := returnLinesOf(ctx, stringOf(salesReturn, models.SalesReturnFieldId))
	if err != nil {
		return err
	}

	for _, line := range lines {
		orderLineId := stringOf(line, models.SalesReturnLineFieldSalesOrderLineId)
		if orderLineId == "" {
			continue
		}

		orderLine, err := loadRecord(ctx, models.SalesOrderLineSchemaName,
			models.SalesOrderLineFieldId, orderLineId)
		if err != nil {
			return err
		}
		if orderLine == nil {
			continue
		}

		returned := decimalOf(orderLine, models.SalesOrderLineFieldReturnedQuantity).
			Add(decimalOf(line, models.SalesReturnLineFieldQuantity))

		if err := writeChanges(ctx, models.SalesOrderLineSchemaName, orderLine,
			dmodel.DynamicFields{
				models.SalesOrderLineFieldReturnedQuantity: returned,
			}); err != nil {
			return err
		}
	}
	return nil
}

func returnLinesOf(ctx corectx.Context, returnId string) ([]dmodel.DynamicFields, error) {
	return searchBy(ctx, models.SalesReturnLineSchemaName,
		models.SalesReturnLineFieldSalesReturnId, returnId)
}

type CancelReturnParams struct {
	SalesReturnId string
	Reason        string
}

type CancelReturnResult struct {
	SalesReturnId string
	Status        string
}

// CancelReturn calls off a return before anything irreversible has happened. Refused once processing
// has begun, since goods may already be moving or money leaving and there is no honest way to un-ask
// for either; the way out then is another return or a fresh sale.
func CancelReturn(
	ctx corectx.Context,
	params CancelReturnParams,
	dLock lock.DistributedLock,
) (*CancelReturnResult, *ft.ClientErrors, error) {
	if dLock == nil {
		return nil, nil, errors.New(
			"the distributed lock is not available; a return cannot be cancelled unguarded")
	}

	salesReturn, err := loadRecord(ctx, models.SalesReturnSchemaName,
		models.SalesReturnFieldId, params.SalesReturnId)
	if err != nil {
		return nil, nil, err
	}
	if salesReturn == nil {
		return nil, refuseReturn(models.SalesReturnFieldId, ReasonReturnNotFound,
			"no such return"), nil
	}

	key := confirmLockKeyOf(stringOf(salesReturn, models.SalesReturnFieldSalesOrderId))
	acquired, err := dLock.AcquireWithRetry(ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, err
	}
	if !acquired {
		return nil, refuseReturn(models.SalesReturnFieldId, ReasonLockUnavailable,
			"this order is being changed by another request; try again"), nil
	}
	defer func() { _ = dLock.Release(ctx, key) }()

	salesReturn, err = loadRecord(ctx, models.SalesReturnSchemaName,
		models.SalesReturnFieldId, params.SalesReturnId)
	if err != nil {
		return nil, nil, err
	}
	if salesReturn == nil {
		return nil, refuseReturn(models.SalesReturnFieldId, ReasonReturnNotFound,
			"no such return"), nil
	}

	current := stringOf(salesReturn, models.SalesReturnFieldStatus)
	if !CanTransitionReturn(current, string(models.SalesReturnStatusCancelled)) {
		return nil, refuseReturn(models.SalesReturnFieldStatus, ReasonReturnNotCancellable,
			"a return in status "+current+" can no longer be cancelled; goods or money may"+
				" already have moved"), nil
	}

	err = withTransaction(ctx, models.SalesReturnSchemaName, func(tranxCtx corectx.Context) error {
		return writeChanges(tranxCtx, models.SalesReturnSchemaName, salesReturn, dmodel.DynamicFields{
			models.SalesReturnFieldStatus:        string(models.SalesReturnStatusCancelled),
			models.SalesReturnFieldCancelledAt:   model.ModelDateTime(time.Now().UTC()),
			models.SalesReturnFieldFailureReason: params.Reason,
		})
	})
	if err != nil {
		return nil, nil, err
	}

	return &CancelReturnResult{
		SalesReturnId: params.SalesReturnId,
		Status:        string(models.SalesReturnStatusCancelled),
	}, nil, nil
}

func orgIdOf(record dmodel.DynamicFields) string {
	return stringOf(record, basemodel.FieldOrgId)
}
