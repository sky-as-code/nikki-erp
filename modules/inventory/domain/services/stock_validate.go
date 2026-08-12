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

// Validate: the one operation in the whole module that changes an on-hand quantity.
//
// Everything else — confirm, reserve, unreserve, cancel — moves documents and claims around. This
// records that goods physically moved, which is why it is the only place a quant's on-hand figure
// is written, and why the whole thing is one transaction: a source decremented without its
// destination incremented is stock that has ceased to exist (AC-STOCK-006, AC-STOCK-035).

// Validate executes a transfer: it consumes the reservations, moves the stock and closes the
// transfer.
//
// The sequence, all inside one transaction (BR §4.2.3.10, §8.4):
//
//  1. Re-read the transfer and refuse it if it is already closed.
//  2. If it carries the caller's idempotency key and is done, return the earlier result untouched.
//  3. For each open move: lock its source balances, re-validate the quantities inside the lock,
//     decrement the source, increment the destination, stamp the lines and close the move.
//  4. Handle whatever was not processed, per the snapshot backorder policy.
//  5. Close the transfer and stamp completed_at.
//
// Every quantity is re-read inside the lock. A figure fetched before it is stale by definition,
// however few milliseconds ago it was read.
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

// replayedValidate detects a retry of a validate that already succeeded.
//
// This is BR §8.7, and it guards the one failure that cannot be repaired afterwards: a client whose
// request timed out retries, and the goods ship twice. An edit can fix a wrong number in a record;
// nothing can un-ship a second delivery.
//
// It matches only a transfer that is already `done`. A key on a transfer still in flight means the
// earlier attempt did not complete, so the retry should proceed rather than report a success that
// never happened.
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

// executeOneMove moves the stock one move's lines have reserved.
//
// A move with no lines processes nothing. That is not an error: it is a move whose stock was never
// available, and the backorder policy decides what happens to it.
//
// The exception is an incoming move, which is handled by ensureIncomingLine below.
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

	// The source balances are locked before any of them is read or written, so that the whole move
	// sees one consistent picture and no other request can interleave with it.
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
		// Nothing moved, so there is no movement to record. Cancelling rather than marking it done
		// keeps "done" meaning "this stock moved" (STOCK-INV-020).
		next = models.StockMoveStatusCancelled
	}
	if err := updateMoveStatus(ctx, operation.MoveEngine, move, next); err != nil {
		return nil, err
	}

	return &moveOutcome{MoveId: moveId, Demand: demand, Processed: processed}, nil
}

// ensureIncomingLine gives an incoming move the line that reservation could never have made.
//
// Reservation claims stock that is already on hand somewhere. An incoming move draws from a
// supplier — a virtual location that holds no balance and never will — so there is nothing to
// reserve and no line gets created. Without this an incoming transfer would validate to zero and
// no receipt could ever bring stock in, which would leave the whole module unable to acquire
// anything (BR §4.2.1.2, and §4.2.3.10's "actual quantity" precondition).
//
// The line it writes carries the full outstanding demand, because an incoming quantity is what the
// document says arrived rather than what a balance could spare. shipOneLine then applies it to both
// ends as usual: the supplier location goes negative, which is exactly how a virtual counterparty
// records what it has supplied.
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

	// The source balance has to exist before shipOneLine can decrement it, and for a supplier
	// location it will not: nothing has ever been counted there.
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

// shipOneLine moves one line's quantity from the source balance to the destination balance.
//
// Both sides happen here, adjacent and in the same transaction, because they are one fact about
// the world: the stock left one place and arrived at another. Splitting them across functions or
// transactions is what allows a partial commit in which stock evaporates.
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

	// Destination: the goods arrive, unreserved. It usually does not exist yet, so it is created
	// rather than updated — assuming an update here would silently move stock into nowhere.
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
	// The goods keep their lot and owner across the move; the package may change when the operation
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
// exists yet.
//
// Creating rather than assuming an update is what makes a first receipt into a fresh location work
// at all: the destination balance usually does not exist, and neither does a supplier's. The row
// starts at zero and the caller's delta moves it, so the arithmetic is the same either way.
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

// applyQuantDelta adds to a balance's on-hand and reserved figures in one write.
//
// It is the only function that writes on_hand_quantity, and it is reachable only from validate.
// The caller holds a row lock on the balance, so the read below cannot be overtaken between the
// read and the write.
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
		// STOCK-INV-002. Reaching here means a line claimed more than the balance had reserved,
		// which is a bug in the reservation bookkeeping rather than a client mistake.
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

// closeTransfer marks the transfer done, stamping the completion time and the idempotency key.
//
// The key is written in the same transaction as the movements it guards. Written afterwards it
// would leave a window in which the stock has moved but a retry cannot tell.
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
