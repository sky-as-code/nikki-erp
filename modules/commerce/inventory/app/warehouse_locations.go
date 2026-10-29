package app

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ensureFlowLocations creates whatever a newly chosen flow needs and does not already have, under
// the warehouse's root. An existing but suspended stop is resumed rather than duplicated, so
// switching a flow back and forth does not accumulate copies.
func (this *WarehouseAppServiceImpl) ensureFlowLocations(
	ctx corectx.Context, warehouse models.Warehouse, flow string, outgoing bool,
) error {
	specs := services.IncomingFlowLocations(flow)
	if outgoing {
		specs = services.OutgoingFlowLocations(flow)
	}

	for _, spec := range specs {
		existing, err := services.FindWarehouseFlowLocation(ctx, warehouse, spec.Purpose)
		if err != nil {
			return err
		}
		if existing == nil {
			if _, err := this.warehouseSvc.CreateFlowLocation(ctx, warehouse, spec); err != nil {
				return err
			}
			continue
		}

		if derefString(existing.GetStatus()) == models.InventoryLocationStatusSuspended {
			if _, err := this.locationSvc.Resume(ctx, derefString((*string)(existing.GetId()))); err != nil {
				return err
			}
		}
	}
	return nil
}

// suspendObsoleteLocations retires the stops a reduced flow no longer uses. Suspended, never
// deleted: past moves still name those locations and the history has to keep resolving. A location
// holding stock is suspended too, leaving the goods visible where they are.
func (this *WarehouseAppServiceImpl) suspendObsoleteLocations(
	ctx corectx.Context, warehouse models.Warehouse, previousFlow string, nextFlow string, outgoing bool,
) error {
	for _, spec := range services.ObsoleteSystemLocations(previousFlow, nextFlow, outgoing) {
		existing, err := services.FindWarehouseFlowLocation(ctx, warehouse, spec.Purpose)
		if err != nil {
			return err
		}
		if existing == nil || derefString(existing.GetStatus()) == models.InventoryLocationStatusSuspended {
			continue
		}
		if err := services.WriteLocationStatus(ctx,
			derefString((*string)(existing.GetId())), models.InventoryLocationStatusSuspended); err != nil {
			return err
		}
	}
	return nil
}

// resolveLegLocations turns a plan's stops into location ids: "Stock" is the warehouse's root.
// Endpoints outside the warehouse stay unresolved: which vendor or customer location applies
// depends on the transaction, so the movement engine fills those in.
func (this *WarehouseAppServiceImpl) resolveLegLocations(
	ctx corectx.Context, warehouse models.Warehouse, legs []services.FlowLeg,
) ([]itWarehouse.ResolvedLeg, error) {
	resolved := make([]itWarehouse.ResolvedLeg, 0, len(legs))

	for _, leg := range legs {
		fromId, err := this.locationIdForStop(ctx, warehouse, leg.FromCode)
		if err != nil {
			return nil, err
		}
		toId, err := this.locationIdForStop(ctx, warehouse, leg.ToCode)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, itWarehouse.ResolvedLeg{
			FromLocationId: fromId,
			ToLocationId:   toId,
			FromCode:       leg.FromCode,
			ToCode:         leg.ToCode,
		})
	}
	return resolved, nil
}

func (this *WarehouseAppServiceImpl) locationIdForStop(
	ctx corectx.Context, warehouse models.Warehouse, stop string,
) (string, error) {
	if stop == services.FlowEndpointVendor || stop == services.FlowEndpointCustomer {
		return "", nil
	}
	return services.ResolveWarehouseStop(ctx, warehouse, stop)
}
