package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The payment deadline (CR-INV-SALES-WH-RESERVATION §7.1).
//
// An order not successfully paid by valid_until is expired. Expiry is not a stage: the order keeps
// its status and every client-initiated mutation except cancel is refused from then on. The
// decision is made on the clock, here, on every gate; expired_at is written when a request first
// notices, and the sweep writes it for the rest, but nothing waits for that column. A partial
// payment does not save the order. A refund on a paid order does not expire it: the paid latch is
// on the payment status, and the refunded statuses are not the unpaid ones.

const (
	ReasonOrderExpired          = "sales_order.expired"
	ReasonValidUntilNotInFuture = "sales_order.valid_until_not_in_future"
)

// IsOrderExpired reports whether the deadline has passed without a successful payment.
func IsOrderExpired(record dmodel.DynamicFields, now time.Time) bool {
	if dateTimeOf(record, models.SalesOrderFieldExpiredAt) != nil {
		return true
	}
	validUntil := dateTimeOf(record, models.SalesOrderFieldValidUntil)
	if validUntil == nil || validUntil.GoTime().After(now) {
		return false
	}
	switch stringOf(record, models.SalesOrderFieldPaymentStatus) {
	case string(models.SalesOrderPaymentStatusUnpaid), string(models.SalesOrderPaymentStatusPartiallyPaid):
		return true
	}
	return false
}

// assertOrderNotExpired refuses a client mutation of an expired order and records the expiry on
// the row while it is here, so a late callback can see the order was already past saving.
func assertOrderNotExpired(ctx corectx.Context, record dmodel.DynamicFields) (*ft.ClientErrors, error) {
	now := time.Now().UTC()
	if !IsOrderExpired(record, now) {
		return nil, nil
	}
	if err := materializeOrderExpiry(ctx, record, now); err != nil {
		return nil, err
	}
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("valid_until", ReasonOrderExpired,
		"this order's payment deadline has passed; it can only be cancelled"))
	return vErrs, nil
}

// assertOrderNotExpiredById is the same gate for a caller that holds only the order id.
func assertOrderNotExpiredById(ctx corectx.Context, orderId string) (*ft.ClientErrors, error) {
	record, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil || record == nil {
		return nil, err
	}
	return assertOrderNotExpired(ctx, record)
}

// materializeOrderExpiry writes expired_at once. Through writeChanges: the column is closed to
// clients and this is its only writer besides the sweep.
func materializeOrderExpiry(ctx corectx.Context, record dmodel.DynamicFields, now time.Time) error {
	if dateTimeOf(record, models.SalesOrderFieldExpiredAt) != nil {
		return nil
	}
	if stringOf(record, models.SalesOrderFieldStatus) == string(models.SalesOrderStatusCancelled) {
		return nil
	}
	return writeChanges(ctx, models.SalesOrderSchemaName, record, dmodel.DynamicFields{
		models.SalesOrderFieldExpiredAt: model.ModelDateTime(now),
	})
}

// assertValidUntilInFuture refuses a deadline already passed at create.
func assertValidUntilInFuture(validUntil *time.Time, now time.Time) *ft.ClientErrors {
	if validUntil == nil || validUntil.After(now) {
		return nil
	}
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("valid_until", ReasonValidUntilNotInFuture,
		"valid_until must lie in the future"))
	return vErrs
}

// ExpireUnpaidOrders is the sweep's bookkeeping: it stamps expired_at on up to limit orders whose
// deadline passed unpaid and that no request has yet noticed. Availability and the gates never
// wait for it.
func ExpireUnpaidOrders(ctx corectx.Context, now time.Time, limit int) (int, error) {
	engineRepo, err := repoFor(models.SalesOrderSchemaName)
	if err != nil {
		return 0, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.SalesOrderFieldValidUntil, dmodel.LessEqual, model.ModelDateTime(now.UTC())),
		*dmodel.NewSearchNode().NewCondition(models.SalesOrderFieldExpiredAt, dmodel.IsNotSet),
		*dmodel.NewSearchNode().NewCondition(models.SalesOrderFieldPaymentStatus, dmodel.In,
			string(models.SalesOrderPaymentStatusUnpaid), string(models.SalesOrderPaymentStatusPartiallyPaid)),
		*dmodel.NewSearchNode().NewCondition(models.SalesOrderFieldStatus, dmodel.NotIn,
			string(models.SalesOrderStatusCancelled), string(models.SalesOrderStatusCompleted)),
	)
	found, err := engineRepo.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: limit})
	if err != nil {
		return 0, err
	}
	if found == nil || !found.HasData {
		return 0, nil
	}
	expired := 0
	for _, record := range found.Data.Items {
		if !IsOrderExpired(record, now) {
			continue
		}
		if err := materializeOrderExpiry(ctx, record, now); err != nil {
			return expired, err
		}
		expired++
	}
	return expired, nil
}

// isOrderExpiredNow is the gate for a caller that already validates into a list.
func isOrderExpiredNow(record dmodel.DynamicFields) bool {
	return IsOrderExpired(record, time.Now().UTC())
}

const (
	ReasonOrderNotPaid   = "sales.fulfillment.order_not_paid"
	ReasonOrderCancelled = "sales.fulfillment.order_cancelled"
)

// assertOrderDispensable is the pre-flight before a machine is commanded: the order is paid, not
// cancelled, and inside its deadline. Only a paid order's goods leave a kiosk.
func assertOrderDispensable(ctx corectx.Context, orderId string) (*ft.ClientErrors, error) {
	record, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil || record == nil {
		return nil, err
	}
	refuse := func(field, reason, message string) *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(field, reason, message))
		return vErrs
	}
	if stringOf(record, models.SalesOrderFieldStatus) == string(models.SalesOrderStatusCancelled) {
		return refuse("status", ReasonOrderCancelled, "this order was cancelled; nothing may be dispensed for it"), nil
	}
	if vErrs, err := assertOrderNotExpired(ctx, record); err != nil || vErrs != nil {
		return vErrs, err
	}
	if !isPaid(record) {
		return refuse("payment_status", ReasonOrderNotPaid, "this order is not paid; nothing may be dispensed for it"), nil
	}
	return nil, nil
}
