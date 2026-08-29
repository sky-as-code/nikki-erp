package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Corrections (adjustments and scraps) generate a lightweight internal transfer and run it through
// the validate path rather than writing a quant directly, because applyQuantDelta in
// stock_validate.go is the only writer of on_hand_quantity and a second writer would drift from it.

// correctionOperationCode is internal: a correction moves between two of the company's own
// locations, an internal one and a virtual counterparty.
const correctionOperationCode = models.StockOperationCodeInternal

// CorrectionRequest describes a balance correction to apply. Both locations are named explicitly
// rather than derived from a direction flag; swapping them silently inverts the sign of the
// movement.
type CorrectionRequest struct {
	OrgId                 string
	ProductVariantId      string
	Quantity              decimal.Decimal
	SourceLocationId      string
	DestinationLocationId string
	LotRef                string
	PackageRef            string
	OwnerRef              string
	OriginReference       string
	Note                  string

	// IsInventoryAdjustment separates a count correction from a scrap in movement history.
	IsInventoryAdjustment bool
}

// CorrectionResult reports what the correction generated, so the caller can record the link back.
type CorrectionResult struct {
	TransferId string
	MoveId     string
}

// ApplyCorrectionMovement generates a done internal transfer for one quantity of one variant.
//
// It must run inside the caller's transaction and never opens its own: BeginTx returns ErrTxNested,
// and callers already hold the source quant locked FOR UPDATE. A missing transaction fails loudly.
func ApplyCorrectionMovement(
	ctx corectx.Context, request CorrectionRequest,
) (*CorrectionResult, *ft.ClientErrors, error) {
	if vErrs := assertCorrectable(request); vErrs.Count() > 0 {
		return nil, vErrs, nil
	}
	if ctx.GetDbTranx() == nil {
		return nil, nil, errors.New(
			"ApplyCorrectionMovement must run inside a transaction: its caller holds the source " +
				"quant locked, and a correction that committed separately could apply against a " +
				"balance the lock was protecting")
	}

	operation, vErrs, err := prepareCorrectionTransfer(ctx, request)
	if err != nil || vErrs.Count() > 0 {
		return nil, vErrs, err
	}

	moveId, err := insertCorrectionMove(ctx, operation, request)
	if err != nil {
		return nil, nil, err
	}
	if err := insertCorrectionMoveLine(ctx, operation, request, moveId); err != nil {
		return nil, nil, err
	}

	if err := reloadCorrectionMoves(ctx, operation); err != nil {
		return nil, nil, err
	}
	if err := runCorrectionToDone(ctx, operation); err != nil {
		return nil, nil, err
	}

	return &CorrectionResult{
		TransferId: derefString(operation.Transfer.GetId()),
		MoveId:     moveId,
	}, ft.NewClientErrors(), nil
}

// assertCorrectable applies the rules the schema cannot express.
func assertCorrectable(request CorrectionRequest) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	if request.Quantity.LessThanOrEqual(decimal.Zero) {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockMoveSchemaName,
			"stock_correction.quantity_not_positive",
			"a correction must move a quantity greater than zero"))
	}
	if request.SourceLocationId == request.DestinationLocationId {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockMoveSchemaName,
			"stock_correction.same_location",
			"a correction must move between two different locations"))
	}
	return vErrs
}

// prepareCorrectionTransfer resolves the internal operation type and inserts the transfer header.
// It writes through the repository rather than StockTransferDomainServiceImpl.Create, which would
// open a nested transaction; the policy snapshots that path would copy are written here instead.
func prepareCorrectionTransfer(
	ctx corectx.Context, request CorrectionRequest,
) (*transferOperationContext, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	operationType, err := findCorrectionOperationType(ctx, request.OrgId)
	if err != nil {
		return nil, vErrs, err
	}
	if operationType == nil {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockOperationTypeSchemaName,
			"stock_correction.no_operation_type",
			"this organisation has no internal operation type with code '"+
				models.StockCorrectionOperationTypeCode+"'; the correction seeds must be applied"))
		return nil, vErrs, nil
	}

	operation, err := resolveStockEngines()
	if err != nil {
		return nil, vErrs, err
	}

	transferId, err := insertCorrectionTransfer(ctx, operation, request, *operationType)
	if err != nil {
		return nil, vErrs, err
	}

	loaded, err := loadTransferOperation(ctx, transferId)
	if err != nil {
		return nil, vErrs, err
	}
	if loaded == nil {
		return nil, vErrs, errors.New("the correction transfer vanished immediately after insert")
	}
	return loaded, vErrs, nil
}

