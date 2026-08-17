package services

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// warehouseHierarchyScanLimit bounds the walk up the parent chain, guarding against a cycle that
// somehow got persisted rather than expressing a real depth limit.
const warehouseHierarchyScanLimit = 50

// NewWarehouseDomainService derives the warehouse service from the engine's default one.
func NewWarehouseDomainService(base drif.DynamicResourceService) *WarehouseDomainServiceImpl {
	return &WarehouseDomainServiceImpl{DynamicResourceService: base}
}

// WarehouseDomainServiceImpl holds the rules about a warehouse considered on its own: its
// hierarchy, its code, and its operational state.
//
// The operations that also touch locations — creating a warehouse with its tree, reconfiguring a
// flow — live on the application service instead, because they span two resources and have to be
// atomic across both.
type WarehouseDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*WarehouseDomainServiceImpl)(nil)

// Update keeps the hierarchy acyclic and protects the code once the warehouse is in use.
//
// The code is the first segment of every system location's path, so changing it after those exist
// would leave the paths describing a warehouse that no longer goes by that name. Renaming would
// have to rewrite them all, which is a dedicated operation rather than a side effect of an edit.
func (this *WarehouseDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
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
		return warehouseViolationResult("warehouse.code_immutable",
			"the code names this warehouse's locations; it cannot be changed by an update"), nil
	}
	if vErrs, err := this.assertParentValid(ctx, prepared, warehouseId); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	return this.DynamicResourceService.Update(ctx, prepared)
}

// SetArchived guards archiving, and leaves an unarchived warehouse suspended rather than active.
//
// A warehouse still holding up other things — live children, live supply routes — cannot be
// archived, because archiving it would leave those pointing at something withdrawn from use.
func (this *WarehouseDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	warehouseId := readStringParam(params, models.WarehouseFieldId)
	if _, vErrs, err := this.loadWarehouse(ctx, warehouseId); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if readBoolParam(params, paramIsArchived) {
		if vErrs, err := this.assertArchivable(ctx, warehouseId); err != nil {
			return nil, err
		} else if vErrs.Count() > 0 {
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
		}
		return this.DynamicResourceService.SetArchived(ctx, params)
	}

	result, err := this.DynamicResourceService.SetArchived(ctx, params)
	if err != nil || result == nil || result.ClientErrors.Count() > 0 {
		return result, err
	}
	// Unarchiving never returns a warehouse straight to service: what it sat in may have changed
	// while it was away, so someone confirms the configuration through Resume.
	return result, this.WriteStatus(ctx, warehouseId, models.WarehouseStatusSuspended)
}

// Suspend closes a warehouse temporarily, leaving its locations and everything in them untouched.
//
// The child locations are deliberately not cascaded: a warehouse can hold thousands, and rewriting
// them all to say what the warehouse already says would be slow and would lose their own state on
// resume. Usability is read from both records instead — a location in a suspended warehouse is
// unusable however the location itself reads.
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
			"warehouse.already_suspended", "the warehouse is already suspended"), nil
	}

	if err := this.WriteStatus(ctx, warehouseId, models.WarehouseStatusSuspended); err != nil {
		return nil, err
	}
	return locationMutateOk(), nil
}

// Resume puts a suspended warehouse back into service, once its configuration still holds.
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
			"warehouse.already_active", "the warehouse is already active"), nil
	}
	if warehouse.GetIsArchived() != nil && *warehouse.GetIsArchived() {
		return warehouseViolationResult(
			"warehouse.archived", "unarchive the warehouse before resuming it"), nil
	}
	if vErrs, err := this.assertResumable(ctx, *warehouse); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if err := this.WriteStatus(ctx, warehouseId, models.WarehouseStatusActive); err != nil {
		return nil, err
	}
	return locationMutateOk(), nil
}

