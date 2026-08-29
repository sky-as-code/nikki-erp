package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Validate is the only operation that changes an on-hand quantity, and it runs as one transaction:
// a source decremented without its destination incremented is stock that has ceased to exist.

// Validate executes a transfer: it consumes the reservations, moves the stock and closes the
// transfer, all in one transaction. Every quantity is re-read inside the source lock, because a
// figure fetched before the lock is stale by definition.
func (this *StockTransferDomainServiceImpl) Validate(
	ctx corectx.Context, transferId string, idempotencyKey string, createBackorder *bool,
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

		if replayed, err := replayedValidate(operation.Transfer, idempotencyKey); err != nil {
			return err
		} else if replayed != nil {
			result = replayed
			return nil
		}

		status := derefString(operation.Transfer.GetStatus())
		if !IsTransferOpen(status) {
			result = violationResult(
				"stock_transfer.already_closed",
				"a '"+status+"' transfer cannot be validated again")
			return nil
		}
		if len(operation.Moves) == 0 {
			result = violationResult(
				"stock_transfer.no_moves",
				"a transfer with no moves has nothing to validate")
			return nil
		}

		outcome, err := executeMoves(tranxCtx, operation)
		if err != nil {
			return err
		}
		result, err = finishValidate(tranxCtx, operation, outcome, idempotencyKey, createBackorder)
		return err
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// replayedValidate detects a retry of a validate that already succeeded, so a timed-out client
// retrying does not ship the goods twice. It matches only a transfer already `done`: the same key
// on an in-flight transfer means the earlier attempt did not complete, so the retry must proceed.
func replayedValidate(
	transfer models.StockTransfer, idempotencyKey string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if idempotencyKey == "" {
		return nil, nil
	}
	if derefString(transfer.GetIdempotencyKey()) != idempotencyKey {
		return nil, nil
	}
	if derefString(transfer.GetStatus()) != models.StockTransferStatusDone {
		return nil, nil
	}
	return mutateOk(), nil
}

// moveOutcome records what one move actually managed to process.
type moveOutcome struct {
	MoveId    model.Id
	Demand    decimal.Decimal
	Processed decimal.Decimal
}

// Shortfall is the part of the demand this move did not deliver.
func (this moveOutcome) Shortfall() decimal.Decimal {
	short := this.Demand.Sub(this.Processed)
	if short.LessThan(decimal.Zero) {
		return decimal.Zero
	}
	return short
}

// executeMoves runs every open move and reports what each of them processed.
func executeMoves(ctx corectx.Context, operation *transferOperationContext) ([]moveOutcome, error) {
	outcomes := make([]moveOutcome, 0, len(operation.Moves))

	for _, item := range operation.Moves {
		move := models.NewStockMoveFrom(item)
		if !IsMoveOpen(derefString(move.GetStatus())) {
			continue
		}
		outcome, err := executeOneMove(ctx, operation, *move)
		if err != nil {
			return nil, err
		}
		outcomes = append(outcomes, *outcome)
	}
	return outcomes, nil
}

// executeOneMove moves the stock one move's lines have reserved. A move with no lines processes
// nothing, which is not an error: its stock was never available and the backorder policy decides
// what happens to it. Incoming moves are the exception, handled by ensureIncomingLine.
func executeOneMove(
	ctx corectx.Context, operation *transferOperationContext, move models.StockMove,
) (*moveOutcome, error) {
	moveId := derefString(move.GetId())
	demand := orZero(move.GetBaseDemandQuantity())

	if err := ensureIncomingLine(ctx, operation, move); err != nil {
		return nil, err
	}

	lines, err := models.FindMoveLines(
		ctx, operation.MoveLineEngine.ResourceRepository(), moveId, models.MaxMoveLines)
	if err != nil {
		return nil, err
	}

	// Lock the source balances before any is read or written, so the whole move sees one consistent
	// picture and no other request can interleave.
	if _, err := LockQuantsForUpdate(
		ctx, operation.QuantEngine.ResourceRepository().GetBaseRepo(), QuantLockKey{
			OrgId:            derefString(move.GetOrgId()),
			ProductVariantId: derefString(move.GetProductVariantId()),
			LocationId:       derefString(move.GetSourceLocationId()),
		}); err != nil {
		return nil, err
	}

	processed := decimal.Zero
	for _, item := range lines {
		line := models.NewStockMoveLineFrom(item)
		quantity := orZero(line.GetBaseQuantity())
		if quantity.LessThanOrEqual(decimal.Zero) {
			continue
		}
		if err := shipOneLine(ctx, operation, move, *line, quantity); err != nil {
			return nil, err
		}
		processed = processed.Add(quantity)
	}

	next := models.StockMoveStatusDone
	if processed.IsZero() {
		// Nothing moved, so cancel rather than mark done: "done" must mean "this stock moved".
		next = models.StockMoveStatusCancelled
	}
	if err := updateMoveStatus(ctx, operation.MoveEngine, move, next); err != nil {
		return nil, err
	}

	return &moveOutcome{MoveId: moveId, Demand: demand, Processed: processed}, nil
}

// ensureIncomingLine gives an incoming move the line reservation could never make: an incoming move
// draws from a supplier, a virtual location that holds no balance, so nothing is reserved and
// without this the transfer would validate to zero. The line carries the full outstanding demand,
// because an incoming quantity is what the document says arrived rather than what a balance can
// spare; the supplier location then goes negative, which is how a virtual counterparty records what
// it supplied.
func ensureIncomingLine(
	ctx corectx.Context, operation *transferOperationContext, move models.StockMove,
) error {
	if derefString(operation.Transfer.GetOperationCode()) != models.StockOperationCodeIncoming {
		return nil
	}

	moveId := derefString(move.GetId())
	alreadyLined, err := reservedForMove(ctx, operation, moveId)
	if err != nil {
		return err
	}
	outstanding := orZero(move.GetBaseDemandQuantity()).Sub(alreadyLined)
	if outstanding.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	// The source balance must exist before shipOneLine can decrement it, and a supplier location has
	// never been counted.
	if _, err := ensureQuantForDimension(ctx, operation, models.QuantDimension{
		OrgId:            derefString(move.GetOrgId()),
		ProductVariantId: derefString(move.GetProductVariantId()),
		LocationId:       derefString(move.GetSourceLocationId()),
	}); err != nil {
		return err
	}

	line := dmodel.DynamicFields{
		models.StockMoveLineFieldMoveId:                moveId,
		models.StockMoveLineFieldTransferId:            derefString(move.GetTransferId()),
		models.StockMoveLineFieldProductVariantId:      derefString(move.GetProductVariantId()),
		models.StockMoveLineFieldQuantity:              outstanding.String(),
		models.StockMoveLineFieldBaseQuantity:          outstanding.String(),
		models.StockMoveLineFieldSourceLocationId:      derefString(move.GetSourceLocationId()),
		models.StockMoveLineFieldDestinationLocationId: derefString(move.GetDestinationLocationId()),
		models.StockMoveLineFieldLotRef:                "",
		models.StockMoveLineFieldPackageRef:            "",
		models.StockMoveLineFieldResultPackageRef:      "",
		models.StockMoveLineFieldOwnerRef:              "",
		models.StockMoveLineFieldOrgId:                 derefString(move.GetOrgId()),
	}
	_, err = operation.MoveLineEngine.ResourceService().Create(ctx, line)
	return errors.Wrap(err, "ensureIncomingLine")
}

// shipOneLine moves one line's quantity from the source balance to the destination balance. Both
// sides must stay in the same transaction; splitting them allows a partial commit in which stock
// evaporates.
func shipOneLine(
	ctx corectx.Context,
	operation *transferOperationContext,
	move models.StockMove,
	line models.StockMoveLine,
	quantity decimal.Decimal,
) error {
	orgId := derefString(move.GetOrgId())

	sourceId, err := findQuantForLine(ctx, operation, move, line)
	if err != nil {
		return err
	}
	if sourceId == "" {
		return errors.Errorf(
			"stock move line '%s' has no source balance to take from", derefString(line.GetId()))
	}

	// Source: the reservation is consumed and the goods leave.
	if err := applyQuantDelta(ctx, operation, sourceId, quantity.Neg(), quantity.Neg()); err != nil {
		return err
	}

	// Destination: the goods arrive unreserved. The balance usually does not exist yet, so it is
	// created; assuming an update here would silently move stock into nowhere.
	destinationId, err := ensureDestinationQuant(ctx, operation, move, line, orgId)
	if err != nil {
		return err
	}
	if err := applyQuantDelta(ctx, operation, destinationId, quantity, decimal.Zero); err != nil {
		return err
	}

	return stampLineExecuted(ctx, operation, line)
}

// ensureDestinationQuant finds or creates the balance the goods are arriving at.
func ensureDestinationQuant(
	ctx corectx.Context,
	operation *transferOperationContext,
	move models.StockMove,
	line models.StockMoveLine,
	orgId string,
) (model.Id, error) {
	// Goods keep their lot and owner across the move; the package may change when the operation
	// repacks them, which is what result_package_ref records.
	dimension := models.QuantDimension{
		OrgId:            orgId,
		ProductVariantId: derefString(line.GetProductVariantId()),
		LocationId:       derefString(line.GetDestinationLocationId()),
		LotRef:           derefString(line.GetLotRef()),
		PackageRef:       derefString(line.GetPackageRef()),
		OwnerRef:         derefString(line.GetOwnerRef()),
	}

	return ensureQuantForDimension(ctx, operation, dimension)
}

// ensureQuantForDimension returns the balance at an exact dimension, creating an empty one if none
// exists. The new row starts at zero and the caller's delta moves it, so the arithmetic is the same
// either way.
func ensureQuantForDimension(
	ctx corectx.Context, operation *transferOperationContext, dimension models.QuantDimension,
) (model.Id, error) {
	found, err := models.FindQuantForDimension(ctx, operation.QuantEngine.ResourceRepository(), dimension)
	if err != nil {
		return "", err
	}
	if len(found) > 0 {
		return derefString(models.NewStockQuantFrom(found[0]).GetId()), nil
	}

	_, err = operation.QuantEngine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
		models.StockQuantFieldProductVariantId: dimension.ProductVariantId,
		models.StockQuantFieldLocationId:       dimension.LocationId,
		models.StockQuantFieldLotRef:           dimension.LotRef,
		models.StockQuantFieldPackageRef:       dimension.PackageRef,
		models.StockQuantFieldOwnerRef:         dimension.OwnerRef,
		models.StockQuantFieldOnHandQuantity:   "0",
		models.StockQuantFieldReservedQuantity: "0",
		models.StockQuantFieldIncomingDate:     time.Now().UTC(),
		models.StockQuantFieldOrgId:            dimension.OrgId,
	})
	if err != nil {
		return "", errors.Wrap(err, "ensureQuantForDimension")
	}

	// Re-read to get the id the insert generated, by the same unique dimension.
	found, err = models.FindQuantForDimension(ctx, operation.QuantEngine.ResourceRepository(), dimension)
	if err != nil {
		return "", err
	}
	if len(found) == 0 {
		return "", errors.New("a stock balance could not be read back after being created")
	}
	return derefString(models.NewStockQuantFrom(found[0]).GetId()), nil
}

