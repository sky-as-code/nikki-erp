package external

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	invModels "github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The fulfilment adapter translates Sales' single commercial intent into inventory's document
// lifecycle (create, confirm, reserve, validate), so Sales never has to know that sequence.
//
// No physical decision is made here: the adapter names an operation type and nothing else, and
// inventory's Create stamps source and destination from that type's defaults — which is why the
// port carries no warehouse or location field.
//
// Inventory reports a business refusal (no stock, an archived operation type) as ClientErrors with
// a nil Go error; those become Accepted: false with the reason carried across. Only a genuine Go
// error propagates as one, and Sales then leaves the request pending to retry.

type fulfillmentAdapter struct {
	transfers itStock.StockTransferMovementService

	// operationTypes resolves which operation type to raise a transfer against. The ids live in
	// settings because they are part of the deployment's warehouse setup, not a fact about selling.
	operationTypes operationTypeResolver
}

// operationTypeResolver answers which inventory operation type a given intent uses. The settings-
// backed implementation never returns an error (an unreadable setting falls back to its default);
// the error is declared so a resolver that CAN fail keeps "not configured" (empty string) distinct
// from "could not find out", which must not be reported as a configuration gap.
type operationTypeResolver interface {
	OutgoingOperationTypeId(ctx corectx.Context) (string, error)
	IncomingOperationTypeId(ctx corectx.Context) (string, error)
}

// RequestReservation holds stock for a confirmed sale without moving it. It stops after Reserve:
// the goods stay where they are and stay claimed, so Completed is false.
func (this *fulfillmentAdapter) RequestReservation(
	ctx corectx.Context, request itExt.FulfillmentRequest,
) (*itExt.FulfillmentResponse, error) {
	transferId, response, err := this.raiseOutgoingTransfer(ctx, request)
	if err != nil || response != nil {
		return response, err
	}

	if refused, err := this.confirmAndReserve(ctx, transferId); err != nil || refused != nil {
		return withReference(refused, transferId), err
	}

	return &itExt.FulfillmentResponse{
		Accepted:           true,
		InventoryReference: transferId,
		Completed:          false,
	}, nil
}

// RequestGoodsIssue moves the goods out: the same sequence as a reservation, then Validate, which
// is the irreversible step. The two share a path so they cannot drift.
func (this *fulfillmentAdapter) RequestGoodsIssue(
	ctx corectx.Context, request itExt.FulfillmentRequest,
) (*itExt.FulfillmentResponse, error) {
	transferId, response, err := this.raiseOutgoingTransfer(ctx, request)
	if err != nil || response != nil {
		return response, err
	}

	if refused, err := this.confirmAndReserve(ctx, transferId); err != nil || refused != nil {
		return withReference(refused, transferId), err
	}

	// The idempotency key is Sales' own request id, so a retry of a call that timed out after
	// inventory committed cannot move the goods a second time. createBackorder is nil so the
	// shortfall follows the operation type's own backorder policy, which is a warehouse decision.
	result, err := this.transfers.Validate(ctx, transferId, request.IdempotencyKey, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "validating stock transfer '%s'", transferId)
	}
	if refused := refusalOf(result); refused != nil {
		return withReference(refused, transferId), nil
	}

	return &itExt.FulfillmentResponse{
		Accepted:           true,
		InventoryReference: transferId,
		Completed:          true,
	}, nil
}

// RequestReturnReceipt takes returned goods back in as an incoming transfer of its own rather than
// via inventory's CreateReturn, which reverses a specific done transfer by naming its moves — ids
// Sales does not hold. A return arriving without the original transfer must still be receivable.
func (this *fulfillmentAdapter) RequestReturnReceipt(
	ctx corectx.Context, request itExt.FulfillmentRequest,
) (*itExt.FulfillmentResponse, error) {
	operationTypeId, err := this.operationTypes.IncomingOperationTypeId(ctx)
	if err != nil {
		return nil, err
	}
	if operationTypeId == "" {
		return notConfigured("incoming"), nil
	}

	transferId, response, err := this.raiseTransfer(ctx, request, operationTypeId)
	if err != nil || response != nil {
		return response, err
	}

	// Confirmed but not reserved: an incoming transfer draws from outside, so there is no balance
	// to claim.
	if refused, err := this.confirmOnly(ctx, transferId); err != nil || refused != nil {
		return withReference(refused, transferId), err
	}

	result, err := this.transfers.Validate(ctx, transferId, request.IdempotencyKey, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "validating return transfer '%s'", transferId)
	}
	if refused := refusalOf(result); refused != nil {
		return withReference(refused, transferId), nil
	}

	return &itExt.FulfillmentResponse{
		Accepted:           true,
		InventoryReference: transferId,
		Completed:          true,
	}, nil
}

