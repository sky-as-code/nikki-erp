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

// The one path by which a correction changes a balance.
//
// An inventory adjustment and a scrap both need the same thing: move a quantity between an
// internal location and a virtual counterparty, atomically, and end up `done`. Decision F3 says
// they do it by generating a lightweight internal transfer and running it through Phase 2's
// validate path, rather than writing a quant directly.
//
// The reason is that applyQuantDelta in stock_validate.go is the only function in the module that
// writes on_hand_quantity, and it is reachable only from validate. A second writer would be a
// second implementation of the most dangerous code here, and the two would drift. The cost of this
// approach is one transfer row per correction, which is cheap next to that risk — and it has a
// benefit the direct path would not: an adjustment or a scrap shows up in the Transfers list with
// a recognisable operation type, which is exactly what an auditor looks for.

// correctionOperationCode is the operation code of the generated transfer. A correction moves
// between two of the company's own locations — an internal one and a virtual counterparty — so it
// is internal in the sense the direction field means (BR §4.2.1.2).
const correctionOperationCode = models.StockOperationCodeInternal

// CorrectionRequest describes a balance correction to apply.
//
// It names both locations explicitly rather than deriving them from a direction flag: an
// adjustment gaining stock and one losing it differ only in which side is the inventory-loss
// location, and making the caller state both keeps the sign where the caller can see it. Getting
// this backwards produces a plausible-looking movement with the sign inverted, so it is pinned by
// test rather than by comment.
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

	// IsInventoryAdjustment marks the generated move as an adjustment rather than an ordinary
	// movement, which is what separates a count correction from a scrap in movement history.
	IsInventoryAdjustment bool
}

// CorrectionResult reports what the correction generated, so the caller can record the link back.
type CorrectionResult struct {
	TransferId string
	MoveId     string
}

// ApplyCorrectionMovement generates a done internal transfer for one quantity of one variant.
//
// It must run inside the caller's transaction and never opens its own: pgTxClient.BeginTx returns
// ErrTxNested, and both callers (apply-adjustment and do-scrap) already hold a transaction with
// the source quant locked FOR UPDATE. Passing a ctx without one is a programming error, not a
// client error, so it fails loudly.
//
// The sequence is create → confirm-equivalent → validate, all through the Phase 2 machinery:
// executeMoves does the balance arithmetic and closeTransfer marks the document done.
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
//
// The transfer goes through the repository rather than through StockTransferDomainServiceImpl.Create,
// because that method opens its own transaction and this one already runs inside the caller's.
// The policy snapshots the create path would copy are written here explicitly instead, which is
// also where the backorder policy is forced.
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

// insertCorrectionTransfer writes the header and reads back its id.
//
// The read-back by (org_id, transfer_number) is the same trick createBackorderTransfer uses: the
// repository's Insert does not return the generated id, and the composite unique makes the pair a
// safe key to find it by.
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
		// Never `always` or `ask`, whatever the operation type says. A correction has no remainder
		// to carry forward: if the variance could not be applied in full that is an error to
		// surface, not a second document to chase.
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
// This is the piece that makes a correction work at all, and the reason is worth stating: validate
// ships move *lines*, not moves. Reservation creates them for ordinary internal movements, and
// ensureIncomingLine creates one for incoming transfers whose source holds no balance. A correction
// is neither — it is an `internal` transfer that is never reserved, so both of those paths decline
// to act and the move would execute zero. `assertCorrectionComplete` would then roll the whole
// thing back, which is a safe failure but not a working feature.
//
// Writing the line here rather than teaching reservation about corrections keeps the special case
// where its cause is. It also carries the lot/package/owner dimension straight through, which
// matters: the caller has already locked exactly one quant, and the line must name that same
// dimension or shipOneLine would decrement a different balance than the one under the lock.
//
// It goes through the repository rather than the service because the move line's engine refuses
// client writes (see defineStockMoveLineActions) — the same reason reservation writes lines
// directly.
func insertCorrectionMoveLine(
	ctx corectx.Context,
	operation *transferOperationContext,
	request CorrectionRequest,
	moveId string,
) error {
	// The source balance must exist before shipOneLine can decrement it, and for a correction that
	// *gains* stock it will not: the source is the inventory-loss location, which no one has ever
	// counted. Creating it at zero lets the delta take it negative, which is exactly how a virtual
	// counterparty records what it has supplied — the same reasoning as ensureIncomingLine.
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

// reloadCorrectionMoves refreshes the operation's move list after the insert.
//
// loadTransferOperation ran before the move existed, so its snapshot is empty; executeMoves would
// otherwise find nothing to do and the correction would silently close having moved no stock.
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

// runCorrectionToDone confirms and validates the generated transfer in one go.
//
// Per decision F4 a correction is not left for a user to validate: the count or the scrap *is* the
// decision, and a draft correction awaiting confirm would let a variance sit half-applied. It
// reuses executeMoves and closeTransfer rather than calling Validate, which would try to open a
// nested transaction.
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

// assertCorrectionComplete refuses a correction that could only be partly applied.
//
// With backorder policy `never` a shortfall would otherwise be silently dropped, leaving the
// balance somewhere between what it was and what the count said. Failing the transaction rolls the
// whole correction back, which is the honest outcome: the caller re-reads and tries again.
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

// FindLocationByType resolves an org's virtual counterparty location of a given type.
//
// Inventory adjustments balance against `inventory_loss` and scraps move to `scrap` (BR §4.2.7.2,
// §4.2.9.1). Both are seeded per org. The first match wins: an org with two scrap locations has a
// configuration problem, and picking one deterministically beats refusing to scrap at all.
func FindLocationByType(
	ctx corectx.Context, orgId string, locationType string,
) (*models.StockLocation, error) {
	engine, err := engineFor(models.StockLocationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.StockLocationFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(
			models.StockLocationFieldLocationType, dmodel.Equals, locationType),
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
	return models.NewStockLocationFrom(found.Data.Items[0]), nil
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
