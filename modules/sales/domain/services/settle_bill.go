package services

import (
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Settlement and payment status derivation (BR 42 and 76, SALES-028).
//
// # A bill settles only when captured == payable, EXACTLY
//
// BR 76 says so and the word matters. Not "within a rounding tolerance": the amounts are decimals at
// a fixed scale precisely so an exact comparison is meaningful, and a tolerance would let a bill be
// marked settled while a fraction remained owed. Those fractions accumulate across a day of trading
// into a number nobody can account for.
//
// # Status is DERIVED, never set
//
// Every payment status in this file is computed from the money, so it cannot drift from the payments
// that produced it. The derivation is the same one SALES-015 wrote for orders, reused rather than
// re-implemented - two implementations of "what does partially paid mean" would eventually disagree,
// and the disagreement would show up as a bill a customer was asked to pay twice.

// SettleBillResult is what a settlement attempt concluded.
type SettleBillResult struct {
	SalesBillId string

	Status        string
	PaymentStatus string

	CapturedTotal decimal.Decimal
	BillTotal     decimal.Decimal

	// Settled says whether this call moved the bill to settled. False when the money is not all in
	// yet, which is not a failure - it is the ordinary state of a partially paid bill.
	Settled bool
}

// SettleBillIfPaid moves a bill to settled when its money is fully in.
//
// Called after every payment rather than being a separate operator action: a bill whose last payment
// just landed is settled by that fact, and requiring somebody to press a button would leave bills
// sitting open with nothing owed on them.
//
// Idempotent. A bill already settled is answered with its current state rather than an error,
// because the caller that reports a payment cannot know whether an earlier one already closed it.
func SettleBillIfPaid(
	ctx corectx.Context, billId string,
) (*SettleBillResult, *ft.ClientErrors, error) {
	bill, err := loadRecord(ctx, models.SalesBillSchemaName, models.SalesBillFieldId, billId)
	if err != nil {
		return nil, nil, err
	}
	if bill == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("sales_bill_id", ReasonBillNotFound,
			"no bill exists with id '"+billId+"'"))
		return nil, vErrs, nil
	}

	captured, err := capturedTotalOf(ctx, billId)
	if err != nil {
		return nil, nil, err
	}
	billTotal := decimalOf(bill, models.SalesBillFieldTotalAmount)
	paymentStatus := DeriveBillPaymentStatus(billTotal, captured)

	current := models.NewSalesBillFrom(bill)
	if current.IsSettled() {
		// Already closed. Answered rather than refused, and deliberately not re-stamped: settled_at
		// records when the money arrived, and rewriting it on a later call would move the date of a
		// completed sale.
		return &SettleBillResult{
			SalesBillId:   billId,
			Status:        string(models.SalesBillStatusSettled),
			PaymentStatus: paymentStatus,
			CapturedTotal: captured,
			BillTotal:     billTotal,
		}, nil, nil
	}

	// Cancelled bills are left alone. One superseded by a split or a merge may still carry payments,
	// and settling it would close a bill whose value now lives somewhere else.
	if !current.IsOpen() {
		return &SettleBillResult{
			SalesBillId:   billId,
			Status:        stringOf(bill, models.SalesBillFieldStatus),
			PaymentStatus: paymentStatus,
			CapturedTotal: captured,
			BillTotal:     billTotal,
		}, nil, nil
	}

	update := dmodel.DynamicFields{
		models.SalesBillFieldId:            billId,
		models.SalesBillFieldPaymentStatus: paymentStatus,
	}

	// EXACT equality, per BR 76. An overpaid bill is settled too: the money owed is in, and the
	// excess is change the till hands back rather than a reason to keep the bill open.
	settled := !captured.LessThan(billTotal) && billTotal.IsPositive()
	if settled {
		update[models.SalesBillFieldStatus] = string(models.SalesBillStatusSettled)
		update[models.SalesBillFieldSettledAt] = model.ModelDateTime(time.Now().UTC())
	}

	engine, err := engineFor(models.SalesBillSchemaName)
	if err != nil {
		return nil, nil, err
	}
	if _, err := engine.ResourceRepository().Update(ctx, update); err != nil {
		return nil, nil, err
	}

	status := string(models.SalesBillStatusOpen)
	if settled {
		status = string(models.SalesBillStatusSettled)
	}
	return &SettleBillResult{
		SalesBillId:   billId,
		Status:        status,
		PaymentStatus: paymentStatus,
		CapturedTotal: captured,
		BillTotal:     billTotal,
		Settled:       settled,
	}, nil, nil
}

// DeriveBillPaymentStatus answers what a bill's payment status should be, from the money against it.
//
// Pure: it takes the numbers rather than reading them, so the rule can be tested exhaustively and the
// caller controls which transaction the numbers came from. The same shape as SALES-015's
// DerivePaymentStatus, and deliberately so - a bill and an order answer the same question about
// different scopes.
func DeriveBillPaymentStatus(payable, captured decimal.Decimal) string {
	switch {
	case captured.IsZero():
		return string(models.SalesOrderPaymentStatusUnpaid)
	case captured.LessThan(payable):
		return string(models.SalesOrderPaymentStatusPartiallyPaid)
	case captured.Equal(payable):
		return string(models.SalesOrderPaymentStatusPaid)
	default:
		// More money in than the bill is for. A real state (BR 42) rather than an error, because the
		// cash-change policy permits it - the excess is handed back as change.
		return string(models.SalesOrderPaymentStatusOverpaid)
	}
}

// DeriveOrderPaymentStatusFromBills aggregates an order's payment status from its bills.
//
// The order is the sum of its settlement units, so its status is derived from theirs rather than
// tracked separately. Cancelled bills are excluded for the same reason they are excluded from the
// BR 36 check: they were superseded, and their value lives on in whatever replaced them.
func DeriveOrderPaymentStatusFromBills(
	ctx corectx.Context, orderId string,
) (string, decimal.Decimal, error) {
	bills, err := billsOfOrder(ctx, orderId, false)
	if err != nil {
		return "", decimal.Zero, err
	}
	if len(bills) == 0 {
		// A sale with no bills has had nothing asked of the customer, so nothing is owed yet.
		return string(models.SalesOrderPaymentStatusUnpaid), decimal.Zero, nil
	}

	payable := decimal.Zero
	captured := decimal.Zero
	for _, bill := range bills {
		payable = payable.Add(decimalOf(bill, models.SalesBillFieldTotalAmount))

		billCaptured, err := capturedTotalOf(ctx, stringOf(bill, models.SalesBillFieldId))
		if err != nil {
			return "", decimal.Zero, err
		}
		captured = captured.Add(billCaptured)
	}
	return DeriveBillPaymentStatus(payable, captured), captured, nil
}

// SyncOrderPaymentStatus recomputes an order's payment status from its bills and stores it.
//
// Called after a payment settles a bill, so the order reflects its bills without a separate
// reconciliation job. A status that has not changed is still written: the write is cheap, and
// skipping it would mean the stored value depends on how many times this ran.
func SyncOrderPaymentStatus(ctx corectx.Context, orderId string) (string, error) {
	status, _, err := DeriveOrderPaymentStatusFromBills(ctx, orderId)
	if err != nil {
		return "", err
	}

	engine, err := engineFor(models.SalesOrderSchemaName)
	if err != nil {
		return "", err
	}
	_, err = engine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.SalesOrderFieldId:            orderId,
		models.SalesOrderFieldPaymentStatus: status,
	})
	return status, err
}
