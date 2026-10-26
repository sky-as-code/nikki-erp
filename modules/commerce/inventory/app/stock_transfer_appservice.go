package app

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Param names the movement actions read from the request.
const (
	paramIdempotencyKey  = "idempotency_key"
	paramCreateBackorder = "create_backorder"
	paramSourceType      = "source_type"
	paramSourceId        = "source_id"
	paramSourceItems     = "items"
	paramHoldLocationId  = "location_id"
	paramToLocationId    = "to_location_id"
	paramOperationTypeId = "operation_type_id"
	paramOriginReference = "origin_reference"
	paramEventId         = "event_id"
	paramExecutorLocId   = "executor_location_id"
	paramSuccessfulItems = "successful_items"
	paramFailedItems     = "failed_items"
)

// NewStockTransferApplicationService is handed the composable default by the transfer onion.
func NewStockTransferApplicationService(base composable.CrudApplicationService) itStock.StockTransferApplicationService {
	transferSvc, ok := base.DomainService().(*services.StockTransferDomainServiceImpl)
	if !ok {
		panic(errors.New("the stock transfer onion must be built with NewStockTransferDomainService"))
	}
	return &StockTransferApplicationServiceImpl{CrudApplicationService: base, transferSvc: transferSvc}
}

// StockTransferApplicationServiceImpl exposes the movement operations. Each is a POST because
// none is a CRUD verb: a validate is not an update to a transfer, it is an event that happens to
// one.
type StockTransferApplicationServiceImpl struct {
	composable.CrudApplicationService
	transferSvc *services.StockTransferDomainServiceImpl
}

func (this *StockTransferApplicationServiceImpl) Confirm(
	ctx corectx.Context, cmd itStock.ConfirmTransferCommand,
) (*itStock.TransferMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionConfirm, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.transferSvc.Confirm(ctx, readStringField(cmd, paramRecordId))
}

// CheckAvailability takes nothing and changes nothing, so the read permission covers it.
func (this *StockTransferApplicationServiceImpl) CheckAvailability(
	ctx corectx.Context, query itStock.CheckAvailabilityQuery,
) (*itStock.TransferReadResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionRead, query); err != nil || cErrs != nil {
		return anyFailure(cErrs, err)
	}
	return this.transferSvc.CheckAvailability(ctx, readStringField(query, paramRecordId))
}

func (this *StockTransferApplicationServiceImpl) Reserve(
	ctx corectx.Context, cmd itStock.ReserveTransferCommand,
) (*itStock.TransferMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionReserve, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.transferSvc.Reserve(ctx, readStringField(cmd, paramRecordId))
}

func (this *StockTransferApplicationServiceImpl) Unreserve(
	ctx corectx.Context, cmd itStock.UnreserveTransferCommand,
) (*itStock.TransferMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionUnreserve, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.transferSvc.Unreserve(ctx, readStringField(cmd, paramRecordId))
}

func (this *StockTransferApplicationServiceImpl) Validate(
	ctx corectx.Context, cmd itStock.ValidateTransferCommand,
) (*itStock.TransferMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionValidate, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.transferSvc.Validate(ctx, readStringField(cmd, paramRecordId),
		readStringField(cmd, paramIdempotencyKey), readOptionalBool(cmd, paramCreateBackorder))
}

func (this *StockTransferApplicationServiceImpl) Cancel(
	ctx corectx.Context, cmd itStock.CancelTransferCommand,
) (*itStock.TransferMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionCancel, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.transferSvc.Cancel(ctx, readStringField(cmd, paramRecordId))
}

func (this *StockTransferApplicationServiceImpl) CreateReturn(
	ctx corectx.Context, cmd itStock.CreateReturnCommand,
) (*itStock.TransferMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionCreateReturn, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	request, vErrs := readReturnRequest(cmd)
	if vErrs.Count() > 0 {
		return mutateFailure(vErrs, nil)
	}
	return this.transferSvc.CreateReturn(ctx, readStringField(cmd, paramRecordId), request)
}

// The demand-addressed operations, for a caller holding its own reference rather than a transfer
// id. Holding and releasing stock are the same powers the id-addressed actions carry, so they
// reuse those permissions rather than inventing parallel ones.

