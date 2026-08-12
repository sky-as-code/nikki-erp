package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Reservation: claiming stock for a demand without moving it.
//
// Everything here changes reserved_quantity and nothing changes on_hand_quantity (AC-STOCK-005).
// A reservation is a promise about stock that is still physically where it was; only validate
// moves it.

// MoveAvailability reports what one move needs and what it can get.
type MoveAvailability struct {
	MoveId    model.Id        `json:"move_id"`
	VariantId model.Id        `json:"product_variant_id"`
	Demand    decimal.Decimal `json:"demand_quantity"`
	Reserved  decimal.Decimal `json:"reserved_quantity"`
	Available decimal.Decimal `json:"available_quantity"`
	Shortage  decimal.Decimal `json:"shortage_quantity"`
}

// TransferAvailability is the answer to a check-availability request.
type TransferAvailability struct {
	TransferId  model.Id           `json:"transfer_id"`
	Moves       []MoveAvailability `json:"moves"`
	HasShortage bool               `json:"has_shortage"`
}

// CheckAvailability reports whether a transfer's demand can be met, and takes nothing.
//
// It is deliberately read-only, and that is the whole point of it being a separate operation: it
// acquires no lock, writes no status and creates no move line, so the stock it reports as available
// may be reserved by someone else a moment later (BR §4.2.3.7, AC-STOCK-033). A caller that needs
// the stock must reserve it; this only answers "is it worth trying".
//
// Because it takes no ownership it also needs no transaction. Adding one would suggest a guarantee
// it does not provide.
func (this *StockTransferDomainServiceImpl) CheckAvailability(
	ctx corectx.Context, transferId string,
) (*dyn.OpResult[any], error) {
	operation, err := loadTransferOperation(ctx, transferId)
	if err != nil {
		return nil, err
	}
	if operation == nil {
		notFound := notFoundResult(transferId)
		return &dyn.OpResult[any]{ClientErrors: notFound.ClientErrors}, nil
	}

	report := TransferAvailability{
		TransferId: transferId,
		Moves:      make([]MoveAvailability, 0, len(operation.Moves)),
	}

	for _, item := range operation.Moves {
		move := models.NewStockMoveFrom(item)
		if !IsMoveOpen(derefString(move.GetStatus())) {
			continue
		}
		line, err := checkMoveAvailability(ctx, operation, *move)
		if err != nil {
			return nil, err
		}
		if line.Shortage.GreaterThan(decimal.Zero) {
			report.HasShortage = true
		}
		report.Moves = append(report.Moves, *line)
	}

	return &dyn.OpResult[any]{Data: report, HasData: true}, nil
}

// checkMoveAvailability sums the unreserved stock at the move's source, without locking it.
func checkMoveAvailability(
	ctx corectx.Context, operation *transferOperationContext, move models.StockMove,
) (*MoveAvailability, error) {
	demand := orZero(move.GetBaseDemandQuantity())

	reserved, err := reservedForMove(ctx, operation, derefString(move.GetId()))
	if err != nil {
		return nil, err
	}

	quants, err := models.FindQuantsAtLocation(
		ctx, operation.QuantEngine.ResourceRepository(),
		derefString(move.GetOrgId()),
		derefString(move.GetProductVariantId()),
		derefString(move.GetSourceLocationId()),
	)
	if err != nil {
		return nil, err
	}

	available := decimal.Zero
	for _, item := range quants {
		quant := models.NewStockQuantFrom(item)
		rowAvailable := AvailableQuantity(quant.GetOnHandQuantity(), quant.GetReservedQuantity())
		if rowAvailable.GreaterThan(decimal.Zero) {
			available = available.Add(rowAvailable)
		}
	}

	outstanding := demand.Sub(reserved)
	shortage := outstanding.Sub(available)
	if shortage.LessThan(decimal.Zero) {
		shortage = decimal.Zero
	}

	return &MoveAvailability{
		MoveId:    derefString(move.GetId()),
		VariantId: derefString(move.GetProductVariantId()),
		Demand:    demand,
		Reserved:  reserved,
		Available: available,
		Shortage:  shortage,
	}, nil
}

