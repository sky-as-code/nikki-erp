package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewInventoryLocationDomainService derives the location service from the engine's default one.
//
// usage is how the service learns what Stock holds at a location, which the archive guard needs.
// It is injected rather than looked up so the rule can be tested without a stock engine.
func NewInventoryLocationDomainService(
	base drif.DynamicResourceService, usage itStock.LocationUsageReadService,
) *InventoryLocationDomainServiceImpl {
	return &InventoryLocationDomainServiceImpl{DynamicResourceService: base, usage: usage}
}

// InventoryLocationDomainServiceImpl adds the tree rules and the lifecycle guards to the location
// resource.
//
// Location is the shared master of the whole module: Warehouse configures it and Stock references
// it. Everything here is about keeping the tree coherent and stopping a location from being
// retired while something still depends on it. None of it changes a quantity.
type InventoryLocationDomainServiceImpl struct {
	drif.DynamicResourceService

	usage itStock.LocationUsageReadService
}

var _ drif.DynamicResourceService = (*InventoryLocationDomainServiceImpl)(nil)

// Create applies the tree rules and derives the cached path.
//
// is_system_generated is stripped from whatever the client sent: only the warehouse service
// creates the locations a warehouse needs for itself, and a client able to set the flag could mint
// a location that then refuses to be archived.
func (this *InventoryLocationDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	prepared := copyFields(params)
	delete(prepared, models.InventoryLocationFieldIsSystemGenerated)

	vErrs, err := this.assertPlacementValid(ctx, prepared, "")
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}

	if err := this.fillDerivedPath(ctx, prepared); err != nil {
		return nil, err
	}
	return this.DynamicResourceService.Create(ctx, prepared)
}

