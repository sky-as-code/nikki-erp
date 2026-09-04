package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Actually giving the money back.
//
// The legs say how a refund is split across the payments it came in on; this is what moves it. Until
// it existed the legs were written `pending` and nothing ever changed them, so SumCompletedRefunds
// always saw zero, a return's refund step never reached `completed`, and an order stayed `paid`
// however much had been given back.
//
// MONEY RETURNS BY THE ROUTE IT ARRIVED. A leg names the payment it reverses, so a card refund goes
// back to that card through the same gateway and a cash refund is settled at the drawer. Refunding
// cash for a card payment would be an unexplained outflow that reconciliation could never match.

// SettleRefundLegsResult reports what one pass moved.
type SettleRefundLegsResult struct {
	Completed int
	Failed    int

	// StillPending counts legs left alone because nothing could act on them — a gateway leg with no
	// port bound, most often. They are retried on the next processing of the return.
	StillPending int
}

// SettleRefundLegs gives back the money for one return's legs.
func SettleRefundLegs(
	ctx corectx.Context,
	salesReturn dmodel.DynamicFields,
	orders itExt.PaymentOrderExtService,
) (*SettleRefundLegsResult, error) {
	result := &SettleRefundLegsResult{}

	legs, err := searchBy(ctx, models.SalesRefundPaymentSchemaName,
		models.SalesRefundPaymentFieldSalesReturnId,
		stringOf(salesReturn, models.SalesReturnFieldId))
	if err != nil {
		return nil, err
	}

	for _, leg := range legs {
		status := stringOf(leg, models.SalesRefundPaymentFieldStatus)
		if status != string(models.SalesRefundPaymentStatusPending) {
			// Already settled one way or the other. Skipped rather than re-sent: a second refund
			// against the same leg would give the money back twice.
			continue
		}

		outcome, err := settleOneRefundLeg(ctx, salesReturn, leg, orders)
		if err != nil {
			return nil, err
		}
		switch outcome {
		case refundCompleted:
			result.Completed++
		case refundFailed:
			result.Failed++
		default:
			result.StillPending++
		}
	}
	return result, nil
}

type refundLegOutcome int

const (
	refundStillPending refundLegOutcome = iota
	refundCompleted
	refundFailed
)

// settleOneRefundLeg returns one leg's money by the route it arrived on.
func settleOneRefundLeg(
	ctx corectx.Context,
	salesReturn dmodel.DynamicFields,
	leg dmodel.DynamicFields,
	orders itExt.PaymentOrderExtService,
) (refundLegOutcome, error) {
	original, err := loadRecord(ctx, models.SalesPaymentSchemaName,
		models.SalesPaymentFieldId,
		stringOf(leg, models.SalesRefundPaymentFieldOriginalSalesPaymentId))
	if err != nil {
		return refundStillPending, err
	}
	if original == nil {
		// The payment being reversed is gone. Failed rather than left pending: no retry will find
		// it, and a leg that waits forever hides a real problem.
		return refundFailed, failRefundLeg(ctx, leg,
			"the original payment no longer exists, so this refund cannot be routed")
	}

	paymentOrderId := stringOf(original, models.SalesPaymentFieldPaymentOrderId)
	if paymentOrderId == "" {
		// Cash. There is no provider to call: the money comes out of the drawer, and the leg records
		// that it was authorised to.
		return refundCompleted, completeRefundLeg(ctx, leg, "")
	}

	if orders == nil {
		// A gateway leg with no gateway. Left pending deliberately — the money genuinely has not
		// moved, and marking it complete would tell a customer they were repaid when they were not.
		return refundStillPending, nil
	}

	refunded, err := orders.Refund(ctx, itExt.RefundGatewayPaymentCommand{
		OrderId: paymentOrderId,
		Amount:  decimalOf(leg, models.SalesRefundPaymentFieldAmount),
		Content: "Refund for return " + stringOf(salesReturn, models.SalesReturnFieldReturnNumber),
	})
	if err != nil {
		// A transport failure, not a refusal: the gateway may have processed it. Left pending so the
		// next pass asks again — the over-refund guard upstream is what stops a double refund if it
		// did in fact go through.
		return refundStillPending, nil
	}

	if refunded.Refused || !refunded.HasData {
		reason := refunded.RefusalReason
		if reason == "" {
			reason = "the payment provider refused the refund"
		}
		return refundFailed, failRefundLeg(ctx, leg, reason)
	}
	return refundCompleted, completeRefundLeg(ctx, leg, refunded.Data.OrderId)
}

// completeRefundLeg records that the money is back with the customer.
func completeRefundLeg(
	ctx corectx.Context, leg dmodel.DynamicFields, providerReference string,
) error {
	changes := dmodel.DynamicFields{
		models.SalesRefundPaymentFieldStatus: string(
			models.SalesRefundPaymentStatusCompleted),
		models.SalesRefundPaymentFieldCompletedAt: model.ModelDateTime(time.Now().UTC()),
	}
	if providerReference != "" {
		changes[models.SalesRefundPaymentFieldProviderReference] = providerReference
	}
	return writeChanges(ctx, models.SalesRefundPaymentSchemaName, leg, changes)
}

// failRefundLeg records that this leg will not settle as it stands.
func failRefundLeg(ctx corectx.Context, leg dmodel.DynamicFields, reason string) error {
	return writeChanges(ctx, models.SalesRefundPaymentSchemaName, leg, dmodel.DynamicFields{
		models.SalesRefundPaymentFieldStatus: string(
			models.SalesRefundPaymentStatusFailed),
		models.SalesRefundPaymentFieldFailureReason: truncateError(reason),
	})
}
