package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// A warehouse and the locations it manages live and go together. Each is managed by the warehouse
// (managed_by_type commerce/inventory_warehouse, managed_by_id the warehouse) and referenced from it
// by an ON DELETE RESTRICT FK:
//   - root_location_id: a group location named after the warehouse code, the top of its tree. It
//     holds no stock.
//   - stock_location_id: the internal "Stock" child of the root, where the warehouse's goods live.
//   - receiving, quality, packing and output_location_id: the flow stops, children of the root,
//     created when a flow first needs one and archived (FK kept) when a flow stops using it.

// warehouseLocationFields lists every FK a warehouse holds to a location it manages, in the order
// they are archived and deleted: the flow stops, then Stock, then the root everything hangs under.
var warehouseLocationFields = []string{
	models.WarehouseFieldReceivingLocationId,
	models.WarehouseFieldQualityLocationId,
	models.WarehouseFieldPackingLocationId,
	models.WarehouseFieldOutputLocationId,
	models.WarehouseFieldStockLocationId,
	models.WarehouseFieldRootLocationId,
}

// provisionLocations creates the root, Stock and the flow stops of a new warehouse and points the
// warehouse at them. It returns the location FKs written, for the caller's create result. Every
// location created here is marked system-generated, which protects it from being restructured.
func (this *WarehouseDomainServiceImpl) provisionLocations(
	ctx corectx.Context, warehouse models.Warehouse,
) (dmodel.DynamicFields, error) {
	rootId, err := this.createManagedLocation(ctx, warehouse,
		derefString(warehouse.GetCode()), models.InventoryLocationTypeGroup, "")
	if err != nil {
		return nil, err
	}
	stockId, err := this.createManagedLocation(ctx, warehouse,
		warehouseStockLocationCode, models.InventoryLocationTypeInternal, rootId)
	if err != nil {
		return nil, err
	}
	if err := this.pointWarehouseAt(ctx, &warehouse, dmodel.DynamicFields{
		models.WarehouseFieldRootLocationId:  rootId,
		models.WarehouseFieldStockLocationId: stockId,
	}); err != nil {
		return nil, err
	}

	specs := append(incomingFlowLocations(derefString(warehouse.GetIncomingFlow())),
		outgoingFlowLocations(derefString(warehouse.GetOutgoingFlow()))...)
	if err := this.ensureFlowLocations(ctx, &warehouse, specs); err != nil {
		return nil, err
	}
	written := dmodel.DynamicFields{}
	for _, field := range warehouseLocationFields {
		if id := warehouse.GetFlowLocationId(field); id != nil {
			written[field] = string(*id)
		}
	}
	return written, nil
}

// createManagedLocation creates one location the warehouse manages: the root when parentId is
// empty, otherwise a child of the root.
func (this *WarehouseDomainServiceImpl) createManagedLocation(
	ctx corectx.Context, warehouse models.Warehouse, name string, locationType string, parentId string,
) (string, error) {
	fields := dmodel.DynamicFields{
		models.InventoryLocationFieldName:         name,
		models.InventoryLocationFieldLocationType: locationType,
		models.InventoryLocationFieldOrgId:        derefId(warehouse.GetOrgId()),
		models.InventoryLocationFieldStatus:       models.InventoryLocationStatusActive,
		basemodel.FieldManagedByType:              models.InventoryLocationManagedByWarehouse,
		basemodel.FieldManagedById:                derefString(warehouse.GetId()),
	}
	if parentId != "" {
		fields[models.InventoryLocationFieldParentLocationId] = parentId
	}
	return this.createSystemLocation(ctx, fields, true)
}

// pointWarehouseAt writes location FKs onto the warehouse, and onto the caller's copy of it.
func (this *WarehouseDomainServiceImpl) pointWarehouseAt(
	ctx corectx.Context, warehouse *models.Warehouse, fks dmodel.DynamicFields,
) error {
	if err := writeWarehouseFields(ctx, derefString(warehouse.GetId()), fks); err != nil {
		return err
	}
	for field, value := range fks {
		warehouse.GetFieldData()[field] = value
	}
	return nil
}