// Update applies the same placement rules, and refuses to restructure a system-generated location.
//
// A warehouse's own Stock, Input or Output location may be renamed and given a storage category,
// but re-parenting it or changing what it is for would break the flow that created it.
func (this *InventoryLocationDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	locationId := readStringParam(params, models.InventoryLocationFieldId)
	current, vErrs, err := this.loadLocation(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	prepared := copyFields(params)
	// Structural fields belong to the move and lifecycle operations, which validate what a plain
	// update cannot: the subtree, the stock behind it and the paths that have to be rewritten.
	delete(prepared, models.InventoryLocationFieldIsSystemGenerated)
	delete(prepared, models.InventoryLocationFieldStatus)
	delete(prepared, models.InventoryLocationFieldCompletePath)
	delete(prepared, models.InventoryLocationFieldHierarchyDepth)

	if isSystemGenerated(*current) {
		if vErrs := assertSystemLocationUnchanged(prepared, *current); vErrs.Count() > 0 {
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
		}
	}
	if vErrs, err := this.assertPlacementValid(ctx, prepared, locationId); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	return this.DynamicResourceService.Update(ctx, prepared)
}

// SetArchived guards archiving, and leaves unarchiving to put the location back in a safe state.
//
// Archiving requires the location to be empty of stock and of work in flight, to have no
// unarchived children, and not to be a system location its warehouse still needs. Unarchiving
// always lands on suspended rather than active: the topology it sat in may have changed while it
// was away, so someone has to look before it is used again.
func (this *InventoryLocationDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	locationId := readStringParam(params, models.InventoryLocationFieldId)
	location, vErrs, err := this.loadLocation(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if readBoolParam(params, paramIsArchived) {
		if vErrs, err := this.assertArchivable(ctx, *location); err != nil {
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
	return result, this.writeStatus(ctx, locationId, models.InventoryLocationStatusSuspended)
}

// Suspend takes a location out of use while leaving everything it holds exactly where it is.
//
// Deliberately no emptiness check: suspending a rack that holds goods is the point of the
// operation. What it does check is that suspending would not strand work already under way,
// because silently unreserving or cancelling someone's move is never the right answer.
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
			"inventory_location.already_suspended", "the location is already suspended"), nil
	}

	usage, err := this.readUsage(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if usage.OpenMoveCount > 0 || usage.OpenTransferCount > 0 {
		return locationViolationResult(
			"inventory_location.has_open_operations",
			"the location is part of an operation still in progress; finish or cancel it first",
		), nil
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
			"inventory_location.already_active", "the location is already active"), nil
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

// Move re-parents a location and rewrites the cached path of everything beneath it.
//
// The whole subtree is rewritten in one transaction: a half-applied move would leave descendants
// claiming a path that no longer describes where they are.
func (this *InventoryLocationDomainServiceImpl) Move(
	ctx corectx.Context, locationId string, newParentId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	location, vErrs, err := this.loadLocation(ctx, locationId)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if isSystemGenerated(*location) {
		return locationViolationResult(
			"inventory_location.system_protected",
			"a location the warehouse created for itself cannot be moved",
		), nil
	}
	if vErrs, err := this.assertMoveTargetValid(ctx, *location, newParentId); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	err = withLocationTransaction(ctx, func(tranxCtx corectx.Context) error {
		if err := this.writeParent(tranxCtx, locationId, newParentId); err != nil {
			return err
		}
		return this.rewriteSubtreePaths(tranxCtx, locationId)
	})
	if err != nil {
		return nil, err
	}
	return locationMutateOk(), nil
}

// assertArchivable is the full archive guard: stock, work in flight, children and system status.
//
// Historical movements are deliberately not consulted. A location referenced only by completed
// moves archives cleanly, because those records keep resolving it afterwards.
func (this *InventoryLocationDomainServiceImpl) assertArchivable(
	ctx corectx.Context, location models.InventoryLocation,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	locationId := derefString(location.GetId())

	usage, err := this.readUsage(ctx, locationId)
	if err != nil {
		return vErrs, err
	}
	if !usage.OnHandQuantity.IsZero() {
		appendLocationViolation(vErrs, "inventory_location.has_on_hand_stock",
			"the location still holds stock; move it out before archiving")
	}
	if !usage.ReservedQuantity.IsZero() {
		appendLocationViolation(vErrs, "inventory_location.has_reserved_stock",
			"the location has stock reserved against it")
	}
	if usage.OpenMoveCount > 0 {
		appendLocationViolation(vErrs, "inventory_location.has_open_moves",
			"the location is part of a move that has not finished")
	}
	if usage.OpenTransferCount > 0 {
		appendLocationViolation(vErrs, "inventory_location.has_open_transfers",
			"the location is part of a transfer that has not finished")
	}

	children, err := this.countUnarchivedChildren(ctx, locationId)
	if err != nil {
		return vErrs, err
	}
	if children > 0 {
		appendLocationViolation(vErrs, "inventory_location.has_children",
			"archive or move the locations underneath this one first")
	}

	if isSystemGenerated(location) {
		warehouseUsable, err := this.isOwningWarehouseUsable(ctx, location)
		if err != nil {
			return vErrs, err
		}
		if warehouseUsable {
			appendLocationViolation(vErrs, "inventory_location.system_protected",
				"the warehouse still needs this location; archive the warehouse instead")
		}
	}

	operationTypes, err := this.countOperationTypesUsing(ctx, locationId)
	if err != nil {
		return vErrs, err
	}
	if operationTypes > 0 {
		appendLocationViolation(vErrs, "inventory_location.used_by_operation_type",
			"an operation type still uses this location as a default; change it first")
	}
	return vErrs, nil
}

// assertResumable checks that what the location hangs off is itself usable.
//
// A location cannot be more available than its warehouse or its parent: making it active under a
// suspended parent would advertise it for work that the parent has already been taken out of.
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
			appendLocationViolation(vErrs, "inventory_location.warehouse_not_usable",
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
		appendLocationViolation(vErrs, "inventory_location.parent_archived",
			"the parent location is archived")
	}
	if derefString(parent.GetStatus()) == models.InventoryLocationStatusSuspended {
		appendLocationViolation(vErrs, "inventory_location.parent_suspended",
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

// fillDerivedPath computes complete_path and hierarchy_depth from the parent.
//
// Both are caches of what the tree already says. They exist so a listing can show and sort by
// where a location sits without walking the tree for every row.
func (this *InventoryLocationDomainServiceImpl) fillDerivedPath(
	ctx corectx.Context, params dmodel.DynamicFields,
) error {
	code := readStringParam(params, models.InventoryLocationFieldCode)
	parentId := readStringParam(params, models.InventoryLocationFieldParentLocationId)

	if parentId == "" {
		params[models.InventoryLocationFieldCompletePath] = code
		params[models.InventoryLocationFieldHierarchyDepth] = 0
		return nil
	}

	parent, vErrs, err := this.loadLocation(ctx, parentId)
	if err != nil || vErrs.Count() > 0 {
		return err
	}
	params[models.InventoryLocationFieldCompletePath] = joinPath(derefString(parent.GetCompletePath()), code)
	params[models.InventoryLocationFieldHierarchyDepth] = derefInt(parent.GetHierarchyDepth()) + 1
	return nil
}

func joinPath(parentPath string, code string) string {
	if parentPath == "" {
		return code
	}
	return parentPath + locationPathSeparator + code
}

func isSystemGenerated(location models.InventoryLocation) bool {
	flag := location.GetIsSystemGenerated()
	return flag != nil && *flag
}

// assertSystemLocationUnchanged refuses the edits that would break the flow a system location
// belongs to. Renaming one, or giving it a storage category, is fine.
func assertSystemLocationUnchanged(
	params dmodel.DynamicFields, current models.InventoryLocation,
) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	protected := []string{
		models.InventoryLocationFieldCode,
		models.InventoryLocationFieldParentLocationId,
		models.InventoryLocationFieldWarehouseId,
		models.InventoryLocationFieldLocationUsage,
		models.InventoryLocationFieldPurpose,
	}

	for _, field := range protected {
		incoming, present := params[field]
		if !present {
			continue
		}
		if !strings.EqualFold(toComparableString(incoming), currentFieldString(current, field)) {
			appendLocationViolation(vErrs, "inventory_location.system_protected",
				"'"+field+"' belongs to the warehouse flow that created this location")
			return vErrs
		}
	}
	return vErrs
}