// applyQuantDelta adds to a balance's on-hand and reserved figures in one write. It is the only
// writer of on_hand_quantity, and callers must already hold the row lock on the balance so the
// read below cannot be overtaken before the write.
func applyQuantDelta(
	ctx corectx.Context,
	operation *transferOperationContext,
	quantId model.Id,
	onHandDelta decimal.Decimal,
	reservedDelta decimal.Decimal,
) error {
	found, err := operation.QuantEngine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.StockQuantFieldId: quantId,
	})
	if err != nil {
		return errors.Wrap(err, "applyQuantDelta")
	}
	if found == nil || !found.HasData {
		return errors.Errorf("stock quant '%s' vanished mid-transaction", quantId)
	}

	quant := models.NewStockQuantFrom(found.Data)
	nextOnHand := orZero(quant.GetOnHandQuantity()).Add(onHandDelta)
	nextReserved := orZero(quant.GetReservedQuantity()).Add(reservedDelta)

	if nextReserved.LessThan(decimal.Zero) {
		// Reserved quantity must never go negative. Reaching here means a line claimed more than the
		// balance had reserved, which is a bug in the reservation bookkeeping, not a client mistake.
		return errors.Errorf(
			"validating would drive stock quant '%s' to a negative reserved quantity", quantId)
	}

	_, err = operation.QuantEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.StockQuantFieldId:               quantId,
		models.StockQuantFieldOnHandQuantity:   nextOnHand.String(),
		models.StockQuantFieldReservedQuantity: nextReserved.String(),
		basemodel.FieldEtag:                    derefString(quant.GetEtag()),
	})
	return errors.Wrap(err, "applyQuantDelta")
}

