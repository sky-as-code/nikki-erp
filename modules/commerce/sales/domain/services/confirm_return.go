package services

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
)

// Confirming a refund request (CR-INV-SALES-WH-RESERVATION §6.3).
//
// A refund request is a sales return; confirming it is business approval, distinct from the money
// moving. Confirm records the note on the request and projects it onto the order, locks the amount
// against the order's other pending requests, moves the request to approved, and only then
// dispatches it through the existing processing step. A request already confirmed replays its
// current state and never overwrites the note it recorded first.

const (
	ReasonRefundAmountExceeded = "sales_return.refund_amount_exceeded"
	ReasonReturnNotConfirmable = "sales_return.not_confirmable"

	maxNoteLength = 2000
)

type ConfirmRefundRequestParams struct {
	SalesReturnId string

	// RefundNote is the confirmer's note, optional. Automatic sets it to exactly AutoConfirmedNote.
	RefundNote string
	Automatic  bool
}

type ConfirmRefundRequestResult struct {
	SalesReturnId    string
	Status           string
	ConfirmationNote string
	LockedAmount     decimal.Decimal

	// AlreadyConfirmed says this call found the approval done and changed nothing.
	AlreadyConfirmed bool

	// Processing is what dispatching produced, nil when nothing was dispatched this call.
	Processing *ProcessReturnResult
}

// RefundProcessingDeps are the ports dispatching a refund needs, bundled so a cancel that raises
// and confirms a refund can carry them in one value.
type RefundProcessingDeps struct {
	Fulfillment   itExt.FulfillmentExtService
	Invoicing     itInvoicing.InvoicingExtService
	PaymentOrders itExt.PaymentOrderExtService
}

// ConfirmRefundRequest approves a draft refund request and dispatches it, under the order lock.
func ConfirmRefundRequest(
	ctx corectx.Context,
	params ConfirmRefundRequestParams,
	dLock lock.DistributedLock,
	deps RefundProcessingDeps,
) (*ConfirmRefundRequestResult, *ft.ClientErrors, error) {
	if dLock == nil {
		return nil, nil, errors.New(
			"the distributed lock is not available; a refund request cannot be confirmed unguarded")
	}
	salesReturn, err := loadRecord(ctx, models.SalesReturnSchemaName, models.SalesReturnFieldId, params.SalesReturnId)
	if err != nil {
		return nil, nil, err
	}
	if salesReturn == nil {
		return nil, refuseReturn(models.SalesReturnFieldId, ReasonReturnNotFound, "no such return"), nil
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

	return confirmRefundRequestUnderLock(ctx, params, deps)
}

func confirmRefundRequestUnderLock(
	ctx corectx.Context, params ConfirmRefundRequestParams, deps RefundProcessingDeps,
) (*ConfirmRefundRequestResult, *ft.ClientErrors, error) {
	salesReturn, err := loadRecord(ctx, models.SalesReturnSchemaName, models.SalesReturnFieldId, params.SalesReturnId)
	if err != nil {
		return nil, nil, err
	}
	if salesReturn == nil {
		return nil, refuseReturn(models.SalesReturnFieldId, ReasonReturnNotFound, "no such return"), nil
	}

	status := stringOf(salesReturn, models.SalesReturnFieldStatus)
	switch status {
	case string(models.SalesReturnStatusDraft):
	case string(models.SalesReturnStatusApproved), string(models.SalesReturnStatusProcessing),
		string(models.SalesReturnStatusCompleted):
		return &ConfirmRefundRequestResult{
			SalesReturnId:    params.SalesReturnId,
			Status:           status,
			ConfirmationNote: stringOf(salesReturn, models.SalesReturnFieldConfirmationNote),
			LockedAmount:     decimalOf(salesReturn, models.SalesReturnFieldLockedRefundAmount),
			AlreadyConfirmed: true,
		}, nil, nil
	default:
		return nil, refuseReturn(models.SalesReturnFieldStatus, ReasonReturnNotConfirmable,
			"a return in status "+status+" cannot be confirmed"), nil
	}

	orderId := stringOf(salesReturn, models.SalesReturnFieldSalesOrderId)
	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, nil, err
	}
	if order == nil {
		return nil, refuseReturn(models.SalesReturnFieldSalesOrderId, ReasonReturnOrderNotFound, "no such order"), nil
	}

	refundTotal := decimalOf(salesReturn, models.SalesReturnFieldRefundTotal)
	if refundTotal.IsPositive() {
		refundable, err := refundableAmountOf(ctx, orderId, params.SalesReturnId)
		if err != nil {
			return nil, nil, err
		}
		if refundTotal.GreaterThan(refundable) {
			return nil, refuseReturn(models.SalesReturnFieldRefundTotal, ReasonRefundAmountExceeded,
				"this request asks for "+refundTotal.String()+" but only "+refundable.String()+
					" of the order's money is still refundable"), nil
		}
	}

	note := normaliseNote(params.RefundNote)
	if params.Automatic {
		note = models.AutoConfirmedNote
	}
	now := time.Now().UTC()
	err = withTransaction(ctx, models.SalesReturnSchemaName, func(tranxCtx corectx.Context) error {
		if err := writeChanges(tranxCtx, models.SalesReturnSchemaName, salesReturn, dmodel.DynamicFields{
			models.SalesReturnFieldStatus:             string(models.SalesReturnStatusApproved),
			models.SalesReturnFieldConfirmationNote:   note,
			models.SalesReturnFieldConfirmedAt:        model.ModelDateTime(now),
			models.SalesReturnFieldLockedRefundAmount: refundTotal,
		}); err != nil {
			return err
		}
		// The order's refund_note is the latest confirmed note, a projection; the request keeps
		// its own copy above, so nothing is lost when the next request overwrites this.
		return writeChanges(tranxCtx, models.SalesOrderSchemaName, order, dmodel.DynamicFields{
			models.SalesOrderFieldRefundNote: note,
		})
	})
	if err != nil {
		return nil, nil, err
	}

	result := &ConfirmRefundRequestResult{
		SalesReturnId:    params.SalesReturnId,
		Status:           string(models.SalesReturnStatusApproved),
		ConfirmationNote: note,
		LockedAmount:     refundTotal,
	}

	// Dispatch, through the same processing every return goes through. Its outcome is reported,
	// not folded into the approval: a gateway that is down leaves an approved request the sweep
	// retries, not an unconfirmed one.
	confirmed, err := loadRecord(ctx, models.SalesReturnSchemaName, models.SalesReturnFieldId, params.SalesReturnId)
	if err != nil || confirmed == nil {
		return result, nil, err
	}
	processed, vErrs, err := processReturnUnderLock(ctx, confirmed, deps.Fulfillment, deps.Invoicing, deps.PaymentOrders)
	if err != nil {
		return nil, nil, err
	}
	if vErrs != nil {
		return nil, vErrs, nil
	}
	result.Processing = processed
	if processed != nil && processed.Status != "" {
		result.Status = processed.Status
	}
	return result, nil, nil
}

