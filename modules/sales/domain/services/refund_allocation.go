package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Splitting a refund across the payments that made the sale.
//
// Money returns by the route it arrived: every leg names an original payment and is capped at what
// that payment captured, so a customer who paid 200,000 in cash and 300,000 by card cannot be
// refunded 300,000 in cash. The split is proportional to what each payment captured, returning each
// method the same share of the sale it funded.

// RefundLeg is one payment's share of a refund.
type RefundLeg struct {
	OriginalPaymentId string
	Amount            decimal.Decimal
	CurrencyCode      string
}

// capturedPayment carries what one original payment can absorb.
type capturedPayment struct {
	Id           string
	Captured     decimal.Decimal
	CurrencyCode string
}

// allocateRefund splits a refund across the captured payments, proportionally and capped.
//
// Proportional shares rounded independently do not sum to the total, so the residual lands on the
// largest payment, where a rounding unit is proportionally least visible, and lands deterministically
// so the same refund always splits the same way.
//
// Capping runs after allocation and can leave a shortfall: if every payment is capped and the total
// is still unmet, the remainder is simply not allocated, and the caller sees legs summing to less
// than the total.
func allocateRefund(total decimal.Decimal, payments []capturedPayment) []RefundLeg {
	if !total.IsPositive() || len(payments) == 0 {
		return nil
	}

	capturedTotal := decimal.Zero
	for _, payment := range payments {
		capturedTotal = capturedTotal.Add(payment.Captured)
	}
	if !capturedTotal.IsPositive() {
		return nil
	}

	// Never refund more than was taken, whatever the return is worth.
	if total.GreaterThan(capturedTotal) {
		total = capturedTotal
	}

	legs := make([]RefundLeg, 0, len(payments))
	allocated := decimal.Zero
	largestIndex := 0

	for index, payment := range payments {
		share := total.Mul(payment.Captured).Div(capturedTotal).Round(refundScale)
		if share.GreaterThan(payment.Captured) {
			share = payment.Captured
		}
		legs = append(legs, RefundLeg{
			OriginalPaymentId: payment.Id,
			Amount:            share,
			CurrencyCode:      payment.CurrencyCode,
		})
		allocated = allocated.Add(share)

		if payment.Captured.GreaterThan(payments[largestIndex].Captured) {
			largestIndex = index
		}
	}

	// Push the rounding residual onto the largest payment, up to its own cap.
	if residual := total.Sub(allocated); !residual.IsZero() {
		room := payments[largestIndex].Captured.Sub(legs[largestIndex].Amount)
		if residual.GreaterThan(room) {
			residual = room
		}
		legs[largestIndex].Amount = legs[largestIndex].Amount.Add(residual)
	}

	// Drop zero legs: a refund of nothing against a payment is a row that explains nothing.
	kept := make([]RefundLeg, 0, len(legs))
	for _, leg := range legs {
		if leg.Amount.IsPositive() {
			kept = append(kept, leg)
		}
	}
	return kept
}

// refundScale is the money scale refunds are rounded to: four places, matching every money column in
// the module's schemas. Not the org's display rounding — a coarser scale would make the legs disagree
// with the total they came from.
const refundScale = int32(4)

// capturedPaymentsOfOrder reads only captured payments: an authorized one is a hold the provider may
// still release, and refunding against it would send back money that never arrived.
func capturedPaymentsOfOrder(ctx corectx.Context, orderId string) ([]capturedPayment, error) {
	bills, err := searchBy(ctx, models.SalesBillSchemaName, models.SalesBillFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	captured := make([]capturedPayment, 0, len(bills))
	for _, bill := range bills {
		payments, err := searchBy(ctx, models.SalesPaymentSchemaName,
			models.SalesPaymentFieldSalesBillId, stringOf(bill, models.SalesBillFieldId))
		if err != nil {
			return nil, err
		}
		for _, payment := range payments {
			if stringOf(payment, models.SalesPaymentFieldStatus) != string(models.SalesPaymentStatusCaptured) {
				continue
			}
			captured = append(captured, capturedPayment{
				Id:           stringOf(payment, models.SalesPaymentFieldId),
				Captured:     decimalOf(payment, models.SalesPaymentFieldAmount),
				CurrencyCode: stringOf(payment, models.SalesPaymentFieldCurrencyCode),
			})
		}
	}
	return captured, nil
}

// issuedBillOfOrder finds a bill of the order that carries an issued fiscal document. Nil when no
// invoice was ever issued, which is not a failure: a return before invoicing has nothing to adjust,
// and the pending request is cancelled instead.
func issuedBillOfOrder(ctx corectx.Context, orderId string) (dmodel.DynamicFields, error) {
	bills, err := searchBy(ctx, models.SalesBillSchemaName, models.SalesBillFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	for _, bill := range bills {
		requests, err := searchBy(ctx, models.SalesFiscalRequestSchemaName,
			models.SalesFiscalRequestFieldSalesBillId, stringOf(bill, models.SalesBillFieldId))
		if err != nil {
			return nil, err
		}
		for _, request := range requests {
			if stringOf(request, models.SalesFiscalRequestFieldStatus) == string(models.SalesFiscalStatusIssued) {
				return bill, nil
			}
		}
	}
	return nil, nil
}

// writeRefundLegs stores the allocation in one transaction, all legs or none: a partial write would
// refund some of what is owed and leave the rest with no record that it was ever due.
func writeRefundLegs(
	ctx corectx.Context, salesReturn dmodel.DynamicFields, legs []RefundLeg,
) error {
	if len(legs) == 0 {
		return nil
	}

	returnId := stringOf(salesReturn, models.SalesReturnFieldId)
	orgId := orgIdOf(salesReturn)

	return withTransaction(ctx, models.SalesRefundPaymentSchemaName, func(tranxCtx corectx.Context) error {
		engine, err := engineFor(models.SalesRefundPaymentSchemaName)
		if err != nil {
			return err
		}
		for _, leg := range legs {
			id, err := model.NewId()
			if err != nil {
				return err
			}
			if _, err := engine.ResourceRepository().Insert(tranxCtx, dmodel.DynamicFields{
				models.SalesRefundPaymentFieldId:                     string(*id),
				models.SalesRefundPaymentFieldOrgId:                  orgId,
				models.SalesRefundPaymentFieldSalesReturnId:          returnId,
				models.SalesRefundPaymentFieldOriginalSalesPaymentId: leg.OriginalPaymentId,
				models.SalesRefundPaymentFieldAmount:                 leg.Amount,
				models.SalesRefundPaymentFieldCurrencyCode:           leg.CurrencyCode,
				models.SalesRefundPaymentFieldStatus:                 string(models.SalesRefundPaymentStatusPending),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}
