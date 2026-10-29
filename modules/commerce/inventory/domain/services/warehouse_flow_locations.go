package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The lifecycle of the locations a warehouse manages after creation: flow stops come and go with the
// flows, and the whole set is archived and unarchived with the warehouse. A stop is never deleted
// while its warehouse lives: the moves that passed through it still name it.

// ReconfigureFlowLocations brings the stops in line with a flow change, in the caller's transaction:
// the stops the new flow uses are created or unarchived, and those it no longer uses are archived.
// An archive the location guard refuses (stock or open work at the stop) refuses the change.
func (this *WarehouseDomainServiceImpl) ReconfigureFlowLocations(
	ctx corectx.Context, warehouse models.Warehouse, previousFlow string, nextFlow string, outgoing bool,
) error {
	if derefId(warehouse.GetRootLocationId()) == "" {
		vErrs := ft.NewClientErrors()
		appendWarehouseViolation(vErrs, "", "warehouse.root_location_missing",
			"the warehouse has no root location to create its flow locations under")
		return composable.ClientErrorSignal{Errors: *vErrs}
	}
	needed := incomingFlowLocations(nextFlow)
	if outgoing {
		needed = outgoingFlowLocations(nextFlow)
	}
	if err := this.ensureFlowLocations(ctx, &warehouse, needed); err != nil {
		return err
	}
	for _, spec := range obsoleteSystemLocations(previousFlow, nextFlow, outgoing) {
		locationId := derefId(warehouse.GetFlowLocationId(spec.Field))
		if locationId == "" {
			continue
		}
		if err := this.setManagedLocationArchived(ctx, locationId, true); err != nil {
			return err
		}
	}
	return nil
}

// ensureFlowLocations makes each stop present and in use: a stop the warehouse already references
// is reused (unarchived when a former flow change archived it), and one it lacks is created under
// the root and written to its FK. Switching a flow back and forth never accumulates copies.
func (this *WarehouseDomainServiceImpl) ensureFlowLocations(
	ctx corectx.Context, warehouse *models.Warehouse, specs []systemLocationSpec,
) error {
	rootId := derefId(warehouse.GetRootLocationId())
	for _, spec := range specs {
		if existing := derefId(warehouse.GetFlowLocationId(spec.Field)); existing != "" {
			if err := this.setManagedLocationArchived(ctx, existing, false); err != nil {
				return err
			}
			continue
		}
		created, err := this.createManagedLocation(ctx, *warehouse,
			spec.Name, models.InventoryLocationTypeInternal, rootId)
		if err != nil {
			return err
		}
		if err := this.pointWarehouseAt(ctx, warehouse, dmodel.DynamicFields{spec.Field: created}); err != nil {
			return err
		}
	}
	return nil
}

// setManagedLocationsArchived applies the warehouse's archive flag to the locations it manages, in
// the caller's transaction. Archiving goes leaves first (stops, Stock, then the root), since a
// location with unarchived children cannot be archived. Unarchiving goes the other way and brings
// back only the stops the current flows use: the others stay archived, as the flow change left them.
func (this *WarehouseDomainServiceImpl) setManagedLocationsArchived(
	ctx corectx.Context, warehouse models.Warehouse, isArchived bool,
) error {
	fields := warehouseLocationFields
	if !isArchived {
		fields = []string{models.WarehouseFieldRootLocationId, models.WarehouseFieldStockLocationId}
		specs := append(incomingFlowLocations(derefString(warehouse.GetIncomingFlow())),
			outgoingFlowLocations(derefString(warehouse.GetOutgoingFlow()))...)
		for _, spec := range specs {
			fields = append(fields, spec.Field)
		}
	}
	for _, field := range fields {
		locationId := derefId(warehouse.GetFlowLocationId(field))
		if locationId == "" {
			continue
		}
		if err := this.setManagedLocationArchived(ctx, locationId, isArchived); err != nil {
			return err
		}
	}
	return nil
}

// setManagedLocationArchived archives or unarchives one managed location through the location
// service, so its archive guard applies, turning a refusal into a signal that rolls back the
// caller's transaction.
func (this *WarehouseDomainServiceImpl) setManagedLocationArchived(
	ctx corectx.Context, locationId string, isArchived bool,
) error {
	result, err := this.locations.SetArchivedManaged(ctx, locationId, isArchived)
	if err != nil {
		return err
	}
	if result != nil && result.ClientErrors.Count() > 0 {
		return composable.ClientErrorSignal{Errors: result.ClientErrors}
	}
	return nil
}