func (this *StockTransferApplicationServiceImpl) ReserveForSource(
	ctx corectx.Context, cmd itStock.ReserveForSourceCommand,
) (*itStock.TransferReadResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionReserve, cmd); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	items, vErrs := readSourceItems(cmd, paramSourceItems)
	if vErrs.Count() > 0 {
		return anyFailure(vErrs, nil)
	}
	return toSourceReservationResult(this.transferSvc.ReserveForSource(ctx, itStock.SourceReservationRequest{
		SourceType:      readStringField(cmd, paramSourceType),
		SourceId:        readStringField(cmd, paramSourceId),
		OrgId:           readStringField(cmd, models.StockTransferFieldOrgId),
		LocationId:      readStringField(cmd, paramHoldLocationId),
		OperationTypeId: readStringField(cmd, paramOperationTypeId),
		OriginReference: readStringField(cmd, paramOriginReference),
		Items:           items,
	}))
}

func (this *StockTransferApplicationServiceImpl) ReleaseForSource(
	ctx corectx.Context, cmd itStock.ReleaseForSourceCommand,
) (*itStock.TransferMutateResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionUnreserve, cmd); cErrs != nil {
		return mutateFailure(cErrs, nil)
	}
	return this.transferSvc.ReleaseReservationBySource(ctx,
		readStringField(cmd, paramSourceType), readStringField(cmd, paramSourceId))
}

func (this *StockTransferApplicationServiceImpl) ReallocateReservation(
	ctx corectx.Context, cmd itStock.ReallocateReservationCommand,
) (*itStock.TransferReadResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionReserve, cmd); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	items, vErrs := readSourceItems(cmd, paramSourceItems)
	if vErrs.Count() > 0 {
		return anyFailure(vErrs, nil)
	}
	return toSourceReservationResult(this.transferSvc.ReallocateReservation(ctx, itStock.ReservationReallocationRequest{
		SourceType:      readStringField(cmd, paramSourceType),
		SourceId:        readStringField(cmd, paramSourceId),
		OrgId:           readStringField(cmd, models.StockTransferFieldOrgId),
		ToLocationId:    readStringField(cmd, paramToLocationId),
		OperationTypeId: readStringField(cmd, paramOperationTypeId),
		OriginReference: readStringField(cmd, paramOriginReference),
		Items:           items,
	}))
}

// ApplyFulfillmentResult is a machine reporting what it managed to hand over, which is its own
// power, distinct from an operator validating a document.
func (this *StockTransferApplicationServiceImpl) ApplyFulfillmentResult(
	ctx corectx.Context, cmd itStock.ApplyFulfillmentResultCommand,
) (*itStock.TransferReadResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionApplyFulfillmentResult, cmd); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	successful, vErrs := readResultItems(cmd, paramSuccessfulItems)
	failed, failedErrs := readResultItems(cmd, paramFailedItems)
	vErrs.ConcatPtr(failedErrs)
	if vErrs.Count() > 0 {
		return anyFailure(vErrs, nil)
	}

	result, err := this.transferSvc.ApplyFulfillmentResult(ctx, itStock.FulfillmentResultRequest{
		EventId:            readStringField(cmd, paramEventId),
		SourceType:         readStringField(cmd, paramSourceType),
		SourceId:           readStringField(cmd, paramSourceId),
		OrgId:              readStringField(cmd, models.StockTransferFieldOrgId),
		ExecutorLocationId: readStringField(cmd, paramExecutorLocId),
		SuccessfulItems:    successful,
		FailedItems:        failed,
	})
	if err != nil {
		return nil, err
	}
	return &itStock.TransferReadResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.ClientErrors.Count() == 0,
		Data:         result,
	}, nil
}

// toSourceReservationResult shapes a hold's outcome. A refusal travels as ClientErrors with no Go
// error, which is what makes the answer 400 rather than 500: a location that cannot supply the
// goods is something the caller fixes by choosing another.
func toSourceReservationResult(
	result *itStock.SourceReservationResult, err error,
) (*itStock.TransferReadResult, error) {
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[any]{
		ClientErrors: result.ClientErrors,
		HasData:      result.ClientErrors.Count() == 0,
		Data:         result,
	}, nil
}
