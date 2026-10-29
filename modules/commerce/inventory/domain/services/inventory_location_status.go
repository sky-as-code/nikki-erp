package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The operational state of a location: suspending and resuming it, and what each requires.

// Suspend takes a location out of use while leaving everything it holds where it is. There is
// deliberately no emptiness check — suspending a rack that holds goods is the point — but it does
// refuse to strand work already under way rather than silently unreserving or cancelling a move.
func (this *InventoryLocationDomainServiceImpl) Suspend(
	ctx corectx.Context, locationId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	location, vErrs, err := this.loadLocation(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if derefString(location.GetStatus()) == models.InventoryLocationStatusSuspended {
		return locationViolationResult(
			"", "inventory_location.already_suspended", "the location is already suspended"), nil
	}

	usage, err := this.readUsage(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if usage.OpenMoveCount > 0 || usage.OpenTransferCount > 0 {
		return locationViolationResult(
			"", "inventory_location.has_open_operations",
			"the location is part of an operation still in progress; finish or cancel it first",
		), nil
	}
	// Suspending takes the stock here out of the warehouse's sellable pool without moving it,
	// which is exactly what a reservation counting on that stock cannot survive.
	if vErrs, err := assertLocationNotBackingReservations(ctx, *location); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if err := this.writeStatus(ctx, locationId, models.InventoryLocationStatusSuspended); err != nil {
		return nil, err
	}
	return locationMutateOk(), nil
}

// Resume puts a suspended location back into use, once what it depends on is usable too.
func (this *InventoryLocationDomainServiceImpl) Resume(
	ctx corectx.Context, locationId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	location, vErrs, err := this.loadLocation(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if derefString(location.GetStatus()) == models.InventoryLocationStatusActive {
		return locationViolationResult(
			"", "inventory_location.already_active", "the location is already active"), nil
	}
	if vErrs, err := this.assertResumable(ctx, *location); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if err := this.writeStatus(ctx, locationId, models.InventoryLocationStatusActive); err != nil {
		return nil, err
	}
	return locationMutateOk(), nil
}

// assertResumable checks that what the location hangs off is itself usable. A location cannot be
// more available than its warehouse or parent: active under a suspended parent would advertise it
// for work the parent has been taken out of.
func (this *InventoryLocationDomainServiceImpl) assertResumable(
	ctx corectx.Context, location models.InventoryLocation,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	if location.GetWarehouseId() != nil {
		usable, err := this.isOwningWarehouseUsable(ctx, location)
		if err != nil {
			return vErrs, err
		}
		if !usable {
			appendLocationViolation(vErrs, "", "inventory_location.warehouse_not_usable",
				"the warehouse is suspended or archived; resume it first")
		}
	}

	parentId := derefId(location.GetParentLocationId())
	if parentId == "" {
		return vErrs, nil
	}
	parent, parentErrs, err := this.loadLocation(ctx, parentId)
	if err != nil || parentErrs.Count() > 0 {
		return vErrs, err
	}
	if parent.GetIsArchived() != nil && *parent.GetIsArchived() {
		appendLocationViolation(vErrs, "", "inventory_location.parent_archived",
			"the parent location is archived")
	}
	if derefString(parent.GetStatus()) == models.InventoryLocationStatusSuspended {
		appendLocationViolation(vErrs, "", "inventory_location.parent_suspended",
			"the parent location is suspended; resume it first")
	}
	return vErrs, nil
}

// readUsage asks Stock what it holds at a location.
func (this *InventoryLocationDomainServiceImpl) readUsage(
	ctx corectx.Context, locationId string,
) (itStock.LocationUsage, error) {
	if this.usage == nil {
		return itStock.LocationUsage{}, errors.New(
			"the location service was built without a stock usage reader; this is a wiring bug")
	}
	result, err := this.usage.GetLocationUsage(ctx, itStock.GetLocationUsageQuery{LocationId: locationId})
	if err != nil {
		return itStock.LocationUsage{}, err
	}
	if result == nil || !result.HasData {
		return itStock.LocationUsage{}, nil
	}
	return result.Data.Usage, nil
}
