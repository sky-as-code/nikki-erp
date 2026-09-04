package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Applying a gateway's verdict to the payment that was awaiting it.
//
// THIS RUNS AT LEAST ONCE, AND MAY RUN AGAIN FOR THE SAME VERDICT. Two independent paths lead here:
// the settlement announcement, which is fast but may be lost because the bus acknowledges a message
// before it is handled, and the reconciliation sweep, which is slow but cannot be lost. Both are
// wanted — the first for the customer standing at the till, the second so nothing is stranded — and
// the price is that this must be safe to repeat.
//
// It is made safe by refusing to move a payment that has already reached a terminal state. A verdict
// arriving twice finds the payment captured the second time and does nothing; a late failure
// arriving after a success does not un-pay a customer who has paid and whose goods have gone.

// ConfirmPaymentOutcome is what the verdict said.
type ConfirmPaymentOutcome string

const (
	ConfirmPaymentPaid     ConfirmPaymentOutcome = "paid"
	ConfirmPaymentFailed   ConfirmPaymentOutcome = "failed"
	ConfirmPaymentExpired  ConfirmPaymentOutcome = "expired"
	ConfirmPaymentCanceled ConfirmPaymentOutcome = "canceled"
)

// ConfirmPaymentParams names the payment and what became of it.
type ConfirmPaymentParams struct {
	// PaymentOrderId is the correlation the order was opened under, and the usual way in.
	PaymentOrderId string

	// SalesPaymentId is the fallback the announcement carries in its metadata, for the case where
	// the order was opened but the correlation column was never written.
	SalesPaymentId string

	Outcome ConfirmPaymentOutcome

	// RefTransactionId is the gateway's identifier for the completed payment. It lands in
	// external_transaction_id, whose partial unique index is what stops the same money being
	// recorded twice against one bill.
	RefTransactionId string
}

// ConfirmPaymentResult reports what this call did, so a caller can tell a verdict it applied from
// one that had already been applied.
type ConfirmPaymentResult struct {
	SalesPaymentId string
	SalesBillId    string

	// Applied is false when the payment was already terminal, or when no payment matched. Neither is
	// a failure: the first is a replay, the second is a verdict for an order Sales never opened.
	Applied bool

	// Found is false only when nothing matched, which is worth reporting separately — a settlement
	// with no payment behind it means money moved that Sales cannot account for.
	Found bool

	Status string
}

// ConfirmPayment applies a gateway verdict to the payment awaiting it.
func ConfirmPayment(
	ctx corectx.Context, params ConfirmPaymentParams,
) (*ConfirmPaymentResult, error) {
	payment, err := findPaymentAwaitingSettlement(ctx, params)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return &ConfirmPaymentResult{Found: false}, nil
	}

	current := models.NewSalesPaymentFrom(payment)
	salesPaymentId := stringOf(payment, models.SalesPaymentFieldId)
	billId := stringOf(payment, models.SalesPaymentFieldSalesBillId)

	if current.IsTerminal() {
		// A replay. Answered rather than refused: the verdict is already recorded, and the caller —
		// a bus subscriber or a sweep — has nothing to fix.
		return &ConfirmPaymentResult{
			SalesPaymentId: salesPaymentId,
			SalesBillId:    billId,
			Found:          true,
			Applied:        false,
			Status:         stringOf(payment, models.SalesPaymentFieldStatus),
		}, nil
	}

	status := paymentStatusFor(params.Outcome)
	changes := dmodel.DynamicFields{models.SalesPaymentFieldStatus: status}

	if status == string(models.SalesPaymentStatusCaptured) {
		changes[models.SalesPaymentFieldPaidAt] = model.ModelDateTime(time.Now().UTC())
		if params.RefTransactionId != "" {
			changes[models.SalesPaymentFieldExternalTransactionId] = params.RefTransactionId
		}
	}

	if err := writeChanges(ctx, models.SalesPaymentSchemaName, payment, changes); err != nil {
		return nil, err
	}

	if status == string(models.SalesPaymentStatusCaptured) {
		// Announced only on capture, matching RecordPayment: an authorization is a hold the provider
		// may still release, and telling consumers money arrived when it has not is the mistake
		// SumCapturedAmount exists to avoid.
		if err := announceCapturedPayment(ctx, payment, salesPaymentId, billId); err != nil {
			return nil, err
		}
	}

	return &ConfirmPaymentResult{
		SalesPaymentId: salesPaymentId,
		SalesBillId:    billId,
		Found:          true,
		Applied:        true,
		Status:         status,
	}, nil
}