// ReleaseReservation gives back a hold a cancelled sale no longer needs. It unreserves without
// cancelling the transfer, so the transfer's history records that it existed and was released.
func (this *fulfillmentAdapter) ReleaseReservation(
	ctx corectx.Context, inventoryReference string,
) (*itExt.FulfillmentResponse, error) {
	if inventoryReference == "" {
		// Nothing was ever reserved — a sale cancelled before it reached inventory. Accepted,
		// because the caller's intent (hold nothing) is already true.
		return &itExt.FulfillmentResponse{Accepted: true}, nil
	}

	result, err := this.transfers.Unreserve(ctx, inventoryReference)
	if err != nil {
		return nil, errors.Wrapf(err, "unreserving stock transfer '%s'", inventoryReference)
	}
	if refused := refusalOf(result); refused != nil {
		return withReference(refused, inventoryReference), nil
	}

	return &itExt.FulfillmentResponse{
		Accepted:           true,
		InventoryReference: inventoryReference,
	}, nil
}

// raiseOutgoingTransfer creates the draft transfer for a sale leaving the business.
func (this *fulfillmentAdapter) raiseOutgoingTransfer(
	ctx corectx.Context, request itExt.FulfillmentRequest,
) (string, *itExt.FulfillmentResponse, error) {
	operationTypeId, err := this.operationTypes.OutgoingOperationTypeId(ctx)
	if err != nil {
		return "", nil, err
	}
	if operationTypeId == "" {
		return "", notConfigured("outgoing"), nil
	}
	return this.raiseTransfer(ctx, request, operationTypeId)
}

// raiseTransfer creates the draft transfer and its moves. It returns a non-nil response only when
// inventory refused; a successful create returns the id with a nil response.
func (this *fulfillmentAdapter) raiseTransfer(
	ctx corectx.Context, request itExt.FulfillmentRequest, operationTypeId string,
) (string, *itExt.FulfillmentResponse, error) {
	// One call, so the header and its moves commit together: an empty transfer — header written,
	// moves not — validates successfully and reports goods moved that were never named.
	created, err := this.transfers.CreateWithMoves(ctx, dmodel.DynamicFields{
		invModels.StockTransferFieldOperationTypeId: operationTypeId,

		// The sales order is the origin, a plain string because inventory holds no reference to
		// Sales and must not resolve one.
		invModels.StockTransferFieldOriginReference: request.SalesOrderId,

		// Unique per request; inventory recognises a retry by it, so the same intent sent twice
		// raises one transfer rather than two.
		invModels.StockTransferFieldIdempotencyKey: request.IdempotencyKey,
	}, movesOf(request.Lines))
	if err != nil {
		return "", nil, errors.Wrap(err, "creating stock transfer for fulfilment")
	}
	if created.ClientErrors.Count() > 0 {
		return "", refusalFrom(&created.ClientErrors), nil
	}
	if !created.HasData {
		return "", nil, errors.New("inventory created no stock transfer and reported no violation")
	}

	transferId := derefString(created.Data.GetModelId(invModels.StockTransferFieldId))
	if transferId == "" {
		return "", nil, errors.New("the created stock transfer carries no id")
	}
	return transferId, nil, nil
}

// movesOf translates sold lines into the movement request inventory takes. SalesOrderLineId is
// dropped: inventory_stock_move has no per-line origin column, and the transfer's origin_reference
// ties the document back to the order at the granularity inventory records.
func movesOf(lines []itExt.FulfillmentLine) []itStock.TransferMoveRequest {
	moves := make([]itStock.TransferMoveRequest, 0, len(lines))
	for _, line := range lines {
		moves = append(moves, itStock.TransferMoveRequest{
			ProductVariantId: line.ProductVariantId,
			UomId:            line.UomId,
			Quantity:         line.Quantity,
		})
	}
	return moves
}

