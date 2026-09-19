package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Holding a kiosk sale's stock at warehouse level (CR-INV-SALES-WH-RESERVATION §6.2, §9).
//
// The port is a package-level hook rather than a parameter threaded through every confirm, cancel,
// expiry and reassignment path, for the same reason the outbox drain is: those are package
// functions with no container behind them, and a build without the port simply keeps the location
// holds it had. When the port is installed, every kiosk fulfillment holds its goods at the
// warehouse behind the sales point, and its items carry the reservation ids in
// inventory_reservation_ref.

var warehouseHolds itExt.WarehouseReservationExtService

// SetWarehouseReservationPort installs the port, from the module's OnAppStarted.
func SetWarehouseReservationPort(port itExt.WarehouseReservationExtService) {
	warehouseHolds = port
}

// WarehouseReservationPort answers the installed port, or nil in a build without one.
func WarehouseReservationPort() itExt.WarehouseReservationExtService {
	return warehouseHolds
}

const (
	ReasonSalesPointNoWarehouse = "sales_order.sales_point_no_warehouse"

	// firstSourceRevision is the demand revision a fresh fulfillment reserves under. A re-reserve
	// after the first hold lapsed uses the next one, so a retry of the old request is never
	// mistaken for the new need.
	firstSourceRevision = int32(1)
)

// holdDeadline is when a kiosk sale's hold lapses: the order's payment deadline when it has one,
// passed through unchanged (no grace), else the method's TTL. A paid order's holds are protected
// afterwards, which is what makes a deadline safe to pass at all.
func holdDeadline(validUntil *model.ModelDateTime, ttlMinutes *int32, now time.Time) *model.ModelDateTime {
	if validUntil != nil {
		return validUntil
	}
	return reservationDeadline(ttlMinutes, now)
}

