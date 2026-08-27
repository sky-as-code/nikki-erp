package services

import (
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Fulfilment (BR 44, 45, 87.8, SALES-029).
//
// **Sales sends intent and never touches stock** (BR 3.2). This file writes what was asked and
// records what Inventory answered; it computes no availability, chooses no warehouse and moves
// nothing. BR 94.6 tests exactly that.
//
// # One order has many requests (BR 45)
//
// Partial delivery, several warehouses and split shipment each produce another request. That is why
// this is a table rather than a `fulfillment_reference` column on the order - a single column would
// have to be rewritten on the second shipment, losing the first.
//
// # Accepted is not completed
//
// A request Inventory accepted has stock HELD; one it completed has goods MOVED. Only completed
// requests feed `fulfilled_quantity`, because BR 7.3's failure - money captured, goods not dispensed
// - lives exactly between the two, and counting an acceptance would tell a customer their goods had
// shipped when they had not.
//
// # The port is declared but not bound
//
// Inventory publishes no mutation interface (see interfaces/external/fulfillment.go). A request is
// therefore written and left `pending` when no port is available. That is the honest state: the sale
// really has asked for goods, and Inventory really has not answered.

// RaiseFulfillmentResult is what raising a request concluded.
type RaiseFulfillmentResult struct {
	RequestId string
	Status    string

	// Dispatched says whether the request actually reached Inventory. False means it was recorded
	// and is waiting for the port to exist - not that anything failed.
	Dispatched bool

	InventoryReference string
}

// The refusal reasons fulfilment can produce.
const (
	ReasonNothingToFulfill  = "sales_fulfillment.nothing_to_fulfill"
	ReasonOrderNotFulfilled = "sales_fulfillment.order_not_confirmed"
)

// RaiseFulfillmentRequest records an intent and, when the port is bound, sends it.
//
// Only lines that still need fulfilling are included: asking Inventory again for goods it has
// already issued would move them twice.
func RaiseFulfillmentRequest(
	ctx corectx.Context,
	orderId string,
	requestType string,
	fulfillment itExt.FulfillmentExtService,
) (*RaiseFulfillmentResult, *ft.ClientErrors, error) {
	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, nil, err
	}
	if order == nil {
		return nil, OrderNotFoundErrors(orderId), nil
	}

	// A draft has not been sold yet, so there is nothing to ask for. Confirmation is what commits
	// the business, and it is confirmation that raises the first request.
	if stringOf(order, models.SalesOrderFieldStatus) == string(models.SalesOrderStatusDraft) {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("status", ReasonOrderNotFulfilled,
			"a draft order has not been sold and cannot be fulfilled"))
		return nil, vErrs, nil
	}

	outstanding, err := outstandingLines(ctx, orderId)
	if err != nil {
		return nil, nil, err
	}
	if len(outstanding) == 0 {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("lines", ReasonNothingToFulfill,
			"every line of this order has already been fulfilled"))
		return nil, vErrs, nil
	}

	requestId, err := writeFulfillmentRequest(ctx, order, requestType, outstanding)
	if err != nil {
		return nil, nil, err
	}

	// No port bound: the request stands pending. Deliberately not an error - the sale is valid and
	// the goods are genuinely owed, and refusing here would make confirmation impossible.
	if fulfillment == nil {
		return &RaiseFulfillmentResult{
			RequestId: requestId,
			Status:    string(models.SalesFulfillmentStatusPending),
		}, nil, nil
	}

	response, err := dispatchFulfillment(ctx, fulfillment, requestType, orderId, requestId, outstanding)
	if err != nil {
		return nil, nil, err
	}
	if err := recordFulfillmentOutcome(ctx, requestId, response); err != nil {
		return nil, nil, err
	}

	status := string(models.SalesFulfillmentStatusRejected)
	switch {
	case response.Completed:
		status = string(models.SalesFulfillmentStatusCompleted)
	case response.Accepted:
		status = string(models.SalesFulfillmentStatusAccepted)
	}

	return &RaiseFulfillmentResult{
		RequestId:          requestId,
		Status:             status,
		Dispatched:         true,
		InventoryReference: response.InventoryReference,
	}, nil, nil
}

