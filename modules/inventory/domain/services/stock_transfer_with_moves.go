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

// CreateWithMoves raises a draft transfer and its moves in one transaction (SALES-049).
//
// # Why the two halves must not be separable
//
// Create writes the header alone. Inside Inventory that is fine — the engine action that follows it
// writes the moves through the move engine, and both are Inventory's own code. From outside it is
// not: the move engine is unpublished, so a consumer could raise a header and have no supported way
// to fill it. An empty transfer is the dangerous shape, because it VALIDATES SUCCESSFULLY and
// reports goods moved that were never named. Committing both together makes that unreachable
// rather than merely discouraged.
//
// # What the caller does not get to say
//
// Locations, status, sequence and base quantity are all derived here. The caller names a variant, a
// unit and an amount; everything about where the goods sit and what state the move is in comes from
// the transfer, which took it from the operation type. A caller cannot assert a location it has no
// way to know is right.
func (this *StockTransferDomainServiceImpl) CreateWithMoves(
	ctx corectx.Context, params dmodel.DynamicFields, moves []itStock.TransferMoveRequest,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	if len(moves) == 0 {
		// Refused rather than accepted-and-empty. See above: an empty transfer is worse than no
		// transfer, because it succeeds.
		return transferViolation(
			"stock_transfer.no_moves",
			"a transfer must be created with at least one move",
		), nil
	}
	if result := assertMovesWellFormed(moves); result != nil {
		return result, nil
	}

	var created *dyn.OpResult[dmodel.DynamicFields]

	err := withTransferTransaction(ctx, func(tranxCtx corectx.Context) error {
		// The header goes through Create rather than straight to the repository, so that the
		// operation type's defaults, the generated number and the archived-type refusal all apply
		// exactly as they do for a transfer raised from a request.
		result, err := this.Create(tranxCtx, params)
		if err != nil {
			return err
		}
		if result.ClientErrors.Count() > 0 || !result.HasData {
			created = result
			return nil
		}

		transfer := models.NewStockTransferFrom(result.Data)
		transferId := derefString(transfer.GetId())
		if transferId == "" {
			return errors.New("the created stock transfer carries no id")
		}

		if err := this.writeTransferMoves(tranxCtx, *transfer, transferId, moves); err != nil {
			return err
		}

		created = result
		return nil
	})

	if err != nil {
		return nil, err
	}
	return created, nil
}

// writeTransferMoves inserts one move per requested line, deriving what the caller may not set.
func (this *StockTransferDomainServiceImpl) writeTransferMoves(
	ctx corectx.Context,
	transfer models.StockTransfer,
	transferId string,
	moves []itStock.TransferMoveRequest,
) error {
	engine, err := engineFor(models.StockMoveSchemaName)
	if err != nil {
		return err
	}

	for index, move := range moves {
		quantity := move.Quantity.String()

		fields := dmodel.DynamicFields{
			models.StockMoveFieldTransferId:       transferId,
			models.StockMoveFieldSequence:         index + 1,
			models.StockMoveFieldProductVariantId: move.ProductVariantId,
			models.StockMoveFieldDemandQuantity:   quantity,

			// Base equals demand, which is what every other move-writing path in this module does:
			// no UoM conversion exists here yet. Written explicitly rather than left absent because
			// reservation and validation both read the BASE quantity — a move without one would
			// reserve nothing and appear fully satisfied.
			models.StockMoveFieldBaseDemandQuantity: quantity,

			// Copied from the transfer, which took them from the operation type's defaults.
			models.StockMoveFieldSourceLocationId:      derefString(transfer.GetSourceLocationId()),
			models.StockMoveFieldDestinationLocationId: derefString(transfer.GetDestinationLocationId()),

			models.StockMoveFieldStatus: models.StockMoveStatusDraft,
			models.StockMoveFieldOrgId:  derefString(transfer.GetOrgId()),
		}
		if move.UomId != "" {
			fields[models.StockMoveFieldUomId] = move.UomId
		}

		if _, err := engine.ResourceRepository().Insert(ctx, fields); err != nil {
			return errors.Wrapf(err, "creating move %d of transfer '%s'", index+1, transferId)
		}
	}
	return nil
}

// assertMovesWellFormed refuses a line that could not move anything.
//
// Checked before the transfer is created rather than after, so a bad line leaves no orphan header
// behind. A zero or negative quantity is the case worth naming: it would insert cleanly, reserve
// nothing, and validate as complete — a move that reports success having moved no goods.
func assertMovesWellFormed(moves []itStock.TransferMoveRequest) *dyn.OpResult[dmodel.DynamicFields] {
	vErrs := ft.NewClientErrors()

	for index, move := range moves {
		position := decimal.NewFromInt(int64(index + 1)).String()

		if move.ProductVariantId == "" {
			vErrs.Append(*ft.NewBusinessViolation(
				models.StockMoveSchemaName,
				"stock_move.product_variant_required",
				"move "+position+" names no product variant",
			))
		}
		if move.Quantity.LessThanOrEqual(decimal.Zero) {
			vErrs.Append(*ft.NewBusinessViolation(
				models.StockMoveSchemaName,
				"stock_move.quantity_must_be_positive",
				"move "+position+" must ask for a quantity greater than zero",
			))
		}
	}

	if vErrs.Count() == 0 {
		return nil
	}
	return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}
}

// transferViolation is a refusal carrying no data.
func transferViolation(key, message string) *dyn.OpResult[dmodel.DynamicFields] {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName, key, message))
	return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}
}