// insertCorrectionTransfer writes the header and reads back its id by (org_id, transfer_number),
// since the repository's Insert does not return the generated id and that pair is unique.
func insertCorrectionTransfer(
	ctx corectx.Context,
	operation *transferOperationContext,
	request CorrectionRequest,
	operationType models.StockOperationType,
) (string, error) {
	transferNumber, err := generateTransferNumber(correctionOperationCode)
	if err != nil {
		return "", err
	}

	_, err = operation.TransferEngine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
		models.StockTransferFieldTransferNumber:        transferNumber,
		models.StockTransferFieldOperationTypeId:       derefString(operationType.GetId()),
		models.StockTransferFieldOperationCode:         correctionOperationCode,
		models.StockTransferFieldSourceLocationId:      request.SourceLocationId,
		models.StockTransferFieldDestinationLocationId: request.DestinationLocationId,
		models.StockTransferFieldStatus:                models.StockTransferStatusDraft,
		models.StockTransferFieldReservationMethod:     derefString(operationType.GetReservationMethod()),
		// Never `always` or `ask`, whatever the operation type says: a correction has no remainder to
		// carry forward, so a shortfall is an error rather than a second document.
		models.StockTransferFieldBackorderPolicy: models.StockBackorderPolicyNever,
		models.StockTransferFieldShippingPolicy:  derefString(operationType.GetShippingPolicy()),
		models.StockTransferFieldOriginReference: request.OriginReference,
		models.StockTransferFieldNote:            request.Note,
		models.StockTransferFieldOrgId:           request.OrgId,
	})
	if err != nil {
		return "", errors.Wrap(err, "insertCorrectionTransfer")
	}

	return findTransferByNumber(ctx, operation, request.OrgId, transferNumber)
}

// insertCorrectionMove writes the single move the correction consists of.
func insertCorrectionMove(
	ctx corectx.Context, operation *transferOperationContext, request CorrectionRequest,
) (string, error) {
	quantity := request.Quantity.String()
	_, err := operation.MoveEngine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
		models.StockMoveFieldTransferId:            derefString(operation.Transfer.GetId()),
		models.StockMoveFieldProductVariantId:      request.ProductVariantId,
		models.StockMoveFieldDemandQuantity:        quantity,
		models.StockMoveFieldBaseDemandQuantity:    quantity,
		models.StockMoveFieldSourceLocationId:      request.SourceLocationId,
		models.StockMoveFieldDestinationLocationId: request.DestinationLocationId,
		models.StockMoveFieldStatus:                models.StockMoveStatusDraft,
		models.StockMoveFieldIsInventoryAdjustment: request.IsInventoryAdjustment,
		models.StockMoveFieldOrgId:                 request.OrgId,
	})
	if err != nil {
		return "", errors.Wrap(err, "insertCorrectionMove")
	}

	moves, err := models.FindTransferMoves(
		ctx, operation.MoveEngine.ResourceRepository(),
		derefString(operation.Transfer.GetId()), models.MaxTransferMoves)
	if err != nil {
		return "", err
	}
	if len(moves) == 0 {
		return "", errors.New("the correction move vanished immediately after insert")
	}
	return derefString(models.NewStockMoveFrom(moves[0]).GetId()), nil
}

