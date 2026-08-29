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

// Fulfilment. Sales sends intent and never touches stock: this file writes what was asked and
// records what Inventory answered, computing no availability and moving nothing.
//
// One order has many requests - partial delivery, several warehouses, split shipment - hence a
// table rather than a `fulfillment_reference` column the second shipment would overwrite.
//
// Accepted is not completed: an accepted request has stock HELD, a completed one has goods MOVED.
// Only completed requests feed `fulfilled_quantity`, because counting an acceptance would tell a
// customer their goods had shipped when they had not. The port is declared but not bound, so a
// request is written and left `pending` when none is available.

type RaiseFulfillmentResult struct {
	RequestId string
	Status    string

	// Dispatched false means the request was recorded and awaits the port, not that anything failed.
	Dispatched bool

	InventoryReference string
}

// The refusal reasons fulfilment can produce.
const (
	ReasonNothingToFulfill  = "sales_fulfillment.nothing_to_fulfill"
	ReasonOrderNotFulfilled = "sales_fulfillment.order_not_confirmed"
)

// RaiseFulfillmentRequest includes only lines that still need fulfilling: asking again for goods
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

	// A draft has not been sold yet. Confirmation raises the first request.
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

	// No port bound: the request stands pending. Deliberately not an error - the goods are genuinely
	// owed, and refusing here would make confirmation impossible.
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

// outstandingLines skips a line already fulfilled and one needing no fulfilment at all, so an
// order made entirely of those produces no request - which is what makes `not_required` reachable.
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
			// A service or a fee owes no goods. An order made entirely of these produces no request,
			// which makes `not_required` reachable rather than pending forever.
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

func dispatchFulfillment(
	ctx corectx.Context,
	fulfillment itExt.FulfillmentExtService,
	requestType, orderId, requestId string,
	lines []itExt.FulfillmentLine,
) (*itExt.FulfillmentResponse, error) {
	request := itExt.FulfillmentRequest{
		SalesOrderId:              orderId,
		SalesFulfillmentRequestId: requestId,

		// The request id IS the idempotency key: unique, already stored, so a retry of the same
		// request cannot issue the goods twice.
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

			// Announced inside the same transaction as the request, so a consumer can never be told
			// goods were asked for by a request that then rolled back.
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
// Derived rather than incremented: an increment compounds any update that did not land, while a
// recount is self-correcting.
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

// DeriveOrderFulfillmentStatus reuses the pure derivation rather than restating the rule, so a
// line and an order cannot disagree about what "partially fulfilled" means.
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

			// Absent reads as TRUE, matching the column default. A missing value must not silently make
			// a line need no goods, which would report an unshipped sale as complete.
			RequiresFulfillment: boolOrTrue(record, models.SalesOrderLineFieldRequiresFulfillment),
		})
	}
	return DeriveFulfillmentStatus(quantities), nil
}

// boolOrTrue treats an absent boolean as TRUE - the opposite of the usual zero-value reading, and
// deliberate. It backs requires_fulfillment, where a line wrongly needing goods holds an order
// open until somebody notices, while one wrongly needing none reports a sale as shipped.
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