// reserveAtWarehouse holds every line of a fulfillment at the warehouse behind its target, all or
// nothing, and stamps each item with its reservation id.
func reserveAtWarehouse(
	ctx corectx.Context,
	fulfillmentId, orgId string,
	resolved *ResolvedFulfillmentMethod,
	lines []itExt.FulfillmentLine,
	itemIds []string,
	deadline *model.ModelDateTime,
	revision int32,
	port itExt.WarehouseReservationExtService,
) (*ft.ClientErrors, error) {
	warehouseId, err := port.ResolveWarehouseOfLocation(ctx, model.Id(orgId), model.Id(resolved.TargetLocationId))
	if err != nil {
		return nil, errors.Wrapf(err, "resolving the warehouse of fulfillment %s", fulfillmentId)
	}
	if warehouseId == "" {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(models.SalesOrderFulfillmentSchemaName, ReasonSalesPointNoWarehouse,
			"the sales point's location belongs to no warehouse, so nothing can be held for it"))
		return vErrs, nil
	}

	request := itExt.ReserveWarehouseStockRequest{
		OrgId:          model.Id(orgId),
		WarehouseId:    warehouseId,
		ReservedUntil:  itExt.DeadlineOrNil(deadline),
		SourceModule:   itExt.SourceModuleSales,
		SourceType:     itExt.SourceTypeOrderFulfillment,
		SourceId:       fulfillmentId,
		SourceRevision: revision,
		IdempotencyKey: warehouseHoldKey(fulfillmentId, revision),
	}
	for index, line := range lines {
		request.Lines = append(request.Lines, itExt.WarehouseReservationLine{
			SourceLineId:     itemIds[index],
			ProductVariantId: model.Id(line.ProductVariantId),
			UomId:            model.Id(line.UomId),
			Quantity:         line.Quantity,
		})
	}

	result, err := port.ReserveWarehouseStock(ctx, request)
	if err != nil {
		return nil, errors.Wrapf(err, "reserving warehouse stock for fulfillment %s", fulfillmentId)
	}
	if result == nil {
		return reservationRefusal(nil, ReasonFulfillmentReserveFailed), nil
	}
	if result.Refused() {
		reason := ReasonFulfillmentReserveFailed
		if hasReason(&result.ClientErrors, "stock_reservation.insufficient_warehouse_stock") {
			reason = ReasonFulfillmentInsufficientStock
		}
		return reservationRefusal(&itExt.FulfillmentReservationResponse{
			FailureReason: describeFirstViolation(&result.ClientErrors),
		}, reason), nil
	}

	for _, reserved := range result.Reservations {
		if err := stampReservationRefs(ctx, []string{reserved.SourceLineId}, string(reserved.ReservationId)); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// warehouseHoldKey is the idempotency key of a fulfillment's reserve request: one per demand
// revision, so a retried confirm replays and a re-reserve after expiry does not.
func warehouseHoldKey(fulfillmentId string, revision int32) string {
	return "fulfillment:" + fulfillmentId + ":rev:" + itoa32(revision)
}

func itoa32(value int32) string {
	digits := "0123456789"
	if value == 0 {
		return "0"
	}
	out := ""
	for value > 0 {
		out = string(digits[value%10]) + out
		value /= 10
	}
	return out
}

// releaseWarehouseHolds gives back every warehouse hold a fulfillment has, whatever revision it
// was taken under. A hold already gone answers success, so this is safe to run beside the
// location-hold release and on every retry.
func releaseWarehouseHolds(ctx corectx.Context, orgId, fulfillmentId string) error {
	port := warehouseHolds
	if port == nil || orgId == "" {
		return nil
	}
	for revision := firstSourceRevision; revision <= maxSourceRevision; revision++ {
		holds, err := port.ReservationsOfSource(ctx, model.Id(orgId),
			itExt.SourceModuleSales, itExt.SourceTypeOrderFulfillment, fulfillmentId, revision)
		if err != nil {
			return errors.Wrapf(err, "listing the warehouse holds of fulfillment %s", fulfillmentId)
		}
		if len(holds) == 0 {
			return nil
		}
		for _, hold := range holds {
			if hold.Status != "active" {
				continue
			}
			released, err := port.ReleaseReservation(ctx, itExt.ReleaseReservationRequest{
				OrgId:         model.Id(orgId),
				ReservationId: hold.ReservationId,
				Reason:        "sales_fulfillment_released",
			})
			if err != nil {
				return errors.Wrapf(err, "releasing warehouse hold %s", hold.ReservationId)
			}
			if released != nil && released.Refused() {
				return errors.New("inventory refused to release hold " + string(hold.ReservationId) + ": " +
					describeFirstViolation(&released.ClientErrors))
			}
		}
	}
	return nil
}

// maxSourceRevision bounds how many re-reserves one fulfillment may go through; a demand that
// lapsed and was re-held more often than this is an operational problem, not a loop to run.
const maxSourceRevision = int32(8)

// currentSourceRevision finds the highest revision a fulfillment holds anything under, or zero.
func currentSourceRevision(ctx corectx.Context, orgId, fulfillmentId string, port itExt.WarehouseReservationExtService) (int32, error) {
	current := int32(0)
	for revision := firstSourceRevision; revision <= maxSourceRevision; revision++ {
		holds, err := port.ReservationsOfSource(ctx, model.Id(orgId),
			itExt.SourceModuleSales, itExt.SourceTypeOrderFulfillment, fulfillmentId, revision)
		if err != nil {
			return 0, err
		}
		if len(holds) == 0 {
			return current, nil
		}
		current = revision
	}
	return current, nil
}

// fulfillmentHoldLines rebuilds the reserve lines of a fulfillment from its stored items, for a
// re-reserve after the first hold lapsed. Only what is still owed is held again.
func fulfillmentHoldLines(items []dmodel.DynamicFields) ([]itExt.FulfillmentLine, []string) {
	lines := make([]itExt.FulfillmentLine, 0, len(items))
	itemIds := make([]string, 0, len(items))
	for _, item := range items {
		outstanding := decimalOf(item, models.SalesOrderFulfillmentItemFieldOrderedQty).
			Sub(decimalOf(item, models.SalesOrderFulfillmentItemFieldFulfilledQty)).
			Sub(decimalOf(item, models.SalesOrderFulfillmentItemFieldRefundedQty))
		if !outstanding.IsPositive() {
			continue
		}
		lines = append(lines, itExt.FulfillmentLine{
			SalesOrderLineId: stringOf(item, models.SalesOrderFulfillmentItemFieldSalesOrderLineId),
			ProductVariantId: stringOf(item, models.SalesOrderFulfillmentItemFieldProductVariantId),
			UomId:            stringOf(item, models.SalesOrderFulfillmentItemFieldUomId),
			Quantity:         outstanding,
			SourceLocationId: stringOf(item, models.SalesOrderFulfillmentItemFieldSourceLocationId),
		})
		itemIds = append(itemIds, stringOf(item, models.SalesOrderFulfillmentItemFieldId))
	}
	return lines, itemIds
}