// insertCorrectionMoveLine writes the execution line the correction's move will ship.
//
// Validate ships move lines, not moves, and a correction is an internal transfer that is never
// reserved, so no other path creates one and the move would execute zero. The line must carry the
// same lot/package/owner dimension the caller locked, or shipOneLine decrements a different balance
// than the one under the lock. It writes through the repository because the move line engine
// refuses client writes.
func insertCorrectionMoveLine(
	ctx corectx.Context,
	operation *transferOperationContext,
	request CorrectionRequest,
	moveId string,
) error {
	// The source balance must exist before shipOneLine can decrement it; for a correction that gains
	// stock the source is the inventory-loss location, which has never been counted. Creating it at
	// zero lets the delta take it negative, which is how a virtual counterparty records what it
	// supplied.
	if _, err := ensureQuantForDimension(ctx, operation, models.QuantDimension{
		OrgId:            request.OrgId,
		ProductVariantId: request.ProductVariantId,
		LocationId:       request.SourceLocationId,
		LotRef:           request.LotRef,
		PackageRef:       request.PackageRef,
		OwnerRef:         request.OwnerRef,
	}); err != nil {
		return err
	}

	quantity := request.Quantity.String()
	_, err := operation.MoveLineEngine.ResourceRepository().Insert(ctx, dmodel.DynamicFields{
		models.StockMoveLineFieldMoveId:                moveId,
		models.StockMoveLineFieldTransferId:            derefString(operation.Transfer.GetId()),
		models.StockMoveLineFieldProductVariantId:      request.ProductVariantId,
		models.StockMoveLineFieldQuantity:              quantity,
		models.StockMoveLineFieldBaseQuantity:          quantity,
		models.StockMoveLineFieldSourceLocationId:      request.SourceLocationId,
		models.StockMoveLineFieldDestinationLocationId: request.DestinationLocationId,
		models.StockMoveLineFieldLotRef:                request.LotRef,
		models.StockMoveLineFieldPackageRef:            request.PackageRef,
		models.StockMoveLineFieldResultPackageRef:      request.PackageRef,
		models.StockMoveLineFieldOwnerRef:              request.OwnerRef,
		models.StockMoveLineFieldOrgId:                 request.OrgId,
	})
	return errors.Wrap(err, "insertCorrectionMoveLine")
}

// reloadCorrectionMoves refreshes the operation's move list after the insert. loadTransferOperation
// ran before the move existed, so without this executeMoves finds nothing and the correction closes
// having moved no stock.
func reloadCorrectionMoves(ctx corectx.Context, operation *transferOperationContext) error {
	moves, err := models.FindTransferMoves(
		ctx, operation.MoveEngine.ResourceRepository(),
		derefString(operation.Transfer.GetId()), models.MaxTransferMoves)
	if err != nil {
		return err
	}
	operation.Moves = moves
	return nil
}

// runCorrectionToDone confirms and validates the generated transfer in one go, so a variance never
// sits half-applied. It reuses executeMoves and closeTransfer because Validate would open a nested
// transaction.
func runCorrectionToDone(ctx corectx.Context, operation *transferOperationContext) error {
	if err := confirmMoves(ctx, operation); err != nil {
		return err
	}
	if err := reloadCorrectionMoves(ctx, operation); err != nil {
		return err
	}

	outcomes, err := executeMoves(ctx, operation)
	if err != nil {
		return err
	}
	if err := assertCorrectionComplete(outcomes); err != nil {
		return err
	}
	return closeTransfer(ctx, operation, "")
}

// assertCorrectionComplete refuses a correction that could only be partly applied: with backorder
// policy `never` a shortfall would be silently dropped, leaving the balance between the old value
// and the counted one. Failing rolls the whole correction back.
func assertCorrectionComplete(outcomes []moveOutcome) error {
	for _, outcome := range outcomes {
		if outcome.Shortfall().GreaterThan(decimal.Zero) {
			return errors.Errorf(
				"the correction could only move %s of %s; the source balance changed under the lock",
				outcome.Processed.String(), outcome.Demand.String())
		}
	}
	return nil
}

// FindLocationByType resolves an org's virtual counterparty location of a given type. Inventory
// adjustments balance against `inventory_loss` and scraps move to `scrap`, both seeded per org.
// The first match wins; two matching locations is a configuration problem, not a reason to refuse.
func FindLocationByType(
	ctx corectx.Context, orgId string, locationType string,
) (*models.InventoryLocation, error) {
	engine, err := engineFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldLocationUsage, dmodel.Equals, locationType),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "FindLocationByType")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewInventoryLocationFrom(found.Data.Items[0]), nil
}

// findCorrectionOperationType resolves the seeded internal operation type corrections run through.
func findCorrectionOperationType(
	ctx corectx.Context, orgId string,
) (*models.StockOperationType, error) {
	engine, err := engineFor(models.StockOperationTypeSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockOperationTypeFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(
			models.StockOperationTypeFieldCode, dmodel.Equals, models.StockCorrectionOperationTypeCode),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "findCorrectionOperationType")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewStockOperationTypeFrom(found.Data.Items[0]), nil
}