// ConfirmPaymentAndSettle applies the verdict and then closes the bill if the money is now all in.
//
// Settlement is deliberately a second step rather than part of the write above: a bill is settled by
// the total of its captured payments, not by any one of them, and a split tender is only finished
// when its last leg lands.
func ConfirmPaymentAndSettle(
	ctx corectx.Context, params ConfirmPaymentParams,
) (*ConfirmPaymentResult, error) {
	result, err := ConfirmPayment(ctx, params)
	if err != nil || result == nil || !result.Applied {
		return result, err
	}
	if result.Status != string(models.SalesPaymentStatusCaptured) {
		return result, nil
	}

	if _, _, err := SettleBillIfPaid(ctx, result.SalesBillId); err != nil {
		return nil, err
	}
	return result, nil
}

// paymentStatusFor maps a verdict onto the payment's own vocabulary.
//
// Cancelled and failed are kept apart because they free the method slot for different reasons: a
// declined card is the customer's problem to retry, a cancelled order is the till's.
func paymentStatusFor(outcome ConfirmPaymentOutcome) string {
	switch outcome {
	case ConfirmPaymentPaid:
		return string(models.SalesPaymentStatusCaptured)
	case ConfirmPaymentCanceled:
		return string(models.SalesPaymentStatusCancelled)
	}
	// Failed and expired both mean the collection is over without the money: an expiry is a failure
	// with a reason, and Sales has no separate state for it.
	return string(models.SalesPaymentStatusFailed)
}

// findPaymentAwaitingSettlement locates the payment a verdict belongs to.
//
// The correlation column is tried first because it is the one written when the order was opened. The
// id from the announcement's metadata is the fallback for the narrow case where the order was opened
// and the process died before the column was written — the money is real, and it would otherwise be
// unaccounted for.
func findPaymentAwaitingSettlement(
	ctx corectx.Context, params ConfirmPaymentParams,
) (dmodel.DynamicFields, error) {
	if params.PaymentOrderId != "" {
		found, err := searchBy(ctx, models.SalesPaymentSchemaName,
			models.SalesPaymentFieldPaymentOrderId, params.PaymentOrderId)
		if err != nil {
			return nil, err
		}
		if len(found) > 0 {
			return found[0], nil
		}
	}

	if params.SalesPaymentId != "" {
		return loadRecord(ctx,
			models.SalesPaymentSchemaName, models.SalesPaymentFieldId, params.SalesPaymentId)
	}
	return nil, nil
}

// announceCapturedPayment writes the outbox event a captured payment produces, the same one
// RecordPayment writes when money is taken at the counter.
//
// The two paths produce ONE event type on purpose: a downstream consumer cares that a bill was paid,
// not whether the customer tapped a card or handed over notes.
func announceCapturedPayment(
	ctx corectx.Context, payment dmodel.DynamicFields, salesPaymentId, billId string,
) error {
	bill, err := loadRecord(ctx, models.SalesBillSchemaName, models.SalesBillFieldId, billId)
	if err != nil {
		return err
	}

	orderId := ""
	if bill != nil {
		orderId = stringOf(bill, models.SalesBillFieldSalesOrderId)
	}

	_, err = RecordEvent(ctx, RecordEventParams{
		EventType:   models.EventSalesPaymentCaptured,
		AggregateId: billId,
		OrgId:       stringOf(payment, basemodel.FieldOrgId),
		Payload: map[string]any{
			"sales_payment_id":  salesPaymentId,
			"sales_bill_id":     billId,
			"sales_order_id":    orderId,
			"payment_method_id": stringOf(payment, models.SalesPaymentFieldPaymentMethodId),
			"amount":            decimalOf(payment, models.SalesPaymentFieldAmount),
			"currency_code":     stringOf(payment, models.SalesPaymentFieldCurrencyCode),
		},
	})
	return err
}
