package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Create Reverse Transfer: the correction for goods that physically come back.
//
// Unlike an adjustment or a scrap this is not auto-validated: it produces a draft the user
// confirms, reserves and validates, because when the document is raised the goods have not arrived.
//
// The original transfer is never edited, reopened or cancelled. History must read original move →
// reverse move, and mutating the original would erase the fact the return exists to record.

// ReturnRequest names how much of each move to send back; empty Lines means everything still
// returnable.
//
// It must stay an alias of the port's type, not a structurally identical copy: Go matches method
// signatures by identity, so a copy would make CreateReturn silently fail to satisfy
// StockTransferMovementService, surfacing as an assertion error at wiring time.
type ReturnRequest = itStock.TransferReturnRequest

// ReturnLineRequest is one move's requested return quantity.
type ReturnLineRequest = itStock.TransferReturnLine

// CreateReturn raises a draft reverse transfer for a done transfer.
func (this *StockTransferDomainServiceImpl) CreateReturn(
	ctx corectx.Context, transferId string, request ReturnRequest,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withTransferTransaction(ctx, func(tranxCtx corectx.Context) error {
		operation, err := loadTransferOperation(tranxCtx, transferId)
		if err != nil {
			return err
		}
		if operation == nil {
			result = notFoundResult(transferId)
			return nil
		}

		status := derefString(operation.Transfer.GetStatus())
		if status != models.StockTransferStatusDone {
			result = violationResult(
				"stock_transfer.not_done",
				"only a completed transfer can be returned; this one is '"+status+"'")
			return nil
		}

		outcome, err := buildReturnTransfer(tranxCtx, operation, request)
		if err != nil {
			return err
		}
		result = outcome
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// buildReturnTransfer computes the returnable quantities, validates the request and writes the
// new transfer with its moves.
func buildReturnTransfer(
	ctx corectx.Context, operation *transferOperationContext, request ReturnRequest,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	lines, err := collectReturnableLines(ctx, operation)
	if err != nil {
		return nil, err
	}
	if TotalReturnable(lines).IsZero() {
		return violationResult(
			"stock_return.nothing_returnable",
			"every line of this transfer has already been returned"), nil
	}

	requested, vErrs := resolveRequestedReturns(lines, request)
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	returnId, err := insertReturnTransfer(ctx, operation)
	if err != nil {
		return nil, err
	}
	if err := insertReturnMoves(ctx, operation, returnId, requested); err != nil {
		return nil, err
	}
	return mutateOk(), nil
}

// collectReturnableLines works out what each of the original's moves can still take back. Completed
// comes from the move's lines, counting only those validate marked picked — never from
// demand_quantity, which would let a customer return goods that never left.
func collectReturnableLines(
	ctx corectx.Context, operation *transferOperationContext,
) ([]ReturnableLine, error) {
	alreadyReturned, err := sumAlreadyReturned(ctx, operation)
	if err != nil {
		return nil, err
	}

	lines := make([]ReturnableLine, 0, len(operation.Moves))
	for _, item := range operation.Moves {
		move := models.NewStockMoveFrom(item)
		moveId := derefString(move.GetId())

		completed, err := sumPickedQuantity(ctx, operation, moveId)
		if err != nil {
			return nil, err
		}
		if completed.LessThanOrEqual(decimal.Zero) {
			continue
		}

		lines = append(lines, ReturnableLine{
			MoveId:           moveId,
			ProductVariantId: derefString(move.GetProductVariantId()),
			Completed:        completed,
			AlreadyReturned:  alreadyReturned[moveId],
		})
	}
	return lines, nil
}

// sumPickedQuantity totals what a move actually executed.
func sumPickedQuantity(
	ctx corectx.Context, operation *transferOperationContext, moveId string,
) (decimal.Decimal, error) {
	items, err := models.FindMoveLines(
		ctx, operation.MoveLineEngine.ResourceRepository(), moveId, models.MaxMoveLines)
	if err != nil {
		return decimal.Zero, err
	}

	total := decimal.Zero
	for _, item := range items {
		line := models.NewStockMoveLineFrom(item)
		if !derefBool(line.GetPicked()) {
			continue
		}
		total = total.Add(orZero(line.GetBaseQuantity()))
	}
	return total, nil
}

// sumAlreadyReturned totals what previous returns have taken back, keyed by original move. Only
// done returns count: a draft return in flight must not reduce the returnable, or two half-finished
// returns would block each other forever.
func sumAlreadyReturned(
	ctx corectx.Context, operation *transferOperationContext,
) (map[string]decimal.Decimal, error) {
	totals := map[string]decimal.Decimal{}

	returns, err := findReturnsOf(ctx, operation, derefString(operation.Transfer.GetId()))
	if err != nil {
		return nil, err
	}

	for _, item := range returns {
		returnTransfer := models.NewStockTransferFrom(item)
		if derefString(returnTransfer.GetStatus()) != models.StockTransferStatusDone {
			continue
		}

		moves, err := models.FindTransferMoves(
			ctx, operation.MoveEngine.ResourceRepository(),
			derefString(returnTransfer.GetId()), models.MaxTransferMoves)
		if err != nil {
			return nil, err
		}
		for _, moveItem := range moves {
			move := models.NewStockMoveFrom(moveItem)
			origin := derefString(move.GetOriginMoveId())
			if origin == "" {
				continue
			}
			shipped, err := sumPickedQuantity(ctx, operation, derefString(move.GetId()))
			if err != nil {
				return nil, err
			}
			totals[origin] = totals[origin].Add(shipped)
		}
	}
	return totals, nil
}

// findReturnsOf lists the transfers already raised as returns of this one.
func findReturnsOf(
	ctx corectx.Context, operation *transferOperationContext, transferId string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		models.StockTransferFieldReturnOfId, dmodel.Equals, transferId))

	found, err := operation.TransferEngine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  models.MaxTransferMoves,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findReturnsOf")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}

// resolveRequestedReturns turns the request into a per-move quantity, defaulting to everything.
// Requesting more than the returnable is refused with no override, deliberately: AssertReturnable
// and ReturnableLine.Returnable() take no waiver parameter either, so the cap stays somewhere a
// caller cannot reach past. Any future override must be an authorisation decision, not a request
// flag.
func resolveRequestedReturns(
	lines []ReturnableLine, request ReturnRequest,
) ([]ReturnLineRequest, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	byMoveId := map[string]ReturnableLine{}
	for _, line := range lines {
		byMoveId[line.MoveId] = line
	}

	if len(request.Lines) == 0 {
		return defaultFullReturn(lines), vErrs
	}

	resolved := make([]ReturnLineRequest, 0, len(request.Lines))
	for _, requested := range request.Lines {
		line, ok := byMoveId[requested.MoveId]
		if !ok {
			vErrs.Append(*ft.NewBusinessViolation(
				models.StockMoveSchemaName, "stock_return.unknown_move",
				"move '"+requested.MoveId+"' is not part of the transfer being returned"))
			continue
		}
		if lineErrs := AssertReturnable(line, requested.Quantity); lineErrs.Count() > 0 {
			vErrs.ConcatPtr(lineErrs)
			continue
		}
		resolved = append(resolved, requested)
	}
	return resolved, vErrs
}

// defaultFullReturn returns everything still returnable. The return lands as a draft and the user
// adjusts its move quantities through ordinary CRUD before confirming.
func defaultFullReturn(lines []ReturnableLine) []ReturnLineRequest {
	resolved := make([]ReturnLineRequest, 0, len(lines))
	for _, line := range lines {
		remaining := line.Returnable()
		if remaining.LessThanOrEqual(decimal.Zero) {
			continue
		}
		resolved = append(resolved, ReturnLineRequest{MoveId: line.MoveId, Quantity: remaining})
	}
	return resolved
}

// insertReturnTransfer writes the reverse transfer's header. The operation code is inverted, since
// a return of an outgoing delivery is an incoming receipt; that inversion is what makes the reverse
// transfer take ensureIncomingLine's path at validate.
func insertReturnTransfer(
	ctx corectx.Context, operation *transferOperationContext,
) (string, error) {
	original := operation.Transfer
	reversedCode := reverseOperationCode(derefString(original.GetOperationCode()))

	operationType, err := findOperationTypeByCode(ctx, derefString(original.GetOrgId()), reversedCode)
	if err != nil {
		return "", err
	}

	transferNumber, err := generateTransferNumber("RET")
	if err != nil {
		return "", err
	}

	fields := dmodel.DynamicFields{
		models.StockTransferFieldTransferNumber: transferNumber,
		models.StockTransferFieldOperationCode:  reversedCode,
		// Source and destination swap: the goods retrace their steps.
		models.StockTransferFieldSourceLocationId:      derefString(original.GetDestinationLocationId()),
		models.StockTransferFieldDestinationLocationId: derefString(original.GetSourceLocationId()),
		models.StockTransferFieldStatus:                models.StockTransferStatusDraft,
		models.StockTransferFieldReturnOfId:            derefString(original.GetId()),
		models.StockTransferFieldReservationMethod:     derefString(original.GetReservationMethod()),
		models.StockTransferFieldBackorderPolicy:       models.StockBackorderPolicyNever,
		models.StockTransferFieldShippingPolicy:        derefString(original.GetShippingPolicy()),
		models.StockTransferFieldOriginReference:       derefString(original.GetTransferNumber()),
		models.StockTransferFieldOrgId:                 derefString(original.GetOrgId()),
	}
	// An operation type of the opposite direction is preferred, but its absence must not block a
	// return: the reversed code on the transfer is what drives the movement.
	if operationType != nil {
		fields[models.StockTransferFieldOperationTypeId] = derefString(operationType.GetId())
	} else {
		fields[models.StockTransferFieldOperationTypeId] = derefString(original.GetOperationTypeId())
	}

	if _, err := operation.TransferEngine.ResourceRepository().Insert(ctx, fields); err != nil {
		return "", errors.Wrap(err, "insertReturnTransfer")
	}
	return findTransferByNumber(ctx, operation, derefString(original.GetOrgId()), transferNumber)
}

// insertReturnMoves gives the reverse transfer one move per returned line.
func insertReturnMoves(
	ctx corectx.Context,
	operation *transferOperationContext,
	returnId string,
	requested []ReturnLineRequest,
) error {
	byId := indexMovesById(operation.Moves)

	for _, line := range requested {
		source, ok := byId[line.MoveId]
		if !ok {
			continue
		}
		quantity := line.Quantity.String()

		_, err := operation.MoveEngine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
			models.StockMoveFieldTransferId:         returnId,
			models.StockMoveFieldProductVariantId:   derefString(source.GetProductVariantId()),
			models.StockMoveFieldDemandQuantity:     quantity,
			models.StockMoveFieldBaseDemandQuantity: quantity,
			// Reversed, matching the transfer header.
			models.StockMoveFieldSourceLocationId:      derefString(source.GetDestinationLocationId()),
			models.StockMoveFieldDestinationLocationId: derefString(source.GetSourceLocationId()),
			models.StockMoveFieldStatus:                models.StockMoveStatusDraft,
			// Points at the move being reversed, so history reads original move → reverse move and the
			// next return can tell what has already come back.
			models.StockMoveFieldOriginMoveId: line.MoveId,
			models.StockMoveFieldOrgId:        derefString(source.GetOrgId()),
		})
		if err != nil {
			return errors.Wrap(err, "insertReturnMoves")
		}
	}
	return nil
}

// reverseOperationCode flips the direction of a movement. An internal transfer reverses to another
// internal one: both ends are the company's own locations, so there is no direction to invert.
func reverseOperationCode(code string) string {
	switch code {
	case models.StockOperationCodeOutgoing:
		return models.StockOperationCodeIncoming
	case models.StockOperationCodeIncoming:
		return models.StockOperationCodeOutgoing
	default:
		return models.StockOperationCodeInternal
	}
}

// findOperationTypeByCode resolves an org's operation type for a given direction.
func findOperationTypeByCode(
	ctx corectx.Context, orgId, operationCode string,
) (*models.StockOperationType, error) {
	engine, err := engineFor(models.StockOperationTypeSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockOperationTypeFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(
			models.StockOperationTypeFieldOperationCode, dmodel.Equals, operationCode),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findOperationTypeByCode")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewStockOperationTypeFrom(found.Data.Items[0]), nil
}