// outstandingLines finds what an order still owes, as fulfilment lines.
//
// A line whose fulfilled quantity already covers what was ordered is skipped. A line needing no
// fulfilment at all - a service, a digital item - is skipped too, and an order made entirely of
// those produces no request, which is what makes `not_required` reachable.
func outstandingLines(
	ctx corectx.Context, orderId string,
) ([]itExt.FulfillmentLine, error) {
	lineRecords, err := searchBy(ctx,
		models.SalesOrderLineSchemaName, models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	outstanding := make([]itExt.FulfillmentLine, 0, len(lineRecords))
	for _, record := range lineRecords {
		ordered := decimalOf(record, models.SalesOrderLineFieldOrderedQuantity)
		fulfilled := decimalOf(record, models.SalesOrderLineFieldFulfilledQuantity)

		remaining := ordered.Sub(fulfilled)
		if !remaining.IsPositive() {
			continue
		}
		if !boolOrTrue(record, models.SalesOrderLineFieldRequiresFulfillment) {
			// A service or a fee owes no goods, so there is nothing to ask Inventory for. An order
			// made entirely of these produces no request at all, which is what makes
			// `not_required` reachable rather than leaving the order pending forever.
			continue
		}

		outstanding = append(outstanding, itExt.FulfillmentLine{
			SalesOrderLineId: stringOf(record, models.SalesOrderLineFieldId),
			ProductVariantId: stringOf(record, models.SalesOrderLineFieldProductVariantId),
			UomId:            stringOf(record, models.SalesOrderLineFieldUomId),
			Quantity:         remaining,
		})
	}
	return outstanding, nil
}

// dispatchFulfillment sends the intent to whichever port method matches the request type.
func dispatchFulfillment(
	ctx corectx.Context,
	fulfillment itExt.FulfillmentExtService,
	requestType, orderId, requestId string,
	lines []itExt.FulfillmentLine,
) (*itExt.FulfillmentResponse, error) {
	request := itExt.FulfillmentRequest{
		SalesOrderId:              orderId,
		SalesFulfillmentRequestId: requestId,

		// The request id IS the idempotency key. It is unique, it is already stored, and it means a
		// retry of the same request cannot issue the goods twice.
		IdempotencyKey: requestId,
		Lines:          lines,
	}

	switch requestType {
	case string(models.SalesFulfillmentTypeReservation):
		return fulfillment.RequestReservation(ctx, request)
	case string(models.SalesFulfillmentTypeGoodsIssue):
		return fulfillment.RequestGoodsIssue(ctx, request)
	case string(models.SalesFulfillmentTypeReturnReceipt):
		return fulfillment.RequestReturnReceipt(ctx, request)
	}
	// An unrecognised type is not dispatched. Guessing which movement was meant is the one mistake
	// that moves the wrong goods.
	return &itExt.FulfillmentResponse{
		FailureReason: "unrecognised fulfilment request type '" + requestType + "'",
	}, nil
}

// writeFulfillmentRequest stores the intent and its lines, in one transaction.
func writeFulfillmentRequest(
	ctx corectx.Context,
	order dmodel.DynamicFields,
	requestType string,
	lines []itExt.FulfillmentLine,
) (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", err
	}
	requestId := string(*id)
	orgId := stringOf(order, basemodel.FieldOrgId)

	err = withTransaction(ctx, models.SalesFulfillmentRequestSchemaName,
		func(tranxCtx corectx.Context) error {
			engine, err := engineFor(models.SalesFulfillmentRequestSchemaName)
			if err != nil {
				return err
			}
			if _, err := engine.ResourceRepository().Insert(tranxCtx, dmodel.DynamicFields{
				models.SalesFulfillmentRequestFieldId:           requestId,
				models.SalesFulfillmentRequestFieldSalesOrderId: stringOf(order, models.SalesOrderFieldId),
				models.SalesFulfillmentRequestFieldRequestType:  requestType,
				models.SalesFulfillmentRequestFieldStatus:       string(models.SalesFulfillmentStatusPending),
				models.SalesFulfillmentRequestFieldRequestedAt:  model.ModelDateTime(time.Now().UTC()),
				basemodel.FieldOrgId:                            orgId,
			}); err != nil {
				return err
			}

			lineEngine, err := engineFor(models.SalesFulfillmentRequestLineSchemaName)
			if err != nil {
				return err
			}
			for _, line := range lines {
				lineId, err := model.NewId()
				if err != nil {
					return err
				}
				if _, err := lineEngine.ResourceRepository().Insert(tranxCtx, dmodel.DynamicFields{
					models.SalesFulfillmentLineFieldId:          string(*lineId),
					models.SalesFulfillmentLineFieldRequestId:   requestId,
					models.SalesFulfillmentLineFieldOrderLineId: line.SalesOrderLineId,
					models.SalesFulfillmentLineFieldQuantity:    line.Quantity,
					basemodel.FieldOrgId:                        orgId,
				}); err != nil {
					return err
				}
			}

			// Announced inside the same transaction as the request itself, so a consumer can never
			// be told goods were asked for by a request that then rolled back.
			_, err = RecordEvent(tranxCtx, RecordEventParams{
				EventType:   models.EventSalesFulfillmentRequested,
				AggregateId: stringOf(order, models.SalesOrderFieldId),
				OrgId:       orgId,
				Payload: map[string]any{
					"sales_fulfillment_request_id": requestId,
					"sales_order_id":               stringOf(order, models.SalesOrderFieldId),
					"request_type":                 requestType,
					"line_count":                   len(lines),
				},
			})
			return err
		})
	if err != nil {
		return "", err
	}
	return requestId, nil
}

