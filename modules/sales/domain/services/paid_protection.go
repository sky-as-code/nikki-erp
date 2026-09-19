package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Protecting a paid order's holds (CR-INV-SALES-WH-RESERVATION §7.2, §7.3).
//
// Once the money is in, the deadline that protected the shop from an abandoned order protects
// nobody, so every warehouse hold of the order has its deadline cleared. The order's own
// valid_until is left as it was, for audit; it no longer binds because the paid latch is on the
// payment status. A hold that lapsed before the payment was recorded is not revived: the stock may
// already be somebody else's. It is held again under a new revision if the warehouse can cover
// it, and otherwise the order is cancelled by the system and a refund request raised for what
// was collected, per the order's own snapshot.

type ProtectPaidOrderResult struct {
	SalesOrderId string

	ProtectedFulfillmentIds []string
	ReheldFulfillmentIds    []string

	// Cancelled reports that the goods could not be held again and the order was called off.
	Cancelled bool
	RefundId  string
}

// ProtectPaidOrder runs after a payment settled as paid, outside the settlement transaction.
func ProtectPaidOrder(
	ctx corectx.Context,
	billId, paymentReference string,
	dLock lock.DistributedLock,
	reservations itExt.FulfillmentReservationExtService,
	policy SalesPolicy,
	refundDeps RefundProcessingDeps,
) (*ProtectPaidOrderResult, error) {
	port := warehouseHolds
	if port == nil {
		return nil, nil
	}
	bill, err := loadRecord(ctx, models.SalesBillSchemaName, models.SalesBillFieldId, billId)
	if err != nil || bill == nil {
		return nil, err
	}
	orderId := stringOf(bill, models.SalesBillFieldSalesOrderId)
	if orderId == "" {
		return nil, nil
	}
	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil || order == nil {
		return nil, err
	}
	switch stringOf(order, models.SalesOrderFieldPaymentStatus) {
	case string(models.SalesOrderPaymentStatusPaid), string(models.SalesOrderPaymentStatusOverpaid):
	default:
		// Partially paid: not yet a paid order, and its holds keep their deadline.
		return nil, nil
	}
	if stringOf(order, models.SalesOrderFieldStatus) == string(models.SalesOrderStatusCancelled) {
		// Money for a sale already called off: recorded by the settlement, handled by the refund
		// workflow. Nothing here may reopen it.
		return nil, nil
	}

	result := &ProtectPaidOrderResult{SalesOrderId: orderId}
	orgId := stringOf(order, models.SalesOrderFieldOrgId)
	fulfillments, err := FulfillmentsOfOrder(ctx, orderId)
	if err != nil {
		return nil, err
	}
	for _, record := range fulfillments {
		fulfillment := models.NewSalesOrderFulfillmentFrom(record)
		if fulfillment.IsTerminal() || !fulfillment.ExecutesKioskDispense() {
			continue
		}
		fulfillmentId := stringOf(record, models.SalesOrderFulfillmentFieldId)

		protected, err := port.ProtectPaidReservations(ctx, itExt.ProtectPaidReservationsRequest{
			OrgId:            model.Id(orgId),
			SourceModule:     itExt.SourceModuleSales,
			SourceType:       itExt.SourceTypeOrderFulfillment,
			SourceId:         fulfillmentId,
			SourceRevision:   firstSourceRevision,
			PaymentReference: paymentReference,
			IdempotencyKey:   "paid:" + fulfillmentId + ":" + paymentReference,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "protecting the holds of fulfillment %s", fulfillmentId)
		}
		if protected != nil && protected.Refused() {
			return nil, errors.New("inventory refused to protect fulfillment " + fulfillmentId + ": " +
				describeFirstViolation(&protected.ClientErrors))
		}

		if protected != nil && len(protected.Expired) == 0 {
			if err := clearFulfillmentDeadline(ctx, record); err != nil {
				return nil, err
			}
			result.ProtectedFulfillmentIds = append(result.ProtectedFulfillmentIds, fulfillmentId)
			continue
		}

		reheld, vErrs, err := reholdLapsedFulfillment(ctx, order, record, port)
		if err != nil {
			return nil, err
		}
		if reheld {
			result.ReheldFulfillmentIds = append(result.ReheldFulfillmentIds, fulfillmentId)
			continue
		}

		// The goods are gone. The order is called off by the system and the money collected is
		// returned through a refund request; whether that request confirms itself follows the
		// order's snapshot, never today's channel setting.
		if err := setFulfillmentStatus(ctx, fulfillmentId, models.FulfillmentStatusExpired); err != nil {
			return nil, err
		}
		reason := "stock could not be held after payment"
		if vErrs != nil {
			reason = describeFirstViolation(vErrs)
		}
		cancelled, cancelErrs, err := CancelOrderWith(ctx, orderId, CancelOrderOptions{
			Reason:          reason,
			SystemInitiated: true,
			Policy:          policy,
			RefundDeps:      refundDeps,
		}, dLock, reservations)
		if err != nil {
			return nil, err
		}
		if cancelErrs != nil {
			return nil, errors.New("the order could not be cancelled after its stock lapsed: " +
				describeFirstViolation(cancelErrs))
		}
		result.Cancelled = true
		result.RefundId = cancelled.RefundRequestId
		return result, nil
	}
	return result, nil
}

