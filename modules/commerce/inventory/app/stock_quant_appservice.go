package app

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Param names the counting actions read from the request body.
const (
	paramCountedQuantity = "counted_quantity"
	paramCountReasonCode = "count_reason_code"
	paramCountReasonText = "count_reason_text"
	paramNextCountDate   = "next_count_date"
	paramAssignedUserId  = "count_assigned_user_id"
)

// NewStockQuantApplicationService is handed the composable default by the quant onion.
func NewStockQuantApplicationService(base composable.CrudApplicationService) itStock.StockQuantApplicationService {
	quantSvc, ok := base.DomainService().(*services.StockQuantDomainServiceImpl)
	if !ok {
		panic(errors.New("the stock quant onion must be built with NewStockQuantDomainService"))
	}
	return &StockQuantApplicationServiceImpl{CrudApplicationService: base, quantSvc: quantSvc}
}

// StockQuantApplicationServiceImpl closes the quant's write surface and opens the counting and
// product-facing read operations. A balance is the running total of completed movements, not
// something a client sets: create, update and delete are refused so an on-hand quantity never
// appears with no movement behind it. Corrections go through an adjustment, a transfer or a scrap.
//
// The actions are refused rather than removed so a caller gets a 400 naming the reason instead of
// a 404 that reads as "wrong URL".
type StockQuantApplicationServiceImpl struct {
	composable.CrudApplicationService
	quantSvc *services.StockQuantDomainServiceImpl
}

// The counting operations write to a resource whose CRUD is refused above, which is not a
// contradiction: none of them touches on_hand_quantity. Enter and Reset write only count
// metadata; Apply changes the balance solely by generating a movement.

func (this *StockQuantApplicationServiceImpl) EnterCount(
	ctx corectx.Context, cmd itStock.EnterCountCommand,
) (*itStock.CountMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionEnterCount, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	counted, vErrs := readDecimalField(cmd, models.StockQuantSchemaName, paramCountedQuantity)
	if vErrs.Count() > 0 {
		return mutateFailure(vErrs, nil)
	}
	return this.quantSvc.EnterCount(ctx, readStringField(cmd, paramRecordId), counted,
		readStringField(cmd, paramCountReasonCode), readStringField(cmd, paramCountReasonText))
}

func (this *StockQuantApplicationServiceImpl) ResetCount(
	ctx corectx.Context, cmd itStock.ResetCountCommand,
) (*itStock.CountMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionResetCount, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.quantSvc.ResetCount(ctx, readStringField(cmd, paramRecordId))
}

func (this *StockQuantApplicationServiceImpl) ApplyAdjustment(
	ctx corectx.Context, cmd itStock.ApplyAdjustmentCommand,
) (*itStock.CountMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionApplyAdjustment, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.quantSvc.ApplyAdjustment(ctx, readStringField(cmd, paramRecordId))
}

func (this *StockQuantApplicationServiceImpl) ScheduleCount(
	ctx corectx.Context, cmd itStock.ScheduleCountCommand,
) (*itStock.CountMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionScheduleCount, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.quantSvc.ScheduleCount(ctx, readStringField(cmd, paramRecordId), readStringField(cmd, paramNextCountDate))
}

func (this *StockQuantApplicationServiceImpl) AssignCounter(
	ctx corectx.Context, cmd itStock.AssignCounterCommand,
) (*itStock.CountMutateResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionAssignCounter, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.quantSvc.AssignCounter(ctx, readStringField(cmd, paramRecordId), readStringField(cmd, paramAssignedUserId))
}