// createSystemLocation writes one location the warehouse owns through the location service, so the
// tree rules and derived fields apply, then sets is_system_generated directly, because the service
// strips that flag from anything a caller sends.
func (this *WarehouseDomainServiceImpl) createSystemLocation(
	ctx corectx.Context, fields dmodel.DynamicFields, managed bool,
) (string, error) {
	var result *dyn.OpResult[dmodel.DynamicFields]
	var err error
	if managed {
		result, err = this.locations.CreateManaged(ctx, fields)
	} else {
		result, err = this.locations.Create(ctx, fields)
	}
	if err != nil {
		return "", err
	}
	if result == nil || result.ClientErrors.Count() > 0 {
		return "", composable.ClientErrorSignal{Errors: clientErrorsOfFields(result)}
	}
	locationId := readStringParam(result.Data, models.InventoryLocationFieldId)
	return locationId, MarkLocationSystemGenerated(ctx, locationId)
}

// Delete deletes the warehouse and then every location it manages, in one transaction. Deleting the
// warehouse first releases its RESTRICT FKs and sets warehouse_id NULL on its locations; deleting a
// managed location then fails, rolling everything back, while a user's location still hangs under it
// or a move still names it.
func (this *WarehouseDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	return this.deleteWithLocations(ctx, cmd, options, this.CrudDomainService.Delete)
}

// DeleteManaged is the manager's delete of a managed warehouse, with the same cascade as Delete.
func (this *WarehouseDomainServiceImpl) DeleteManaged(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	return this.deleteWithLocations(ctx, cmd, options, this.CrudDomainService.DeleteManaged)
}

type deleteFn = func(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error)

func (this *WarehouseDomainServiceImpl) deleteWithLocations(
	ctx corectx.Context, cmd composable.DeleteCommand, options []composable.DeleteOptions, del deleteFn,
) (*composable.MutateResult, error) {
	warehouse, vErrs, err := this.loadWarehouse(ctx, readStringParam(cmd, models.WarehouseFieldId))
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: *vErrs}, nil
	}
	return this.inRefusableTransaction(ctx, func(tranxCtx corectx.Context) (*composable.MutateResult, error) {
		result, err := del(tranxCtx, cmd, options...)
		if err != nil || result == nil || result.ClientErrors.Count() > 0 {
			return result, err
		}
		for _, field := range warehouseLocationFields {
			locationId := derefId(warehouse.GetFlowLocationId(field))
			if locationId == "" {
				continue
			}
			deleted, err := this.locations.DeleteManaged(tranxCtx, dmodel.DynamicFields{basemodel.FieldId: locationId})
			if err != nil || deleted == nil || deleted.ClientErrors.Count() > 0 {
				return deleted, err
			}
		}
		return result, nil
	})
}

// inRefusableTransaction runs body in one transaction, joining the caller's when it has one, and
// rolls back when the body's result carries client errors.
func (this *WarehouseDomainServiceImpl) inRefusableTransaction(
	ctx corectx.Context, body func(tranxCtx corectx.Context) (*dyn.OpResult[dyn.MutateResultData], error),
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]
	cErrs, err := composable.InRefusableTransaction(ctx, this.Repository(),
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			res, err := body(tranxCtx)
			if err != nil {
				return nil, err
			}
			result = res
			if res == nil {
				return nil, errors.New("the warehouse operation returned no result")
			}
			return &res.ClientErrors, nil
		})
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *cErrs}, nil
	}
	return result, nil
}

// AssertWarehouseFlowsKnown refuses a command naming a flow that does not exist, before anything
// is written.
func AssertWarehouseFlowsKnown(fields dmodel.DynamicFields) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	for _, field := range []string{models.WarehouseFieldIncomingFlow, models.WarehouseFieldOutgoingFlow} {
		flow := readStringParam(fields, field)
		if flow != "" && !isKnownFlow(flow) {
			appendWarehouseViolation(vErrs, field, "warehouse.flow_invalid", "'"+flow+"' is not a known flow")
		}
	}
	return vErrs
}
