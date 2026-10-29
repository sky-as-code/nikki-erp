package services

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// What a location inherits from where it sits in the tree: canonical_name, warehouse_id and
// hierarchy_depth. All three are derived from parent_location_id, the source of truth, and are
// rewritten for the whole subtree whenever a location is renamed or re-parented.

// locationNode is one location's derived state, as its children read it.
type locationNode struct {
	id          string
	canonical   string
	warehouseId string
	depth       int
}

// BuildLocationCanonicalName is the plain-string twin of BuildCanonicalName: the parent's canonical
// name, the separator, then the location's own name. A location without a parent is its own name.
func BuildLocationCanonicalName(parentCanonical string, name string) string {
	if parentCanonical == "" {
		return name
	}
	return parentCanonical + models.ProductCategoryCanonicalSeparator + name
}

// deriveLocationNode computes what a location inherits. A location under a parent takes the
// parent's warehouse; one without a parent belongs to no warehouse, except a warehouse's root,
// which belongs to the warehouse managing it.
func deriveLocationNode(
	location models.InventoryLocation, parent *models.InventoryLocation, name string,
) locationNode {
	node := locationNode{id: derefString(location.GetId()), canonical: name}
	if parent == nil {
		if location.IsWarehouseRoot() {
			node.warehouseId = derefId(location.GetManagedById())
		}
		return node
	}
	parentCanonical := derefString(parent.GetCanonicalName())
	if parentCanonical == "" {
		parentCanonical = derefString(parent.GetName())
	}
	node.canonical = BuildLocationCanonicalName(parentCanonical, name)
	node.warehouseId = derefId(parent.GetWarehouseId())
	node.depth = derefInt(parent.GetHierarchyDepth()) + 1
	return node
}

// applyLocationNode writes the derived fields onto a pending write.
func applyLocationNode(fields dmodel.DynamicFields, node locationNode) {
	fields[models.InventoryLocationFieldCanonicalName] = node.canonical
	fields[models.InventoryLocationFieldHierarchyDepth] = int32(node.depth)
	if node.warehouseId == "" {
		fields[models.InventoryLocationFieldWarehouseId] = nil
		return
	}
	fields[models.InventoryLocationFieldWarehouseId] = node.warehouseId
}

// rejectDerivedLocationFields refuses a command carrying canonical_name: the server derives it
// from the tree. warehouse_id needs no refusal, the schema drops it as auto-generated.
func rejectDerivedLocationFields(cmd dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	if _, sent := cmd[models.InventoryLocationFieldCanonicalName]; sent {
		vErrs.Append(*ft.NewBusinessViolation(models.InventoryLocationFieldCanonicalName,
			ft.ErrorKey("inventory_location.canonical_name_server_managed", "inventory"),
			"canonical_name is maintained by the server and cannot be set"))
	}
}

// deriveOnCreate is the create's ValidateExtra: the entity is the one inserted next, so the derived
// fields are persisted with it. It runs after schema validation, which strips the auto-generated
// fields from client input.
func (this *InventoryLocationDomainServiceImpl) deriveOnCreate(
	ctx corectx.Context, input *composable.DynamicEntity, _ *composable.DynamicEntity, _ *ft.ClientErrors,
) error {
	location := models.NewInventoryLocationFrom(input.GetFieldData())
	parent, err := this.findLocation(ctx, derefId(location.GetParentLocationId()))
	if err != nil {
		return errors.Wrap(err, "deriveOnCreate")
	}
	applyLocationNode(input.GetFieldData(), deriveLocationNode(*location, parent, derefString(location.GetName())))
	return nil
}

// deriveOnUpdate re-derives the location's own fields when its name or parent changes; the subtree
// follows in cascadeSubtree, inside the same transaction.
func (this *InventoryLocationDomainServiceImpl) deriveOnUpdate(
	ctx corectx.Context, input *composable.DynamicEntity, found *composable.DynamicEntity, _ *ft.ClientErrors,
) error {
	fields := input.GetFieldData()
	if found == nil || !changesLocationTree(fields) {
		return nil
	}
	stored := models.NewInventoryLocationFrom(found.GetFieldData())
	parentId := derefId(stored.GetParentLocationId())
	if _, sent := fields[models.InventoryLocationFieldParentLocationId]; sent {
		parentId = readStringParam(fields, models.InventoryLocationFieldParentLocationId)
	}
	name := derefString(stored.GetName())
	if _, sent := fields[models.InventoryLocationFieldName]; sent {
		name = readStringParam(fields, models.InventoryLocationFieldName)
	}

	parent, err := this.findLocation(ctx, parentId)
	if err != nil {
		return errors.Wrap(err, "deriveOnUpdate")
	}
	applyLocationNode(fields, deriveLocationNode(*stored, parent, name))
	return nil
}

