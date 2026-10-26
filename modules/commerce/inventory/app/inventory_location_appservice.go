package app

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewInventoryLocationApplicationService is handed the composable default by the location onion.
func NewInventoryLocationApplicationService(base composable.CrudApplicationService) itWarehouse.InventoryLocationApplicationService {
	locationSvc, ok := base.DomainService().(*services.InventoryLocationDomainServiceImpl)
	if !ok {
		panic(errors.New("the inventory location onion must be built with NewInventoryLocationDomainService"))
	}
	return &InventoryLocationApplicationServiceImpl{CrudApplicationService: base, locationSvc: locationSvc}
}

type InventoryLocationApplicationServiceImpl struct {
	composable.CrudApplicationService
	locationSvc *services.InventoryLocationDomainServiceImpl
}

func (this *InventoryLocationApplicationServiceImpl) Suspend(
	ctx corectx.Context, cmd itWarehouse.SuspendInventoryLocationCommand,
) (*itWarehouse.SuspendInventoryLocationResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionSuspend, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.locationSvc.Suspend(ctx, readStringField(cmd, paramRecordId))
}

func (this *InventoryLocationApplicationServiceImpl) Resume(
	ctx corectx.Context, cmd itWarehouse.ResumeInventoryLocationCommand,
) (*itWarehouse.ResumeInventoryLocationResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionResume, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.locationSvc.Resume(ctx, readStringField(cmd, paramRecordId))
}

// Move re-parents a location. An empty parent makes it a root, which is why the parameter is
// read without being required.
func (this *InventoryLocationApplicationServiceImpl) Move(
	ctx corectx.Context, cmd itWarehouse.MoveInventoryLocationCommand,
) (*itWarehouse.MoveInventoryLocationResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionMoveLocation, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.locationSvc.Move(ctx, readStringField(cmd, paramRecordId), readStringField(cmd, "parent_location_id"))
}
