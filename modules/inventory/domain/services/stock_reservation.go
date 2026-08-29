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

// Reservation claims stock for a demand without moving it: everything here changes
// reserved_quantity and nothing changes on_hand_quantity. Only validate moves stock.

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

// CheckAvailability reports whether a transfer's demand can be met, and claims nothing. It takes no
// lock and needs no transaction, so the stock it reports as available may be reserved by someone
// else a moment later; a caller that needs the stock must reserve it.
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

// Reserve claims stock for every open move of a transfer. The order per move is fixed: lock the
// candidate balances, recompute availability from what the lock returned, then allocate — a figure
// read before the lock is stale by definition. Reserving an already-reserved transfer is a no-op,
// since each move asks only for its outstanding quantity.
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

	// An incoming transfer draws from a supplier, which holds no balance to claim, so there is
	// nothing to reserve; its lines are written by validate instead — see ensureIncomingLine.
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
		// Fully reserved already; re-entrant by construction rather than by a flag.
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

// applyReservation raises a balance's reserved quantity and records the claim as a move line. Both
// writes are needed: without the line there is no way to release exactly this reservation later,
// and without the quant another demand would be told the stock is free.
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

// addToQuantReserved adds a delta to a balance's reserved quantity. Callers must already hold the
// row lock on the quant, so no other writer can slip between this read and this write. It writes
// reserved_quantity although the field is `no_update`: that flag is enforced in the service layer
// to close the field to clients, while the repository writes what it is given.
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
		// Reserved quantity must never go negative. Reaching here means the caller's bookkeeping
		// disagrees with the stored balance, which is a bug, so it fails loudly.
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

// Unreserve gives back everything a transfer is holding. On-hand is untouched: the goods never
// moved, so releasing the claim restores availability and nothing else.
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
			// A done move's lines are its recorded movement, not a reservation; deleting them would
			// erase the history of stock that has already moved.
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

// findQuantForLine locates the balance a move line drew from, by its full dimension key, so a
// release goes back to exactly that balance: returning it to the wrong lot leaves both rows wrong
// while their total stays right.
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
