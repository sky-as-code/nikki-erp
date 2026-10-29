package services

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The location tree: the rules that keep it acyclic and coherent. Deriving canonical_name,
// warehouse_id and hierarchy_depth from it is inventory_location_derive.go's.

// locationTreeScanLimit caps how far a tree walk goes. It guards against a persisted cycle rather
// than being a real depth limit: looping forever inside a transaction is worse than refusing.
const locationTreeScanLimit = 100

// assertPlacementValid checks the parent a new location is put under: it exists and is usable.
// The warehouse is not the client's to choose; it follows the parent (inventory_location_derive.go).
func (this *InventoryLocationDomainServiceImpl) assertPlacementValid(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	parentId := readStringParam(params, models.InventoryLocationFieldParentLocationId)
	if parentId == "" {
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
	}
	return vErrs, nil
}

// assertMoveTargetValid checks a re-parent: the new parent exists, is usable, and is not inside the
// subtree being moved — the last stops a branch being grafted onto itself, detaching it from the
// tree in a circle. The parent may sit in another warehouse, or none: the subtree's warehouse_id
// follows it, and assertMoveKeepsReservationsBacked refuses a move that would strand a reservation.
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

// isDescendantOf walks up from candidateId to see whether ancestorId is above it. Walking up is
// bounded by tree depth, where walking down would visit the whole subtree.
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

// countUnarchivedChildren reports how many live locations sit directly under one.
func (this *InventoryLocationDomainServiceImpl) countUnarchivedChildren(
	ctx corectx.Context, parentId string,
) (int, error) {
	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldParentLocationId, dmodel.Equals, parentId),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldIsArchived, dmodel.Equals, false),
	)

	total, err := countMatching(ctx, engine, graph)
	return total, errors.Wrap(err, "countUnarchivedChildren")
}

// countOperationTypesUsing reports how many live operation types name this location as a default.
// Archiving one they still point at would leave the next transfer unable to resolve where to move
// goods.
func (this *InventoryLocationDomainServiceImpl) countOperationTypesUsing(
	ctx corectx.Context, locationId string,
) (int, error) {
	engine, err := repoFor(models.StockOperationTypeSchemaName)
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

// isOwningWarehouseUsable reports whether the location's warehouse is active and unarchived. A
// location with no warehouse — vendor, customer or shared transit — is never blocked by this.
func (this *InventoryLocationDomainServiceImpl) isOwningWarehouseUsable(
	ctx corectx.Context, location models.InventoryLocation,
) (bool, error) {
	return isWarehouseUsable(ctx, derefId(location.GetWarehouseId()))
}
