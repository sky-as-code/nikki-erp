package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Create, Confirm and Cancel move a transfer through its lifecycle without touching a balance.
// Reserve and Validate, which do touch balances, are in their own files.

// Create overrides the built-in create to stamp the fields a client may not choose: the transfer
// number is generated so two transfers cannot collide or impersonate each other; the operation
// type's three policies are copied as snapshots, so reconfiguring the type later cannot
// reinterpret an existing transfer; and the status is forced to draft, since a transfer created
// `done` would be a completed movement with nothing behind it. It creates no stock and reserves
// nothing.
func (this *StockTransferDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	operationTypeId := readStringParam(params, models.StockTransferFieldOperationTypeId)
	if operationTypeId == "" {
		// The field is required by the schema, so the base call reports the omission.
		return this.DynamicResourceService.Create(ctx, params)
	}

	operationType, vErrs, err := loadUsableOperationType(ctx, operationTypeId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}

	prepared, err := prepareTransferForCreate(params, *operationType)
	if err != nil {
		return nil, err
	}
	if clientErrs := assertCreatableTransfer(prepared); clientErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *clientErrs}, nil
	}

	return this.DynamicResourceService.Create(ctx, prepared)
}

// loadUsableOperationType reads the operation type and refuses an archived one. The check must stay
// at create time, not read time: an archived type may not start new business, but a transfer
// created before it was archived must still resolve it or its history becomes unreadable.
func loadUsableOperationType(
	ctx corectx.Context, operationTypeId string,
) (*models.StockOperationType, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	engine, err := engineFor(models.StockOperationTypeSchemaName)
	if err != nil {
		return nil, vErrs, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.StockOperationTypeFieldId: operationTypeId,
	})
	if err != nil {
		return nil, vErrs, errors.Wrap(err, "loadUsableOperationType")
	}
	if found == nil || !found.HasData {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockTransferSchemaName,
			"stock_transfer.operation_type_not_found",
			"no stock operation type with id '"+operationTypeId+"'",
		))
		return nil, vErrs, nil
	}

	operationType := models.NewStockOperationTypeFrom(found.Data)
	if isArchived(found.Data) {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockTransferSchemaName,
			"stock_transfer.operation_type_archived",
			"an archived stock operation type cannot start a new transfer",
		))
		return nil, vErrs, nil
	}
	return operationType, vErrs, nil
}

// prepareTransferForCreate stamps the server-owned fields over whatever the client sent. Client
// values are overwritten rather than rejected, so a client echoing a record back does not fail.
func prepareTransferForCreate(
	params dmodel.DynamicFields, operationType models.StockOperationType,
) (dmodel.DynamicFields, error) {
	operationCode := derefString(operationType.GetOperationCode())

	transferNumber, err := generateTransferNumber(operationCode)
	if err != nil {
		return nil, err
	}

	prepared := make(dmodel.DynamicFields, len(params)+6)
	for key, value := range params {
		prepared[key] = value
	}

	prepared[models.StockTransferFieldTransferNumber] = transferNumber
	prepared[models.StockTransferFieldStatus] = models.StockTransferStatusDraft
	prepared[models.StockTransferFieldOperationCode] = operationCode
	prepared[models.StockTransferFieldReservationMethod] = derefString(operationType.GetReservationMethod())
	prepared[models.StockTransferFieldBackorderPolicy] = derefString(operationType.GetBackorderPolicy())
	prepared[models.StockTransferFieldShippingPolicy] = derefString(operationType.GetShippingPolicy())

	// Defaults from the operation type, applied only where the client named nothing.
	applyDefaultLocation(prepared, models.StockTransferFieldSourceLocationId,
		operationType.GetDefaultSourceLocationId())
	applyDefaultLocation(prepared, models.StockTransferFieldDestinationLocationId,
		operationType.GetDefaultDestinationLocationId())

	return prepared, nil
}

func applyDefaultLocation(params dmodel.DynamicFields, field string, fallback *string) {
	if readStringParam(params, field) != "" || fallback == nil || *fallback == "" {
		return
	}
	params[field] = *fallback
}

// assertCreatableTransfer applies the rules the schema cannot express.
func assertCreatableTransfer(params dmodel.DynamicFields) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()

	source := readStringParam(params, models.StockTransferFieldSourceLocationId)
	destination := readStringParam(params, models.StockTransferFieldDestinationLocationId)
	if source != "" && source == destination {
		// A transfer from a location to itself moves nothing, but would still generate moves and
		// consume reservations against the balance it is about to put back.
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockTransferSchemaName,
			"stock_transfer.same_source_and_destination",
			"a transfer's source and destination locations must differ",
		))
	}
	return vErrs
}

func isArchived(fields dmodel.DynamicFields) bool {
	value, ok := fields[basemodel.FieldIsArchived]
	if !ok || value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case *bool:
		return typed != nil && *typed
	default:
		return false
	}
}