// stampLineExecuted marks a move line as a recorded movement rather than a reservation.
func stampLineExecuted(
	ctx corectx.Context, operation *transferOperationContext, line models.StockMoveLine,
) error {
	_, err := operation.MoveLineEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.StockMoveLineFieldId:          derefString(line.GetId()),
		models.StockMoveLineFieldPicked:      true,
		models.StockMoveLineFieldOperationAt: time.Now().UTC(),
		basemodel.FieldEtag:                  derefString(line.GetEtag()),
	})
	return errors.Wrap(err, "stampLineExecuted")
}

// finishValidate handles the unprocessed remainder and closes the transfer.
func finishValidate(
	ctx corectx.Context,
	operation *transferOperationContext,
	outcomes []moveOutcome,
	idempotencyKey string,
	createBackorder *bool,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	decision, vErrs := DecideBackorder(
		derefString(operation.Transfer.GetBackorderPolicy()), outcomes, createBackorder)
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if decision == BackorderCreate {
		if err := createBackorderTransfer(ctx, operation, outcomes); err != nil {
			return nil, err
		}
	}

	if err := closeTransfer(ctx, operation, idempotencyKey); err != nil {
		return nil, err
	}
	return mutateOk(), nil
}

// closeTransfer marks the transfer done, stamping the completion time and the idempotency key. The
// key must be written in the same transaction as the movements it guards, or a retry cannot tell
// that the stock already moved.
func closeTransfer(
	ctx corectx.Context, operation *transferOperationContext, idempotencyKey string,
) error {
	current := derefString(operation.Transfer.GetStatus())
	if !CanTransitionTransfer(current, models.StockTransferStatusDone) {
		return errors.Errorf("a transfer cannot go from '%s' to 'done'", current)
	}

	update := dmodel.DynamicFields{
		models.StockTransferFieldId:          derefString(operation.Transfer.GetId()),
		models.StockTransferFieldStatus:      models.StockTransferStatusDone,
		models.StockTransferFieldCompletedAt: time.Now().UTC(),
		basemodel.FieldEtag:                  derefString(operation.Transfer.GetEtag()),
	}
	if idempotencyKey != "" {
		update[models.StockTransferFieldIdempotencyKey] = idempotencyKey
	}

	_, err := operation.TransferEngine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "closeTransfer")
}
