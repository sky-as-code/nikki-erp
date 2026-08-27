package services

import (
	"strings"
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockTransferDomainService derives the transfer service from the engine's default one.
//
// base is the Stock Transfer engine's own resource service, which this type embeds: built-in CRUD
// keeps running through the default implementation, and the movement operations are layered on
// top. The result is installed with Engine.SetResourceService.
func NewStockTransferDomainService(base drif.DynamicResourceService) *StockTransferDomainServiceImpl {
	return &StockTransferDomainServiceImpl{DynamicResourceService: base}
}

// StockTransferDomainServiceImpl adds the movement operations to the transfer resource.
//
// Every one of them is a transaction over several resources — the transfer, its moves, their move
// lines and the balances on both sides — so they live on the service rather than in an engine
// callback: a dynamicengines callback may adapt and validate, but the writes belong here. See
// docs/wiki/07 §6.7.
type StockTransferDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*StockTransferDomainServiceImpl)(nil)

// The published goods-movement port (SALES-049). Asserted here so that changing one of the six
// operations' signatures breaks the build in this file, naming the contract it broke — rather than
// at the deps.Register in index.go, or, worse, at a consumer in another module that can no longer
// bind. The port is the promise; this line is what keeps the promise checked.
var _ itStock.StockTransferMovementService = (*StockTransferDomainServiceImpl)(nil)

// transferOperationContext carries what every movement operation needs: the engines it writes
// through, and the transfer it is acting on. Assembling it in one place keeps each operation's
// body about its own rule rather than about lookups.
type transferOperationContext struct {
	TransferEngine drif.DynamicResourceEngine
	MoveEngine     drif.DynamicResourceEngine
	MoveLineEngine drif.DynamicResourceEngine
	QuantEngine    drif.DynamicResourceEngine

	Transfer models.StockTransfer
	Moves    []dmodel.DynamicFields
}

// loadTransferOperation resolves the engines and reads the transfer with its moves.
//
// The read happens inside the caller's transaction when there is one, because ctx carries it: the
// same call is therefore safe both for the read-only availability check and for the operations
// that go on to write.
func loadTransferOperation(
	ctx corectx.Context, transferId string,
) (*transferOperationContext, error) {
	engines, err := resolveStockEngines()
	if err != nil {
		return nil, err
	}

	found, err := engines.TransferEngine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.StockTransferFieldId: transferId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadTransferOperation")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	moves, err := models.FindTransferMoves(
		ctx, engines.MoveEngine.ResourceRepository(), transferId, models.MaxTransferMoves)
	if err != nil {
		return nil, err
	}

	engines.Transfer = *models.NewStockTransferFrom(found.Data)
	engines.Moves = moves
	return engines, nil
}

func resolveStockEngines() (*transferOperationContext, error) {
	transferEngine, err := engineFor(models.StockTransferSchemaName)
	if err != nil {
		return nil, err
	}
	moveEngine, err := engineFor(models.StockMoveSchemaName)
	if err != nil {
		return nil, err
	}
	moveLineEngine, err := engineFor(models.StockMoveLineSchemaName)
	if err != nil {
		return nil, err
	}
	quantEngine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return nil, err
	}
	return &transferOperationContext{
		TransferEngine: transferEngine,
		MoveEngine:     moveEngine,
		MoveLineEngine: moveLineEngine,
		QuantEngine:    quantEngine,
	}, nil
}

// withTransferTransaction runs body inside one transaction on a scoped copy of the context.
//
// The transaction goes on a clone, never on ctx itself: setting it on the caller's context would
// leave a committed transaction visible to whatever runs next. CloneRequestContext carries the
// caller's identity across, which the audit columns need. See docs/wiki/02 §5.1.
//
// There is no "join an existing transaction" branch, because pgTxClient.BeginTx returns
// ErrTxNested: these operations are entry points, and nesting one inside another is a bug rather
// than a case to handle.
func withTransferTransaction(
	ctx corectx.Context, body func(tranxCtx corectx.Context) error,
) error {
	engine, err := engineFor(models.StockTransferSchemaName)
	if err != nil {
		return err
	}

	tranx, err := engine.ResourceRepository().BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "withTransferTransaction")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrap(tranx.Commit(), "withTransferTransaction")
}