// confirmAndReserve moves the draft into the flow and claims stock for it.
func (this *fulfillmentAdapter) confirmAndReserve(
	ctx corectx.Context, transferId string,
) (*itExt.FulfillmentResponse, error) {
	if refused, err := this.confirmOnly(ctx, transferId); err != nil || refused != nil {
		return refused, err
	}

	result, err := this.transfers.Reserve(ctx, transferId)
	if err != nil {
		return nil, errors.Wrapf(err, "reserving stock transfer '%s'", transferId)
	}
	return refusalOf(result), nil
}

func (this *fulfillmentAdapter) confirmOnly(
	ctx corectx.Context, transferId string,
) (*itExt.FulfillmentResponse, error) {
	result, err := this.transfers.Confirm(ctx, transferId)
	if err != nil {
		return nil, errors.Wrapf(err, "confirming stock transfer '%s'", transferId)
	}
	return refusalOf(result), nil
}

// refusalOf turns Inventory's client errors into a refusal, or nil when it accepted.
func refusalOf(result *dyn.OpResult[dyn.MutateResultData]) *itExt.FulfillmentResponse {
	if result == nil || result.ClientErrors.Count() == 0 {
		return nil
	}
	return refusalFrom(&result.ClientErrors)
}

// refusalFrom carries inventory's own words across: an operator told only "rejected" cannot tell
// whether to wait for stock, re-route the order or refund it.
func refusalFrom(vErrs *ft.ClientErrors) *itExt.FulfillmentResponse {
	return &itExt.FulfillmentResponse{
		Accepted:      false,
		FailureReason: describeViolations(vErrs),
	}
}

// describeViolations renders inventory's refusals into one line. Both key and message are kept: the
// key is stable and translatable for a UI, the message is what a person reading a log understands.
func describeViolations(vErrs *ft.ClientErrors) string {
	if vErrs == nil || len(*vErrs) == 0 {
		return "inventory refused the request without giving a reason"
	}

	parts := make([]string, 0, len(*vErrs))
	for _, item := range *vErrs {
		switch {
		case item.Message != "" && item.Key != "":
			parts = append(parts, item.Key+": "+item.Message)
		case item.Message != "":
			parts = append(parts, item.Message)
		case item.Key != "":
			parts = append(parts, item.Key)
		}
	}
	if len(parts) == 0 {
		return "inventory refused the request without giving a reason"
	}
	return strings.Join(parts, "; ")
}

// derefString reads an optional id as a plain string.
func derefString[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

// notConfigured answers when the deployment has not said which operation type to use. It is a
// refusal rather than a Go error, because a 500 would send an operator hunting a software fault
// instead of closing a configuration gap.
func notConfigured(direction string) *itExt.FulfillmentResponse {
	return &itExt.FulfillmentResponse{
		Accepted: false,
		FailureReason: "no " + direction + " stock operation type is configured for sales " +
			"fulfilment; set it in the Sales settings before fulfilling an order",
	}
}

// withReference attaches the transfer id to a refusal: a refused fulfilment still leaves a real
// transfer behind, and without the id an operator must hunt for it by origin reference.
func withReference(
	response *itExt.FulfillmentResponse, transferId string,
) *itExt.FulfillmentResponse {
	if response != nil && response.InventoryReference == "" {
		response.InventoryReference = transferId
	}
	return response
}

// settingsOperationTypes resolves the operation types from Sales' own org settings. It is a
// separate type because this lookup fails differently: a genuinely absent value is a configuration
// gap, not a stock problem.
type settingsOperationTypes struct {
	settings itExt.EffectiveSettingsExtService
}

func (this *settingsOperationTypes) OutgoingOperationTypeId(
	ctx corectx.Context,
) (string, error) {
	return services.ResolveSalesPolicy(ctx, this.settings).OutgoingOperationTypeId, nil
}

func (this *settingsOperationTypes) IncomingOperationTypeId(
	ctx corectx.Context,
) (string, error) {
	return services.ResolveSalesPolicy(ctx, this.settings).IncomingOperationTypeId, nil
}
