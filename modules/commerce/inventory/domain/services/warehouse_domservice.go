package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewWarehouseDomainService derives the warehouse service from the engine's default one. locations
// is the location service a warehouse provisions, archives and deletes its managed locations through.
func NewWarehouseDomainService(
	base composable.CrudDomainService, locations *InventoryLocationDomainServiceImpl,
) *WarehouseDomainServiceImpl {
	return &WarehouseDomainServiceImpl{CrudDomainService: base, locations: locations}
}

// WarehouseDomainServiceImpl holds the rules about a warehouse: its code, its operational state, and
// the locations it manages (root, Stock, flow stops), which it creates, archives, unarchives and
// deletes together with itself, in one transaction.
type WarehouseDomainServiceImpl struct {
	composable.CrudDomainService

	locations *InventoryLocationDomainServiceImpl
}

var _ composable.CrudDomainService = (*WarehouseDomainServiceImpl)(nil)

// Create is the client's create. The managed fields are stripped: a client able to set them could
// mint a warehouse that then refuses to be deleted, as if a kiosk owned it.
func (this *WarehouseDomainServiceImpl) Create(
	ctx corectx.Context, cmd dmodel.DynamicFields, options ...composable.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	prepared := copyFields(cmd)
	delete(prepared, basemodel.FieldManagedByType)
	delete(prepared, basemodel.FieldManagedById)
	return this.create(ctx, prepared, options...)
}

// CreateManaged is the create of a warehouse something manages - a kiosk's - and keeps the
// managed_by_type and managed_by_id the caller set. No REST route reaches it.
func (this *WarehouseDomainServiceImpl) CreateManaged(
	ctx corectx.Context, cmd dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	return this.create(ctx, copyFields(cmd))
}

// create writes the warehouse, its root, its Stock and the stops its flows need, in one transaction
// (the caller's, when ctx carries one): a warehouse whose locations failed is not created either.
// The location FKs are written last, once the locations exist.
func (this *WarehouseDomainServiceImpl) create(
	ctx corectx.Context, cmd dmodel.DynamicFields, options ...composable.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	if vErrs := AssertWarehouseFlowsKnown(cmd); vErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}
	opts := safe.GetOptional(options, composable.CreateOptions{})

	var result *dyn.OpResult[dmodel.DynamicFields]
	err := runInTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		created, err := this.CrudDomainService.Create(tranxCtx, cmd, opts)
		if err != nil {
			return err
		}
		if created == nil || created.ClientErrors.Count() > 0 {
			return composable.ClientErrorSignal{Errors: clientErrorsOfFields(created)}
		}
		result = created
		locationFks, err := this.provisionLocations(tranxCtx, *models.NewWarehouseFrom(created.Data))
		if err != nil {
			return err
		}
		for field, value := range locationFks {
			created.Data[field] = value
		}
		return nil
	})
	if signal, ok := composable.AsClientErrorSignal(err); ok {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: signal.Errors}, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update protects the code once the warehouse exists: the code named its root location, and