// clearFulfillmentDeadline drops the fulfillment's own copy of the deadline once the hold behind
// it no longer lapses, so the attempt gate stops refusing a paid customer.
func clearFulfillmentDeadline(ctx corectx.Context, fulfillment dmodel.DynamicFields) error {
	if dateTimeOf(fulfillment, models.SalesOrderFulfillmentFieldReservationExpiresAt) == nil {
		return nil
	}
	return writeChanges(ctx, models.SalesOrderFulfillmentSchemaName, fulfillment, dmodel.DynamicFields{
		models.SalesOrderFulfillmentFieldReservationExpiresAt: nil,
	})
}

// reholdLapsedFulfillment tries a fresh all-or-nothing hold, without a deadline, under the next
// source revision. False with violations means the warehouse cannot cover it now.
func reholdLapsedFulfillment(
	ctx corectx.Context, order, fulfillment dmodel.DynamicFields, port itExt.WarehouseReservationExtService,
) (bool, *ft.ClientErrors, error) {
	orgId := stringOf(order, models.SalesOrderFieldOrgId)
	fulfillmentId := stringOf(fulfillment, models.SalesOrderFulfillmentFieldId)

	current, err := currentSourceRevision(ctx, orgId, fulfillmentId, port)
	if err != nil {
		return false, nil, err
	}
	if current >= maxSourceRevision {
		return false, nil, nil
	}

	items, err := ItemsOfFulfillment(ctx, fulfillmentId)
	if err != nil {
		return false, nil, err
	}
	lines, itemIds := fulfillmentHoldLines(items)
	if len(lines) == 0 {
		return true, nil, nil
	}

	resolved := &ResolvedFulfillmentMethod{
		TargetLocationId: targetLocationOfFulfillment(items),
	}
	if resolved.TargetLocationId == "" {
		return false, nil, nil
	}
	vErrs, err := reserveAtWarehouse(ctx, fulfillmentId, orgId, resolved, lines, itemIds, nil, current+1, port)
	if err != nil {
		return false, nil, err
	}
	if vErrs != nil && vErrs.Count() > 0 {
		return false, vErrs, nil
	}
	if err := clearFulfillmentDeadline(ctx, fulfillment); err != nil {
		return false, nil, err
	}
	return true, nil, setFulfillmentStatus(ctx, fulfillmentId, models.FulfillmentStatusReserved)
}

// targetLocationOfFulfillment reads the location the items were held against; every item of a
// kiosk fulfillment names the same machine.
func targetLocationOfFulfillment(items []dmodel.DynamicFields) string {
	for _, item := range items {
		if location := stringOf(item, models.SalesOrderFulfillmentItemFieldSourceLocationId); location != "" {
			return location
		}
	}
	return ""
}

// The protector hook: installed by app/ with its ports, called by every settlement path after
// its transaction committed, so the reconciliation backstop protects exactly as the live event
// handler does. A build that installs nothing keeps the deadline on its holds.
var paidOrderProtector func(ctx corectx.Context, billId, paymentReference string) error

func SetPaidOrderProtector(protector func(ctx corectx.Context, billId, paymentReference string) error) {
	paidOrderProtector = protector
}

// NotifyOrderPaid runs the installed protector; a nil one is a working system without protection.
func NotifyOrderPaid(ctx corectx.Context, billId, paymentReference string) error {
	if paidOrderProtector == nil || ctx.GetDbTranx() != nil {
		return nil
	}
	return paidOrderProtector(ctx, billId, paymentReference)
}
