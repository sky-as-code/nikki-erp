package services

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The location tree: the rules that keep it acyclic and coherent, and the path maintenance that
// follows from a change to it.

// locationPathSeparator is what complete_path joins codes with, giving 'MAIN/Stock/Zone A'.
const locationPathSeparator = "/"

// locationTreeScanLimit caps how far a tree walk will go before giving up.
//
// It is a guard against a cycle that somehow got persisted, not a real depth limit: a warehouse
// nested a hundred levels deep is already a configuration problem, and looping forever inside a
// transaction is worse than refusing.
const locationTreeScanLimit = 100

// assertPlacementValid checks where a location is being put: the warehouse, the parent, and that
// the two agree.
//
// selfId is empty on create and the location's own id on update, which is what lets the cycle
// check know whether "the parent is me" is possible.
func (this *InventoryLocationDomainServiceImpl) assertPlacementValid(
	ctx corectx.Context, params dmodel.DynamicFields, selfId string,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	parentId := readStringParam(params, models.InventoryLocationFieldParentLocationId)
	if parentId == "" {
		return vErrs, nil
	}

	if parentId == selfId {
		appendLocationViolation(vErrs, "inventory_location.hierarchy_cycle",
			"a location cannot be its own parent")
		return vErrs, nil
	}

	parent, parentErrs, err := this.loadLocation(ctx, parentId)
	if err != nil {
		return vErrs, err
	}
	if parentErrs.Count() > 0 {
		appendLocationViolation(vErrs, "inventory_location.parent_not_found",
			"no location with id '"+parentId+"'")
		return vErrs, nil
	}
	if parent.GetIsArchived() != nil && *parent.GetIsArchived() {
		appendLocationViolation(vErrs, "inventory_location.parent_archived",
			"the parent location is archived")
		return vErrs, nil
	}

	// A location and its parent must sit in the same warehouse when both are in one. Crossing
	// warehouses through the tree would make a path claim something the warehouses do not.
	warehouseId := readStringParam(params, models.InventoryLocationFieldWarehouseId)
	parentWarehouseId := derefId(parent.GetWarehouseId())
	if warehouseId != "" && parentWarehouseId != "" && warehouseId != parentWarehouseId {
		appendLocationViolation(vErrs, "inventory_location.warehouse_mismatch",
			"a location and its parent must belong to the same warehouse")
	}
	return vErrs, nil
}

// assertMoveTargetValid checks a re-parent: the new parent exists, is usable, is in the same
// warehouse, and is not inside the subtree being moved.
//
// The last rule is what stops a branch being grafted onto itself, which would detach it from the
// tree entirely and leave it pointing in a circle.
func (this *InventoryLocationDomainServiceImpl) assertMoveTargetValid(
	ctx corectx.Context, location models.InventoryLocation, newParentId string,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	locationId := derefString(location.GetId())

	if newParentId == "" {
		return vErrs, nil
	}
	if newParentId == locationId {
		appendLocationViolation(vErrs, "inventory_location.hierarchy_cycle",
			"a location cannot be its own parent")
		return vErrs, nil
	}

	parent, parentErrs, err := this.loadLocation(ctx, newParentId)
	if err != nil {
		return vErrs, err
	}
	if parentErrs.Count() > 0 {
		appendLocationViolation(vErrs, "inventory_location.parent_not_found",
			"no location with id '"+newParentId+"'")
		return vErrs, nil
	}
	if parent.GetIsArchived() != nil && *parent.GetIsArchived() {
		appendLocationViolation(vErrs, "inventory_location.parent_archived",
			"the parent location is archived")
		return vErrs, nil
	}

	if derefId(location.GetWarehouseId()) != derefId(parent.GetWarehouseId()) {
		appendLocationViolation(vErrs, "inventory_location.warehouse_mismatch",
			"a location can only move within its own warehouse")
		return vErrs, nil
	}

	inside, err := this.isDescendantOf(ctx, newParentId, locationId)
	if err != nil {
		return vErrs, err
	}
	if inside {
		appendLocationViolation(vErrs, "inventory_location.hierarchy_cycle",
			"the new parent sits underneath the location being moved")
	}
	return vErrs, nil
}

// isDescendantOf walks up from candidateId to see whether ancestorId is above it.
//
// Walking up is bounded by the depth of the tree, where walking down would visit every location in
// the subtree; for a check that only needs a yes or no, up is the cheaper direction.
func (this *InventoryLocationDomainServiceImpl) isDescendantOf(
	ctx corectx.Context, candidateId string, ancestorId string,
) (bool, error) {
	currentId := candidateId
	for hops := 0; hops < locationTreeScanLimit; hops++ {
		if currentId == "" {
			return false, nil
		}
		if currentId == ancestorId {
			return true, nil
		}
		current, vErrs, err := this.loadLocation(ctx, currentId)
		if err != nil || vErrs.Count() > 0 {
			return false, err
		}
		currentId = derefId(current.GetParentLocationId())
	}
	return false, errors.New("the location hierarchy is deeper than " +
		fmt.Sprint(locationTreeScanLimit) + " levels, or contains a cycle")
}