// renaming needs its own operation. A name change leaves the root location's name alone.
func (this *WarehouseDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.UpdateOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	warehouseId := readStringParam(params, models.WarehouseFieldId)
	current, vErrs, err := this.loadWarehouse(ctx, warehouseId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	prepared := copyFields(params)
	// Status moves through suspend and resume, which check what those transitions require.
	delete(prepared, models.WarehouseFieldStatus)

	if code := readStringParam(prepared, models.WarehouseFieldCode); code != "" &&
		code != derefString(current.GetCode()) {
		return warehouseViolationResult(models.WarehouseFieldCode, "warehouse.code_immutable",
			"the code named this warehouse's root location; it cannot be changed by an update"), nil
	}
	return this.CrudDomainService.Update(ctx, prepared, options...)
}

// SetArchived archives or unarchives the warehouse and the locations it manages together. One with
// live supply routes or reservations cannot be archived, and each location's own guard applies on
// top: a warehouse whose Stock or stops still hold stock, open work or a user's unarchived
// sub-location stays as it is. An unarchived warehouse lands on suspended rather than active.
func (this *WarehouseDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	warehouseId := readStringParam(params, models.WarehouseFieldId)
	warehouse, vErrs, err := this.loadWarehouse(ctx, warehouseId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	archiving := readBoolParam(params, paramIsArchived)
	if archiving {
		if vErrs, err := this.assertArchivable(ctx, *warehouse); err != nil {
			return nil, err
		} else if vErrs.Count() > 0 {
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
		}
	}
	return this.inRefusableTransaction(ctx, func(tranxCtx corectx.Context) (*dyn.OpResult[dyn.MutateResultData], error) {
		result, err := this.CrudDomainService.SetArchived(tranxCtx, params)
		if err != nil || result == nil || result.ClientErrors.Count() > 0 {
			return result, err
		}
		if err := this.setManagedLocationsArchived(tranxCtx, *warehouse, archiving); err != nil {
			return nil, err
		}
		if archiving {
			return result, nil
		}
		// Unarchiving never returns a warehouse straight to service: the configuration may have
		// changed while it was away, so someone confirms it through Resume.
		return result, this.WriteStatus(tranxCtx, warehouseId, models.WarehouseStatusSuspended)
	})
}

// Suspend closes a warehouse temporarily, leaving its locations untouched. Child locations are
// deliberately not cascaded — rewriting thousands would be slow and would lose their own state on
// resume — so usability is read from both records: a location in a suspended warehouse is unusable
// however the location itself reads.
func (this *WarehouseDomainServiceImpl) Suspend(
	ctx corectx.Context, warehouseId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	warehouse, vErrs, err := this.loadWarehouse(ctx, warehouseId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if derefString(warehouse.GetStatus()) == models.WarehouseStatusSuspended {
		return warehouseViolationResult(
			"", "warehouse.already_suspended", "the warehouse is already suspended"), nil
	}
	// A suspended warehouse's stock is no longer eligible, so a reservation in force on it would
	// lose its backing silently.
	orgId := derefId(warehouse.GetOrgId())
	if vErrs, err := assertWarehouseHasNoEffectiveReservations(ctx, orgId, warehouseId); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if err := this.WriteStatus(ctx, warehouseId, models.WarehouseStatusSuspended); err != nil {
		return nil, err
	}
	return locationMutateOk(), nil
}

// Resume puts a suspended warehouse back into service.
func (this *WarehouseDomainServiceImpl) Resume(
	ctx corectx.Context, warehouseId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	warehouse, vErrs, err := this.loadWarehouse(ctx, warehouseId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if derefString(warehouse.GetStatus()) == models.WarehouseStatusActive {
		return warehouseViolationResult(
			"", "warehouse.already_active", "the warehouse is already active"), nil
	}
	if warehouse.GetIsArchived() != nil && *warehouse.GetIsArchived() {
		return warehouseViolationResult(
			"", "warehouse.archived", "unarchive the warehouse before resuming it"), nil
	}

	if err := this.WriteStatus(ctx, warehouseId, models.WarehouseStatusActive); err != nil {
		return nil, err
	}
	return locationMutateOk(), nil
}

// assertArchivable refuses to archive a warehouse other live configuration still depends on.
func (this *WarehouseDomainServiceImpl) assertArchivable(
	ctx corectx.Context, warehouse models.Warehouse,
) (*ft.ClientErrors, error) {
	warehouseId := derefString(warehouse.GetId())
	vErrs, err := assertWarehouseHasNoEffectiveReservations(ctx, derefId(warehouse.GetOrgId()), warehouseId)
	if err != nil {
		return vErrs, err
	}

	relations, err := this.countUnarchivedSupplyRelations(ctx, warehouseId)
	if err != nil {
		return vErrs, err
	}
	if relations > 0 {
		appendWarehouseViolation(vErrs, "", "warehouse.has_supply_relations",
			"archive the supply relations that name this warehouse first")
	}
	return vErrs, nil
}

// clientErrorsOfFields reads the violations off a create result, tolerating a nil one.
func clientErrorsOfFields(result *dyn.OpResult[dmodel.DynamicFields]) ft.ClientErrors {
	if result == nil {
		return *ft.NewClientErrors()
	}
	return result.ClientErrors
}
