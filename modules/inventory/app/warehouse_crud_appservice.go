package app

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewWarehouseApplicationService is handed the composable default by the warehouse onion, plus
// the location domain service the flow reconfigurations provision locations through.
func NewWarehouseApplicationService(
	base composable.CrudApplicationService, locationSvc *services.InventoryLocationDomainServiceImpl,
) itWarehouse.WarehouseApplicationService {
	warehouseSvc, ok := base.DomainService().(*services.WarehouseDomainServiceImpl)
	if !ok {
		panic(errors.New("the warehouse onion must be built with NewWarehouseDomainService"))
	}
	return &WarehouseApplicationServiceImpl{
		CrudApplicationService: base,
		WarehouseAppService:    NewWarehouseAppService(warehouseSvc, locationSvc),
		warehouseSvc:           warehouseSvc,
	}
}

// WarehouseApplicationServiceImpl is the authorized CRUD and lifecycle of the warehouse, plus the
// orchestration port other code calls directly. The port methods assert nothing: their callers
// act on their own authority. The *Action methods are what the REST routes reach, and they do.
type WarehouseApplicationServiceImpl struct {
	composable.CrudApplicationService
	itWarehouse.WarehouseAppService
	warehouseSvc *services.WarehouseDomainServiceImpl
}

func (this *WarehouseApplicationServiceImpl) Suspend(
	ctx corectx.Context, cmd itWarehouse.SuspendWarehouseCommand,
) (*itWarehouse.SuspendWarehouseResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionSuspend, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.warehouseSvc.Suspend(ctx, readStringField(cmd, paramRecordId))
}

func (this *WarehouseApplicationServiceImpl) Resume(
	ctx corectx.Context, cmd itWarehouse.ResumeWarehouseCommand,
) (*itWarehouse.ResumeWarehouseResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionResume, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.warehouseSvc.Resume(ctx, readStringField(cmd, paramRecordId))
}

// The flow reconfigurations reach the orchestration port rather than the domain service because
// each writes the warehouse and provisions its locations together.
func (this *WarehouseApplicationServiceImpl) ConfigureIncomingFlowAction(
	ctx corectx.Context, cmd itWarehouse.ConfigureWarehouseFlowCommand,
) (*itWarehouse.ConfigureWarehouseFlowResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionConfigureIncomingFlow, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return toFlowMutateResult(this.ConfigureIncomingFlow(ctx, flowCommand(cmd)))
}

func (this *WarehouseApplicationServiceImpl) ConfigureOutgoingFlowAction(
	ctx corectx.Context, cmd itWarehouse.ConfigureWarehouseFlowCommand,
) (*itWarehouse.ConfigureWarehouseFlowResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionConfigureOutgoingFlow, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return toFlowMutateResult(this.ConfigureOutgoingFlow(ctx, flowCommand(cmd)))
}

func flowCommand(cmd itWarehouse.ConfigureWarehouseFlowCommand) itWarehouse.ConfigureFlowCommand {
	return itWarehouse.ConfigureFlowCommand{
		WarehouseId: readStringField(cmd, paramRecordId),
		Flow:        readStringField(cmd, "flow"),
	}
}

func toFlowMutateResult(
	result *itWarehouse.ConfigureFlowResult, err error,
) (*itWarehouse.ConfigureWarehouseFlowResult, error) {
	if err != nil {
		return nil, err
	}
	out := &itWarehouse.ConfigureWarehouseFlowResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
	}
	if result.HasData {
		out.Data = dyn.MutateResultData{AffectedCount: result.Data.AffectedCount}
	}
	return out, nil
}