// rewriteSubtreePaths recomputes complete_path and hierarchy_depth for a location and everything
// beneath it.
//
// Called after a move, inside the same transaction as the re-parent, because the cached paths and
// the tree they describe have to change together.
func (this *InventoryLocationDomainServiceImpl) rewriteSubtreePaths(
	ctx corectx.Context, rootId string,
) error {
	root, vErrs, err := this.loadLocation(ctx, rootId)
	if err != nil || vErrs.Count() > 0 {
		return err
	}

	parentPath, parentDepth := "", -1
	if parentId := derefId(root.GetParentLocationId()); parentId != "" {
		parent, parentErrs, err := this.loadLocation(ctx, parentId)
		if err != nil || parentErrs.Count() > 0 {
			return err
		}
		parentPath = derefString(parent.GetCompletePath())
		parentDepth = derefInt(parent.GetHierarchyDepth())
	}

	return this.rewritePathsFrom(ctx, *root, parentPath, parentDepth+1, 0)
}

// rewritePathsFrom writes one location's derived fields, then recurses into its children.
func (this *InventoryLocationDomainServiceImpl) rewritePathsFrom(
	ctx corectx.Context, location models.InventoryLocation, parentPath string, depth int, hops int,
) error {
	if hops >= locationTreeScanLimit {
		return errors.New("the location hierarchy is deeper than " +
			fmt.Sprint(locationTreeScanLimit) + " levels, or contains a cycle")
	}

	locationId := derefString(location.GetId())
	path := joinPath(parentPath, derefString(location.GetCode()))

	if err := this.writeDerivedPath(ctx, locationId, path, depth); err != nil {
		return err
	}

	children, err := this.listChildren(ctx, locationId)
	if err != nil {
		return err
	}
	for _, child := range children {
		if err := this.rewritePathsFrom(ctx, child, path, depth+1, hops+1); err != nil {
			return err
		}
	}
	return nil
}

// listChildren returns the unarchived locations directly under one location.
//
// Archived children keep whatever path they had: they are out of the working set, and rewriting
// them would be churn nobody reads.
func (this *InventoryLocationDomainServiceImpl) listChildren(
	ctx corectx.Context, parentId string,
) ([]models.InventoryLocation, error) {
	engine, err := engineFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldParentLocationId, dmodel.Equals, parentId),
	)

	children := make([]models.InventoryLocation, 0)
	for page := 0; ; page++ {
		found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  usageScanPageSize,
		})
		if err != nil {
			return nil, errors.Wrap(err, "listChildren")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}
		for _, row := range found.Data.Items {
			children = append(children, *models.NewInventoryLocationFrom(row))
		}
		if len(found.Data.Items) < usageScanPageSize {
			break
		}
	}
	return children, nil
}

// countUnarchivedChildren reports how many live locations sit directly under one.
func (this *InventoryLocationDomainServiceImpl) countUnarchivedChildren(
	ctx corectx.Context, parentId string,
) (int, error) {
	engine, err := engineFor(models.InventoryLocationSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldParentLocationId, dmodel.Equals, parentId),
	)

	total, err := countMatching(ctx, engine, graph)
	return total, errors.Wrap(err, "countUnarchivedChildren")
}

// countOperationTypesUsing reports how many live operation types name this location as a default.
//
// Archiving a location an operation type still points at would leave the next transfer of that
// type unable to resolve where it is meant to move goods.
func (this *InventoryLocationDomainServiceImpl) countOperationTypesUsing(
	ctx corectx.Context, locationId string,
) (int, error) {
	engine, err := engineFor(models.StockOperationTypeSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().Or(
			*dmodel.NewSearchNode().NewCondition(
				models.StockOperationTypeFieldDefaultSourceLocationId, dmodel.Equals, locationId),
			*dmodel.NewSearchNode().NewCondition(
				models.StockOperationTypeFieldDefaultDestinationLocationId, dmodel.Equals, locationId),
		),
	)

	total, err := countMatching(ctx, engine, graph)
	return total, errors.Wrap(err, "countOperationTypesUsing")
}

// isOwningWarehouseUsable reports whether the location's warehouse is active and unarchived.
//
// A location with no warehouse — a vendor, customer or shared transit location — is never blocked
// by this, because there is no warehouse to be unusable.
func (this *InventoryLocationDomainServiceImpl) isOwningWarehouseUsable(
	ctx corectx.Context, location models.InventoryLocation,
) (bool, error) {
	warehouseId := derefId(location.GetWarehouseId())
	if warehouseId == "" {
		return false, nil
	}

	engine, err := engineFor(models.WarehouseSchemaName)
	if err != nil {
		return false, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.WarehouseFieldId: warehouseId,
	})
	if err != nil {
		return false, errors.Wrap(err, "isOwningWarehouseUsable")
	}
	if found == nil || !found.HasData {
		return false, nil
	}

	warehouse := models.NewWarehouseFrom(found.Data)
	archived := warehouse.GetIsArchived() != nil && *warehouse.GetIsArchived()
	return !archived && derefString(warehouse.GetStatus()) == models.WarehouseStatusActive, nil
}
