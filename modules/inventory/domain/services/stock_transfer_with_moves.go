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

// CreateWithMoves raises a draft transfer and its moves in one transaction.
//
// The halves must not be separable: the move engine is unpublished, so a consumer could raise a
// header with no supported way to fill it, and an empty transfer VALIDATES SUCCESSFULLY, reporting
// goods moved that were never named.
//
// Locations, status, sequence and base quantity are derived here from the transfer, which took them
// from the operation type; the caller names only a variant, a unit and an amount.
func (this *StockTransferDomainServiceImpl) CreateWithMoves(
	ctx corectx.Context, params dmodel.DynamicFields, moves []itStock.TransferMoveRequest,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	if len(moves) == 0 {
		// Refused rather than accepted-and-empty: an empty transfer is worse than none, because it
		// validates successfully.
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
		// The header goes through Create, not the repository, so the operation type's defaults, the
		// generated number and the archived-type refusal all apply.
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

			// Base equals demand: no UoM conversion exists here yet. It must be written explicitly,
			// because reservation and validation both read the BASE quantity and a move without one
			// would reserve nothing and appear fully satisfied.
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

// assertMovesWellFormed refuses a line that could not move anything, before the transfer is created
// so a bad line leaves no orphan header. A zero or negative quantity would insert cleanly, reserve
// nothing and validate as complete — a move reporting success having moved no goods.
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