// Reserve claims stock for every open move of a transfer.
//
// The sequence per move is fixed by BR §8.6 and is the reason [INV-STK-204] exists: lock the
// candidate balances, recompute availability from what the lock returned, then allocate. A figure
// read before the lock is stale by definition, so none is carried across.
//
// Reserving an already-reserved transfer is a no-op rather than a second allocation: each move
// asks only for its outstanding quantity, which is zero once it is fully assigned.
func (this *StockTransferDomainServiceImpl) Reserve(
	ctx corectx.Context, transferId string,
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
		if !IsTransferOpen(derefString(operation.Transfer.GetStatus())) {
			result = violationResult(
				"stock_transfer.not_open",
				"a completed or cancelled transfer cannot reserve stock")
			return nil
		}

		if _, err := reserveTransferMoves(tranxCtx, operation); err != nil {
			return err
		}
		result, err = recomputeTransferFromMoves(tranxCtx, operation, transferId)
		return err
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// reserveTransferMoves allocates for each open move and returns the total it managed to claim.
func reserveTransferMoves(
	ctx corectx.Context, operation *transferOperationContext,
) (decimal.Decimal, error) {
	total := decimal.Zero

	// An incoming transfer draws from a supplier, which holds no balance to claim: there is
	// nothing to reserve, and locking it would be work with no possible outcome. Its lines are
	// written by validate instead — see ensureIncomingLine.
	if derefString(operation.Transfer.GetOperationCode()) == models.StockOperationCodeIncoming {
		return total, nil
	}

	for _, item := range operation.Moves {
		move := models.NewStockMoveFrom(item)
		if !IsMoveOpen(derefString(move.GetStatus())) {
			continue
		}
		claimed, err := reserveOneMove(ctx, operation, *move)
		if err != nil {
			return decimal.Zero, err
		}
		total = total.Add(claimed)
	}
	return total, nil
}

// reserveOneMove locks the move's source balances and claims what it still needs.
func reserveOneMove(
	ctx corectx.Context, operation *transferOperationContext, move models.StockMove,
) (decimal.Decimal, error) {
	moveId := derefString(move.GetId())

	alreadyReserved, err := reservedForMove(ctx, operation, moveId)
	if err != nil {
		return decimal.Zero, err
	}
	outstanding := orZero(move.GetBaseDemandQuantity()).Sub(alreadyReserved)
	if outstanding.LessThanOrEqual(decimal.Zero) {
		// Fully reserved already. Re-entrant by construction rather than by a flag.
		return decimal.Zero, nil
	}

	// The lock, then the arithmetic. Never the other way around.
	locked, err := LockQuantsForUpdate(ctx, operation.QuantEngine.ResourceRepository().GetBaseRepo(), QuantLockKey{
		OrgId:            derefString(move.GetOrgId()),
		ProductVariantId: derefString(move.GetProductVariantId()),
		LocationId:       derefString(move.GetSourceLocationId()),
	})
	if err != nil {
		return decimal.Zero, err
	}

	allocations, _ := AllocateFromQuants(outstanding, locked)
	for _, allocation := range allocations {
		if err := applyReservation(ctx, operation, allocation, move); err != nil {
			return decimal.Zero, err
		}
	}

	claimed := TotalAllocated(allocations)
	next := DeriveMoveStatus(
		derefString(move.GetStatus()), orZero(move.GetBaseDemandQuantity()), alreadyReserved.Add(claimed))
	if err := updateMoveStatus(ctx, operation.MoveEngine, move, next); err != nil {
		return decimal.Zero, err
	}
	return claimed, nil
}

// applyReservation raises a balance's reserved quantity and records the claim as a move line.
//
// Both writes are needed and neither is redundant: the quant says how much of the balance is
// spoken for, and the move line says who spoke for it. Without the line there is no way to release
// exactly this reservation later; without the quant another demand would be told the stock is free.
func applyReservation(
	ctx corectx.Context, operation *transferOperationContext, allocation Allocation, move models.StockMove,
) error {
	if err := addToQuantReserved(ctx, operation.QuantEngine, allocation.QuantId, allocation.Quantity); err != nil {
		return err
	}

	line := dmodel.DynamicFields{
		models.StockMoveLineFieldMoveId:                derefString(move.GetId()),
		models.StockMoveLineFieldTransferId:            derefString(move.GetTransferId()),
		models.StockMoveLineFieldProductVariantId:      derefString(move.GetProductVariantId()),
		models.StockMoveLineFieldQuantity:              allocation.Quantity.String(),
		models.StockMoveLineFieldBaseQuantity:          allocation.Quantity.String(),
		models.StockMoveLineFieldSourceLocationId:      derefString(move.GetSourceLocationId()),
		models.StockMoveLineFieldDestinationLocationId: derefString(move.GetDestinationLocationId()),
		models.StockMoveLineFieldLotRef:                allocation.LotRef,
		models.StockMoveLineFieldPackageRef:            allocation.PackageRef,
		models.StockMoveLineFieldResultPackageRef:      allocation.PackageRef,
		models.StockMoveLineFieldOwnerRef:              allocation.OwnerRef,
		models.StockMoveLineFieldOrgId:                 derefString(move.GetOrgId()),
	}
	_, err := operation.MoveLineEngine.ResourceService().Create(ctx, line)
	return errors.Wrap(err, "applyReservation")
}

// addToQuantReserved adds a delta to a balance's reserved quantity.
//
// The read is safe without a further lock because the caller already holds a row lock on this
// quant: LockQuantsForUpdate ran in the same transaction, so no other writer can be between this
// read and this write.
//
// It writes reserved_quantity even though the field is `no_update` in the schema. That flag is
// enforced by ModelSchema.Validate in the service layer, which is what closes the field to clients;
// the repository writes what it is given, which is how a balance can be changed by a movement and
// only by a movement.
func addToQuantReserved(
	ctx corectx.Context, quantEngine drif.DynamicResourceEngine, quantId model.Id, delta decimal.Decimal,
) error {
	found, err := quantEngine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.StockQuantFieldId: quantId,
	})
	if err != nil {
		return errors.Wrap(err, "addToQuantReserved")
	}
	if found == nil || !found.HasData {
		return errors.Errorf("stock quant '%s' vanished while it was locked", quantId)
	}

	quant := models.NewStockQuantFrom(found.Data)
	next := orZero(quant.GetReservedQuantity()).Add(delta)
	if next.LessThan(decimal.Zero) {
		// STOCK-INV-002. Reaching here means the caller's bookkeeping disagrees with the stored
		// balance, which is a bug rather than a client mistake, so it fails loudly.
		return errors.Errorf(
			"releasing %s from stock quant '%s' would drive its reserved quantity negative", delta.String(), quantId)
	}

	_, err = quantEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.StockQuantFieldId:               quantId,
		models.StockQuantFieldReservedQuantity: next.String(),
		basemodel.FieldEtag:                    derefString(quant.GetEtag()),
	})
	return errors.Wrap(err, "addToQuantReserved")
}

