package app

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// resolveLegLocations turns a plan's stops into location ids: "Stock" is the warehouse's stock
// location, every other stop the flow location its FK names.
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
	return services.ResolveWarehouseStop(warehouse, stop), nil
}