// refundableAmountOf is captured − already refunded − locked by the order's other pending
// requests. The pending lock is what stops two requests raised at once from both being promised
// the same money (CR §8.2).
func refundableAmountOf(ctx corectx.Context, orderId, exceptReturnId string) (decimal.Decimal, error) {
	captured, err := capturedPaymentsOfOrder(ctx, orderId)
	if err != nil {
		return decimal.Zero, err
	}
	total := decimal.Zero
	for _, payment := range captured {
		total = total.Add(payment.Captured)
	}

	returns, err := searchBy(ctx, models.SalesReturnSchemaName, models.SalesReturnFieldSalesOrderId, orderId)
	if err != nil {
		return decimal.Zero, err
	}
	for _, record := range returns {
		returnId := stringOf(record, models.SalesReturnFieldId)
		legs, err := searchBy(ctx, models.SalesRefundPaymentSchemaName, models.SalesRefundPaymentFieldSalesReturnId, returnId)
		if err != nil {
			return decimal.Zero, err
		}
		for _, leg := range legs {
			if stringOf(leg, models.SalesRefundPaymentFieldStatus) == string(models.SalesRefundPaymentStatusCompleted) {
				total = total.Sub(decimalOf(leg, models.SalesRefundPaymentFieldAmount))
			}
		}
		if returnId == exceptReturnId {
			continue
		}
		switch stringOf(record, models.SalesReturnFieldStatus) {
		case string(models.SalesReturnStatusApproved), string(models.SalesReturnStatusProcessing):
			if stringOf(record, models.SalesReturnFieldRefundStatus) != string(models.SalesReturnStepCompleted) {
				total = total.Sub(decimalOf(record, models.SalesReturnFieldLockedRefundAmount))
			}
		}
	}
	return total, nil
}

// normaliseNote trims and bounds a note the way the schema does, so a note that would fail the
// write is cut rather than refused at a point where the approval has already been decided.
func normaliseNote(note string) string {
	note = strings.TrimSpace(note)
	if len(note) > maxNoteLength {
		note = note[:maxNoteLength]
	}
	return note
}