// reservedForMove sums the move lines already standing against a move, which is how much of its
// demand is currently claimed.
func reservedForMove(
	ctx corectx.Context, operation *transferOperationContext, moveId string,
) (decimal.Decimal, error) {
	lines, err := models.FindMoveLines(
		ctx, operation.MoveLineEngine.ResourceRepository(), moveId, models.MaxMoveLines)
	if err != nil {
		return decimal.Zero, err
	}

	total := decimal.Zero
	for _, item := range lines {
		line := models.NewStockMoveLineFrom(item)
		total = total.Add(orZero(line.GetBaseQuantity()))
	}
	return total, nil
}

// Unreserve gives back everything a transfer is holding, without moving any stock.
//
// On-hand is untouched: the goods never went anywhere, so releasing the claim restores the
// availability and nothing else (BR §4.2.3.9).
func (this *StockTransferDomainServiceImpl) Unreserve(
	ctx corectx.Context, transferId string,
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
		if !IsTransferOpen(derefString(operation.Transfer.GetStatus())) {
			result = violationResult(
				"stock_transfer.not_open",
				"a completed or cancelled transfer holds no reservation to release")
			return nil
		}

		if err := unreserveTransferMoves(tranxCtx, operation); err != nil {
			return err
		}
		result, err = recomputeTransferFromMoves(tranxCtx, operation, transferId)
		return err
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// unreserveTransferMoves releases every open move's claim and deletes the lines recording it.
func unreserveTransferMoves(ctx corectx.Context, operation *transferOperationContext) error {
	for _, item := range operation.Moves {
		move := models.NewStockMoveFrom(item)
		if !IsMoveOpen(derefString(move.GetStatus())) {
			// A done move's lines are its recorded movement, not a reservation: deleting them
			// would erase the history of stock that has already moved.
			continue
		}
		if err := releaseMoveLines(ctx, operation, *move); err != nil {
			return err
		}
		if err := updateMoveStatus(ctx, operation.MoveEngine, *move, models.StockMoveStatusConfirmed); err != nil {
			return err
		}
	}
	return nil
}

// releaseMoveLines gives each of a move's reservations back to the balance it came from.
func releaseMoveLines(
	ctx corectx.Context, operation *transferOperationContext, move models.StockMove,
) error {
	lines, err := models.FindMoveLines(
		ctx, operation.MoveLineEngine.ResourceRepository(), derefString(move.GetId()), models.MaxMoveLines)
	if err != nil {
		return err
	}

	for _, item := range lines {
		line := models.NewStockMoveLineFrom(item)
		quantId, err := findQuantForLine(ctx, operation, move, *line)
		if err != nil {
			return err
		}
		if quantId != "" {
			released := orZero(line.GetBaseQuantity()).Neg()
			if err := addToQuantReserved(ctx, operation.QuantEngine, quantId, released); err != nil {
				return err
			}
		}
		if _, err := operation.MoveLineEngine.ResourceRepository().DeleteOne(ctx, dmodel.DynamicFields{
			models.StockMoveLineFieldId: derefString(line.GetId()),
		}); err != nil {
			return errors.Wrap(err, "releaseMoveLines")
		}
	}
	return nil
}

// findQuantForLine locates the balance a move line drew from, by its full dimension key.
//
// A line records where it took the stock from, so the release goes back to exactly that balance
// rather than to whichever row happens to be first: returning it to the wrong lot would leave both
// rows wrong while their total stayed right.
func findQuantForLine(
	ctx corectx.Context, operation *transferOperationContext, move models.StockMove, line models.StockMoveLine,
) (model.Id, error) {
	found, err := models.FindQuantForDimension(ctx, operation.QuantEngine.ResourceRepository(),
		models.QuantDimension{
			OrgId:            derefString(move.GetOrgId()),
			ProductVariantId: derefString(line.GetProductVariantId()),
			LocationId:       derefString(line.GetSourceLocationId()),
			LotRef:           derefString(line.GetLotRef()),
			PackageRef:       derefString(line.GetPackageRef()),
			OwnerRef:         derefString(line.GetOwnerRef()),
		})
	if err != nil {
		return "", err
	}
	if len(found) == 0 {
		return "", nil
	}
	return derefString(models.NewStockQuantFrom(found[0]).GetId()), nil
}

// recomputeTransferFromMoves re-reads the moves and writes the state they add up to.
func recomputeTransferFromMoves(
	ctx corectx.Context, operation *transferOperationContext, transferId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	refreshed, err := models.FindTransferMoves(
		ctx, operation.MoveEngine.ResourceRepository(), transferId, models.MaxTransferMoves)
	if err != nil {
		return nil, err
	}

	current := derefString(operation.Transfer.GetStatus())
	next := DeriveTransferStatus(current, moveStatuses(refreshed))
	if failed, err := updateTransferStatus(ctx, operation.TransferEngine, operation.Transfer, next); err != nil {
		return nil, err
	} else if failed != nil {
		return failed, nil
	}
	return mutateOk(), nil
}