func changesLocationTree(cmd dmodel.DynamicFields) bool {
	_, renamed := cmd[models.InventoryLocationFieldName]
	_, moved := cmd[models.InventoryLocationFieldParentLocationId]
	return renamed || moved
}

// findLocation reads one location by id; nil when the id is empty or matches nothing.
func (this *InventoryLocationDomainServiceImpl) findLocation(
	ctx corectx.Context, locationId string,
) (*models.InventoryLocation, error) {
	if locationId == "" {
		return nil, nil
	}
	location, vErrs, err := this.loadLocation(ctx, locationId)
	if err != nil || vErrs.Count() > 0 {
		return nil, err
	}
	return location, nil
}

// rederiveSubtree recomputes a location's derived fields from its stored parent, then cascades to
// everything beneath it. For a re-parent written straight to the repository, as Move does.
func (this *InventoryLocationDomainServiceImpl) rederiveSubtree(ctx corectx.Context, locationId string) error {
	location, err := this.findLocation(ctx, locationId)
	if err != nil || location == nil {
		return err
	}
	parent, err := this.findLocation(ctx, derefId(location.GetParentLocationId()))
	if err != nil {
		return err
	}
	fields := dmodel.DynamicFields{}
	applyLocationNode(fields, deriveLocationNode(*location, parent, derefString(location.GetName())))
	if err := this.writeLocationFields(ctx, locationId, fields); err != nil {
		return err
	}
	return this.cascadeSubtree(ctx, locationId)
}

// cascadeSubtree rewrites the derived fields of every descendant of locationId, level by level, from
// the location's stored state. Archived descendants are rewritten too: a name is a name whether or
// not the location is in use.
func (this *InventoryLocationDomainServiceImpl) cascadeSubtree(ctx corectx.Context, locationId string) error {
	start, err := this.findLocation(ctx, locationId)
	if err != nil || start == nil {
		return err
	}
	visited := map[string]bool{locationId: true}
	frontier := []locationNode{{
		id:          locationId,
		canonical:   derefString(start.GetCanonicalName()),
		warehouseId: derefId(start.GetWarehouseId()),
		depth:       derefInt(start.GetHierarchyDepth()),
	}}

	for hops := 0; len(frontier) > 0; hops++ {
		if hops >= locationTreeScanLimit {
			return errors.New("the location hierarchy is deeper than " +
				fmt.Sprint(locationTreeScanLimit) + " levels, or contains a cycle")
		}
		frontier, err = this.rewriteChildLevel(ctx, frontier, visited)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *InventoryLocationDomainServiceImpl) rewriteChildLevel(
	ctx corectx.Context, parents []locationNode, visited map[string]bool,
) ([]locationNode, error) {
	byId := make(map[string]locationNode, len(parents))
	parentIds := make([]any, 0, len(parents))
	for _, parent := range parents {
		byId[parent.id] = parent
		parentIds = append(parentIds, parent.id)
	}

	children, err := this.listChildrenOf(ctx, parentIds)
	if err != nil {
		return nil, err
	}
	next := make([]locationNode, 0, len(children))
	for _, child := range children {
		childId := derefString(child.GetId())
		if visited[childId] {
			continue
		}
		visited[childId] = true
		parent := byId[derefId(child.GetParentLocationId())]
		node := locationNode{
			id:          childId,
			canonical:   BuildLocationCanonicalName(parent.canonical, derefString(child.GetName())),
			warehouseId: parent.warehouseId,
			depth:       parent.depth + 1,
		}
		fields := dmodel.DynamicFields{}
		applyLocationNode(fields, node)
		if err := this.writeLocationFields(ctx, childId, fields); err != nil {
			return nil, err
		}
		next = append(next, node)
	}
	return next, nil
}

// listChildrenOf returns every location directly under one of parentIds, archived or not.
func (this *InventoryLocationDomainServiceImpl) listChildrenOf(
	ctx corectx.Context, parentIds []any,
) ([]models.InventoryLocation, error) {
	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(
		models.InventoryLocationFieldParentLocationId, dmodel.In, parentIds...))

	children := make([]models.InventoryLocation, 0)
	for page := 0; ; page++ {
		found, err := engine.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: page, Size: usageScanPageSize})
		if err != nil {
			return nil, errors.Wrap(err, "listChildrenOf")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			return children, nil
		}
		for _, row := range found.Data.Items {
			children = append(children, *models.NewInventoryLocationFrom(row))
		}
		if len(found.Data.Items) < usageScanPageSize {
			return children, nil
		}
	}
}
