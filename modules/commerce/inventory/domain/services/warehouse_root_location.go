package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// A warehouse and its root location live and go together. The root is a location the warehouse
// manages (managed_by_type commerce/inventory_warehouse, managed_by_id the warehouse), referenced by
// warehouse.root_location_id with an ON DELETE RESTRICT FK. It is where the warehouse's stock lives
// and the parent of the flow stops (Input, Quality Control, Packing, Output) and of every location a
// user adds to the warehouse.

// provisionLocations creates the root and the flow stops of a new warehouse, then points the
// warehouse at its root. Every location created here is marked system-generated, which protects
// it from being restructured while the warehouse needs it.
func (this *WarehouseDomainServiceImpl) provisionLocations(
	ctx corectx.Context, warehouse models.Warehouse,
) (string, error) {
	warehouseId := derefString(warehouse.GetId())
	code := derefString(warehouse.GetCode())
	rootId, err := this.createSystemLocation(ctx, dmodel.DynamicFields{
		models.InventoryLocationFieldName:          code,
		models.InventoryLocationFieldLocationUsage: models.InventoryLocationUsageInternal,
		models.InventoryLocationFieldPurpose:       models.InventoryLocationPurposeStorage,
		models.InventoryLocationFieldOrgId:         derefId(warehouse.GetOrgId()),
		models.InventoryLocationFieldStatus:        models.InventoryLocationStatusActive,
		basemodel.FieldManagedByType:               models.InventoryLocationManagedByWarehouse,
		basemodel.FieldManagedById:                 warehouseId,
	}, true)
	if err != nil {
		return "", err
	}
	if err := writeWarehouseFields(ctx, warehouseId, dmodel.DynamicFields{
		models.WarehouseFieldRootLocationId: rootId,
	}); err != nil {
		return "", err
	}
	rootLocationId := model.Id(rootId)
	warehouse.SetRootLocationId(&rootLocationId)

	specs := append(incomingFlowLocations(derefString(warehouse.GetIncomingFlow())),
		outgoingFlowLocations(derefString(warehouse.GetOutgoingFlow()))...)
	for _, spec := range specs {
		if _, err := this.CreateFlowLocation(ctx, warehouse, SystemLocationSpec(spec)); err != nil {
			return "", err
		}
	}
	return rootId, nil
}

// CreateFlowLocation creates one stop of a warehouse flow under the warehouse's root.
func (this *WarehouseDomainServiceImpl) CreateFlowLocation(
	ctx corectx.Context, warehouse models.Warehouse, spec SystemLocationSpec,
) (string, error) {
	rootId := derefId(warehouse.GetRootLocationId())
	if rootId == "" {
		vErrs := ft.NewClientErrors()
		appendWarehouseViolation(vErrs, "", "warehouse.root_location_missing",
			"the warehouse has no root location to create its flow locations under")
		return "", composable.ClientErrorSignal{Errors: *vErrs}
	}
	return this.createSystemLocation(ctx, dmodel.DynamicFields{
		models.InventoryLocationFieldName:             spec.Name,
		models.InventoryLocationFieldLocationUsage:    models.InventoryLocationUsageInternal,
		models.InventoryLocationFieldPurpose:          spec.Purpose,
		models.InventoryLocationFieldParentLocationId: rootId,
		models.InventoryLocationFieldOrgId:            derefId(warehouse.GetOrgId()),
		models.InventoryLocationFieldStatus:           models.InventoryLocationStatusActive,
	}, false)
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

// setRootArchived archives or unarchives the root with its warehouse, in the caller's transaction.
// A warehouse created before root locations existed has none, and nothing to cascade to.
func (this *WarehouseDomainServiceImpl) setRootArchived(
	ctx corectx.Context, warehouse models.Warehouse, isArchived bool,
) error {
	rootId := derefId(warehouse.GetRootLocationId())
	if rootId == "" {
		return nil
	}
	result, err := this.locations.SetArchivedManaged(ctx, rootId, isArchived)
	if err != nil {
		return err
	}
	if result != nil && result.ClientErrors.Count() > 0 {
		return composable.ClientErrorSignal{Errors: result.ClientErrors}
	}
	return nil
}

// Delete deletes the warehouse and then its root location, in one transaction. Deleting the
// warehouse first releases the RESTRICT FK on root_location_id and sets warehouse_id NULL on its
// locations; deleting the root then fails, rolling everything back, while any location still hangs
// under it.
func (this *WarehouseDomainServiceImpl) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	return this.deleteWithRoot(ctx, cmd, options, this.CrudDomainService.Delete)
}

// DeleteManaged is the manager's delete of a managed warehouse, with the same cascade as Delete.
func (this *WarehouseDomainServiceImpl) DeleteManaged(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	return this.deleteWithRoot(ctx, cmd, options, this.CrudDomainService.DeleteManaged)
}

type deleteFn = func(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error)

func (this *WarehouseDomainServiceImpl) deleteWithRoot(
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
		rootId := derefId(warehouse.GetRootLocationId())
		if rootId == "" {
			return result, nil
		}
		rootResult, err := this.locations.DeleteManaged(tranxCtx, dmodel.DynamicFields{basemodel.FieldId: rootId})
		if err != nil || rootResult == nil || rootResult.ClientErrors.Count() > 0 {
			return rootResult, err
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