// assertParentValid checks a parent warehouse: it exists, is usable, is in the same org, and does
// not sit underneath the warehouse being edited.
func (this *WarehouseDomainServiceImpl) assertParentValid(
	ctx corectx.Context, params dmodel.DynamicFields, selfId string,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	parentId := readStringParam(params, models.WarehouseFieldParentWarehouseId)
	if parentId == "" {
		return vErrs, nil
	}
	if parentId == selfId {
		appendWarehouseViolation(vErrs, "warehouse.hierarchy_cycle",
			"a warehouse cannot be its own parent")
		return vErrs, nil
	}

	parent, parentErrs, err := this.loadWarehouse(ctx, parentId)
	if err != nil {
		return vErrs, err
	}
	if parentErrs.Count() > 0 {
		appendWarehouseViolation(vErrs, "warehouse.parent_not_found",
			"no warehouse with id '"+parentId+"'")
		return vErrs, nil
	}
	if parent.GetIsArchived() != nil && *parent.GetIsArchived() {
		appendWarehouseViolation(vErrs, "warehouse.parent_archived",
			"the parent warehouse is archived")
		return vErrs, nil
	}

	if selfId != "" {
		inside, err := this.isDescendantOf(ctx, parentId, selfId)
		if err != nil {
			return vErrs, err
		}
		if inside {
			appendWarehouseViolation(vErrs, "warehouse.hierarchy_cycle",
				"the parent sits underneath this warehouse")
		}
	}
	return vErrs, nil
}

// assertResumable checks that the parent chain above a warehouse is itself in service.
func (this *WarehouseDomainServiceImpl) assertResumable(
	ctx corectx.Context, warehouse models.Warehouse,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	parentId := derefId(warehouse.GetParentWarehouseId())
	if parentId == "" {
		return vErrs, nil
	}
	parent, parentErrs, err := this.loadWarehouse(ctx, parentId)
	if err != nil || parentErrs.Count() > 0 {
		return vErrs, err
	}
	if parent.GetIsArchived() != nil && *parent.GetIsArchived() {
		appendWarehouseViolation(vErrs, "warehouse.parent_archived",
			"the parent warehouse is archived")
	}
	if derefString(parent.GetStatus()) == models.WarehouseStatusSuspended {
		appendWarehouseViolation(vErrs, "warehouse.parent_suspended",
			"the parent warehouse is suspended; resume it first")
	}
	return vErrs, nil
}

// assertArchivable refuses to archive a warehouse other live configuration still depends on.
func (this *WarehouseDomainServiceImpl) assertArchivable(
	ctx corectx.Context, warehouseId string,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	children, err := this.countUnarchivedChildren(ctx, warehouseId)
	if err != nil {
		return vErrs, err
	}
	if children > 0 {
		appendWarehouseViolation(vErrs, "warehouse.has_children",
			"archive the warehouses underneath this one first")
	}

	relations, err := this.countUnarchivedSupplyRelations(ctx, warehouseId)
	if err != nil {
		return vErrs, err
	}
	if relations > 0 {
		appendWarehouseViolation(vErrs, "warehouse.has_supply_relations",
			"archive the supply relations that name this warehouse first")
	}
	return vErrs, nil
}

// isDescendantOf walks up from candidateId to see whether ancestorId is above it.
func (this *WarehouseDomainServiceImpl) isDescendantOf(
	ctx corectx.Context, candidateId string, ancestorId string,
) (bool, error) {
	currentId := candidateId
	for hops := 0; hops < warehouseHierarchyScanLimit; hops++ {
		if currentId == "" {
			return false, nil
		}
		if currentId == ancestorId {
			return true, nil
		}
		current, vErrs, err := this.loadWarehouse(ctx, currentId)
		if err != nil || vErrs.Count() > 0 {
			return false, err
		}
		currentId = derefId(current.GetParentWarehouseId())
	}
	return false, errors.New("the warehouse hierarchy is deeper than " +
		fmt.Sprint(warehouseHierarchyScanLimit) + " levels, or contains a cycle")
}