// notFoundResult is the answer when an operation names a transfer that does not exist. It is a
// client error rather than a server one: the id came from the caller.
func notFoundResult(transferId string) *dyn.OpResult[dyn.MutateResultData] {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.StockTransferSchemaName,
		"stock_transfer.not_found",
		"no stock transfer with id '"+transferId+"'",
	))
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}

func violationResult(key, message string) *dyn.OpResult[dyn.MutateResultData] {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName, key, message))
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}

// mutateOk is the success envelope of a movement operation.
//
// AffectedCount is 1 for the transfer itself, not a count of the moves, lines and balances the
// operation touched: the caller asked to act on one transfer, and reporting the internal write
// count would make the number mean something different for each operation.
func mutateOk() *dyn.OpResult[dyn.MutateResultData] {
	return &dyn.OpResult[dyn.MutateResultData]{
		Data:    dyn.MutateResultData{AffectedCount: 1},
		HasData: true,
	}
}

// updateTransferStatus writes a new transfer state, refusing a transition the machine forbids.
//
// The guard is here rather than only in the callers because every path that changes a state goes
// through it: an illegal transition should be impossible to write, not merely unlikely.
func updateTransferStatus(
	ctx corectx.Context, engine drif.DynamicResourceEngine, transfer models.StockTransfer, next string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	current := derefString(transfer.GetStatus())
	if current == next {
		return nil, nil
	}

	vErrs := ft.NewClientErrors()
	AssertTransferTransition(current, next, vErrs)
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	update := dmodel.DynamicFields{
		models.StockTransferFieldId:     derefString(transfer.GetId()),
		models.StockTransferFieldStatus: next,
		basemodel.FieldEtag:             derefString(transfer.GetEtag()),
	}
	if next == models.StockTransferStatusDone {
		update[models.StockTransferFieldCompletedAt] = time.Now().UTC()
	}

	_, err := engine.ResourceRepository().Update(ctx, update)
	return nil, errors.Wrap(err, "updateTransferStatus")
}

// updateMoveStatus writes a new move state, refusing a transition the machine forbids.
func updateMoveStatus(
	ctx corectx.Context, engine drif.DynamicResourceEngine, move models.StockMove, next string,
) error {
	current := derefString(move.GetStatus())
	if current == next {
		return nil
	}
	if !CanTransitionMove(current, next) {
		return errors.Errorf("a stock move cannot go from '%s' to '%s'", current, next)
	}

	_, err := engine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.StockMoveFieldId:     derefString(move.GetId()),
		models.StockMoveFieldStatus: next,
		basemodel.FieldEtag:         derefString(move.GetEtag()),
	})
	return errors.Wrap(err, "updateMoveStatus")
}

// moveStatuses reads the state of each move, for summarising into the transfer's own.
func moveStatuses(moves []dmodel.DynamicFields) []string {
	statuses := make([]string, 0, len(moves))
	for _, item := range moves {
		move := models.NewStockMoveFrom(item)
		statuses = append(statuses, derefString(move.GetStatus()))
	}
	return statuses
}

// generateTransferNumber builds the document number a transfer is known by.
//
// It is derived from the operation type's code and a fresh ULID rather than from a counter,
// because a counter needs a sequence table and a lock of its own, and would serialise every
// create in an org. The ULID's leading characters are time-ordered, so numbers still sort roughly
// by creation. The uniqueness that matters is enforced by the composite unique on
// (transfer_number, org_id), not by this function being clever.
func generateTransferNumber(operationCode string) (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", errors.Wrap(err, "generateTransferNumber")
	}
	prefix := strings.ToUpper(operationCode)
	if prefix == "" {
		prefix = "TRF"
	}
	return prefix[:min(3, len(prefix))] + "-" + *id, nil
}