// recordFulfillmentOutcome stamps what Inventory answered onto the request.
func recordFulfillmentOutcome(
	ctx corectx.Context, requestId string, response *itExt.FulfillmentResponse,
) error {
	engine, err := engineFor(models.SalesFulfillmentRequestSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{
		models.SalesFulfillmentRequestFieldId: requestId,
	}
	switch {
	case response.Completed:
		update[models.SalesFulfillmentRequestFieldStatus] =
			string(models.SalesFulfillmentStatusCompleted)
		update[models.SalesFulfillmentRequestFieldCompletedAt] =
			model.ModelDateTime(time.Now().UTC())
	case response.Accepted:
		update[models.SalesFulfillmentRequestFieldStatus] =
			string(models.SalesFulfillmentStatusAccepted)
	default:
		update[models.SalesFulfillmentRequestFieldStatus] =
			string(models.SalesFulfillmentStatusRejected)
	}
	if response.InventoryReference != "" {
		update[models.SalesFulfillmentRequestFieldInventoryRef] = response.InventoryReference
	}
	if response.FailureReason != "" {
		update[models.SalesFulfillmentRequestFieldFailReason] = response.FailureReason
	}

	_, err = engine.ResourceRepository().Update(ctx, update)
	return err
}

// SyncFulfilledQuantities recomputes each line's fulfilled quantity from the COMPLETED requests.
//
// Derived rather than incremented. An increment assumes every previous update landed exactly once,
// and compounds the error if one did not; a recount is self-correcting, and it is what makes the
// stored quantity honestly a summary of the requests rather than a second source of truth that can
// drift from them.
func SyncFulfilledQuantities(ctx corectx.Context, orderId string) error {
	requests, err := searchBy(ctx,
		models.SalesFulfillmentRequestSchemaName,
		models.SalesFulfillmentRequestFieldSalesOrderId, orderId)
	if err != nil {
		return err
	}

	fulfilled := map[string]decimal.Decimal{}
	for _, request := range requests {
		if !models.NewSalesFulfillmentRequestFrom(request).IsCompleted() {
			// Accepted is not completed: stock is held, nothing has moved. Counting it would report
			// goods as shipped that are still on the shelf.
			continue
		}

		lines, err := searchBy(ctx,
			models.SalesFulfillmentRequestLineSchemaName,
			models.SalesFulfillmentLineFieldRequestId,
			stringOf(request, models.SalesFulfillmentRequestFieldId))
		if err != nil {
			return err
		}
		for _, line := range lines {
			lineId := stringOf(line, models.SalesFulfillmentLineFieldOrderLineId)
			fulfilled[lineId] = fulfilled[lineId].Add(
				decimalOf(line, models.SalesFulfillmentLineFieldQuantity))
		}
	}

	engine, err := engineFor(models.SalesOrderLineSchemaName)
	if err != nil {
		return err
	}
	for lineId, quantity := range fulfilled {
		if _, err := engine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
			models.SalesOrderLineFieldId:                lineId,
			models.SalesOrderLineFieldFulfilledQuantity: quantity,
		}); err != nil {
			return err
		}
	}
	return nil
}

// DeriveOrderFulfillmentStatus answers an order's fulfilment status from its lines.
//
// Reuses SALES-015's pure derivation rather than restating the rule, so a line and an order cannot
// come to different conclusions about what "partially fulfilled" means.
func DeriveOrderFulfillmentStatus(
	ctx corectx.Context, orderId string,
) (string, error) {
	lineRecords, err := searchBy(ctx,
		models.SalesOrderLineSchemaName, models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return "", err
	}

	quantities := make([]LineQuantities, 0, len(lineRecords))
	for _, record := range lineRecords {
		quantities = append(quantities, LineQuantities{
			Ordered:   decimalOf(record, models.SalesOrderLineFieldOrderedQuantity),
			Fulfilled: decimalOf(record, models.SalesOrderLineFieldFulfilledQuantity),
			Returned:  decimalOf(record, models.SalesOrderLineFieldReturnedQuantity),

			// Absent reads as TRUE, matching the column default. A missing value must not silently
			// make a line need no goods, which would report an unshipped sale as complete.
			RequiresFulfillment: boolOrTrue(record, models.SalesOrderLineFieldRequiresFulfillment),
		})
	}
	return DeriveFulfillmentStatus(quantities), nil
}

// boolOrTrue reads a boolean field, treating an absent one as TRUE.
//
// The opposite of the usual zero-value reading, and deliberately. This backs requires_fulfillment,
// where the two mistakes are not symmetric: a line wrongly needing goods holds an order open until
// somebody notices, while a line wrongly needing none reports a sale as shipped that never was.
func boolOrTrue(record dmodel.DynamicFields, field string) bool {
	value, present := record[field]
	if !present || value == nil {
		return true
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case *bool:
		if typed != nil {
			return *typed
		}
	}
	return true
}
