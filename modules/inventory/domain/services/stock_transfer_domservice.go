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

// NewStockTransferDomainService derives the transfer service from the engine's default one, which
// it embeds so built-in CRUD keeps running unchanged. Installed with Engine.SetResourceService.
func NewStockTransferDomainService(base drif.DynamicResourceService) *StockTransferDomainServiceImpl {
	return &StockTransferDomainServiceImpl{DynamicResourceService: base}
}

// StockTransferDomainServiceImpl adds the movement operations to the transfer resource. Each is a
// transaction over the transfer, its moves, their lines and the balances on both sides, so they
// live on the service rather than in an engine callback, which may only adapt and validate.
type StockTransferDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*StockTransferDomainServiceImpl)(nil)

// The published goods-movement port, asserted here so changing an operation's signature breaks the
// build in this file rather than at deps.Register or in a consuming module.
var _ itStock.StockTransferMovementService = (*StockTransferDomainServiceImpl)(nil)

// transferOperationContext carries what every movement operation needs: the engines it writes
// through and the transfer it is acting on.
type transferOperationContext struct {
	TransferEngine drif.DynamicResourceEngine
	MoveEngine     drif.DynamicResourceEngine
	MoveLineEngine drif.DynamicResourceEngine
	QuantEngine    drif.DynamicResourceEngine

	Transfer models.StockTransfer
	Moves    []dmodel.DynamicFields
}

// loadTransferOperation resolves the engines and reads the transfer with its moves. The read joins
// the caller's transaction when ctx carries one, so it is safe both for the read-only availability
// check and for the operations that go on to write.
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

// withTransferTransaction runs body inside one transaction on a cloned context. The transaction
// must go on a clone, never on ctx itself, or a committed transaction stays visible to whatever
// runs next; CloneRequestContext carries the caller's identity across for the audit columns.
//
// There is no join-an-existing branch: BeginTx returns ErrTxNested, so nesting is a bug.
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

// notFoundResult is the answer when an operation names a transfer that does not exist. A client
// error, not a server one: the id came from the caller.
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

// mutateOk is the success envelope of a movement operation. AffectedCount is 1 for the transfer
// itself, never a count of the moves, lines and balances touched, which would mean something
// different for each operation.
func mutateOk() *dyn.OpResult[dyn.MutateResultData] {
	return &dyn.OpResult[dyn.MutateResultData]{
		Data:    dyn.MutateResultData{AffectedCount: 1},
		HasData: true,
	}
}

// updateTransferStatus writes a new transfer state, refusing a transition the machine forbids.
// The guard sits here because every state-changing path goes through it, making an illegal
// transition impossible to write rather than merely unlikely.
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

// generateTransferNumber builds the document number from the operation type's code and a fresh
// ULID rather than a counter, which would need a sequence table and serialise every create in an
// org. ULID prefixes are time-ordered, so numbers still sort roughly by creation. Uniqueness is
// enforced by the composite unique on (transfer_number, org_id), not by this function.
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