// Confirm takes a transfer out of draft and, when its snapshot policy says so, reserves for it.
// Confirming commits to the demand, not to particular stock, and changes no on-hand quantity.
// `at_confirmation` reserves here, `manual` waits to be asked, `before_scheduled_date` waits for
// the scheduler.
func (this *StockTransferDomainServiceImpl) Confirm(
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

		status := derefString(operation.Transfer.GetStatus())
		if status != models.StockTransferStatusDraft {
			result = violationResult(
				"stock_transfer.not_draft",
				"only a draft transfer can be confirmed; this one is '"+status+"'")
			return nil
		}
		if len(operation.Moves) == 0 {
			// Confirming an empty transfer would produce a document that can never become ready and has
			// nothing to validate.
			result = violationResult(
				"stock_transfer.no_moves",
				"a transfer with no moves cannot be confirmed")
			return nil
		}

		if err := confirmMoves(tranxCtx, operation); err != nil {
			return err
		}
		result, err = finishConfirm(tranxCtx, operation, transferId)
		return err
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// confirmMoves advances every open move out of draft.
func confirmMoves(ctx corectx.Context, operation *transferOperationContext) error {
	for _, item := range operation.Moves {
		move := models.NewStockMoveFrom(item)
		if !IsMoveOpen(derefString(move.GetStatus())) {
			continue
		}
		if err := updateMoveStatus(ctx, operation.MoveEngine, *move, models.StockMoveStatusConfirmed); err != nil {
			return err
		}
	}
	return nil
}

// finishConfirm reserves when the snapshot policy asks for it, then writes the transfer's state.
// The state is derived from the moves after any reservation rather than chosen here, so the two
// answers cannot drift apart.
func finishConfirm(
	ctx corectx.Context, operation *transferOperationContext, transferId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if derefString(operation.Transfer.GetReservationMethod()) == models.StockReservationMethodAtConfirmation {
		if _, err := reserveTransferMoves(ctx, operation); err != nil {
			return nil, err
		}
		// Reservation rewrote the move states, so they are re-read rather than reused.
		refreshed, err := models.FindTransferMoves(
			ctx, operation.MoveEngine.ResourceRepository(), transferId, models.MaxTransferMoves)
		if err != nil {
			return nil, err
		}
		operation.Moves = refreshed
	}

	next := DeriveTransferStatus(models.StockTransferStatusConfirmed, moveStatuses(operation.Moves))

	// An incoming transfer is ready as soon as it is confirmed: its goods come from a supplier rather
	// than a balance, so there is nothing to reserve and deriving readiness from allocation would
	// leave every receipt permanently un-ready.
	if derefString(operation.Transfer.GetOperationCode()) == models.StockOperationCodeIncoming &&
		next == models.StockTransferStatusConfirmed {
		next = models.StockTransferStatusReady
	}

	if failed, err := updateTransferStatus(ctx, operation.TransferEngine, operation.Transfer, next); err != nil {
		return nil, err
	} else if failed != nil {
		return failed, nil
	}
	return mutateOk(), nil
}

// Cancel abandons a transfer, releasing whatever it was holding.
//
// Reservations must be released BEFORE the moves are cancelled: a cancelled move is closed, and the
// release pass skips closed moves, so cancelling first strands the reserved quantity on the balance
// with nothing pointing at it.
//
// Cancelling a done transfer is refused; the remedy is a reverse transfer, and the error says so.
func (this *StockTransferDomainServiceImpl) Cancel(
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

		status := derefString(operation.Transfer.GetStatus())
		vErrs := ft.NewClientErrors()
		AssertTransferTransition(status, models.StockTransferStatusCancelled, vErrs)
		if vErrs.Count() > 0 {
			result = &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
			return nil
		}

		if err := unreserveTransferMoves(tranxCtx, operation); err != nil {
			return err
		}
		if err := cancelMoves(tranxCtx, operation); err != nil {
			return err
		}

		failed, err := updateTransferStatus(
			tranxCtx, operation.TransferEngine, operation.Transfer, models.StockTransferStatusCancelled)
		if err != nil {
			return err
		}
		if failed != nil {
			result = failed
			return nil
		}
		result = mutateOk()
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// cancelMoves closes every move that is still open. A move already done stays done: cancelling the
// transfer cannot un-record a movement that happened.
func cancelMoves(ctx corectx.Context, operation *transferOperationContext) error {
	for _, item := range operation.Moves {
		move := models.NewStockMoveFrom(item)
		if !IsMoveOpen(derefString(move.GetStatus())) {
			continue
		}
		if err := updateMoveStatus(ctx, operation.MoveEngine, *move, models.StockMoveStatusCancelled); err != nil {
			return err
		}
	}
	return nil
}
