package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// InventoryLocationManagedByWarehouse is the managed_by_type of every location a warehouse creates for
// itself - its root, its Stock and its flow stops - and archives, unarchives and deletes together
// with itself. managed_by_id is the warehouse.
var InventoryLocationManagedByWarehouse = basemodel.ManagedByType(constants.CommerceModuleName, WarehouseSchemaName)

// InventoryLocationType is a location's logistics behaviour and decides whether stock in it is
// company-owned. Only Internal holds real inventory; Group is a node of the tree that holds none
// (a warehouse's root), and the rest are counterparties giving every movement an opposite side.
// Scrap and InventoryLoss are distinct destinations: stock found missing or damaged during a count
// balances against inventory loss, write-offs move to scrap, and the scrap lifecycle rejects a
// location of the wrong type so usable stock is not counted twice.
const (
	InventoryLocationTypeInternal      = "internal"
	InventoryLocationTypeGroup         = "group"
	InventoryLocationTypeCustomer      = "customer"
	InventoryLocationTypeVendor        = "vendor"
	InventoryLocationTypeInventoryLoss = "inventory_loss"
	InventoryLocationTypeProduction    = "production"
	InventoryLocationTypeScrap         = "scrap"
	InventoryLocationTypeTransit       = "transit"
)

// InventoryLocationRemovalStrategy is the order goods are taken from a location's subtree. The
// Stock reservation engine applies it; Warehouse only records it.
const (
	InventoryLocationRemovalStrategyFifo          = "fifo"
	InventoryLocationRemovalStrategyLifo          = "lifo"
	InventoryLocationRemovalStrategyFefo          = "fefo"
	InventoryLocationRemovalStrategyClosest       = "closest"
	InventoryLocationRemovalStrategyLeastPackages = "least_packages"
)

// InventoryLocationStatus is the operational state, separate from is_archived. Suspending a
// location holding stock is allowed; archiving one is refused. Do not share the guard.
const (
	InventoryLocationStatusActive    = "active"
	InventoryLocationStatusSuspended = "suspended"
)

const (
	InventoryLocationSchemaName = "inventory_location"

	InventoryLocationFieldId                        = basemodel.FieldId
	InventoryLocationFieldName                      = "name"
	InventoryLocationFieldCanonicalName             = "canonical_name"
	InventoryLocationFieldLocationType              = "location_type"
	InventoryLocationFieldWarehouseId               = "warehouse_id"
	InventoryLocationFieldParentLocationId          = "parent_location_id"
	InventoryLocationFieldHierarchyDepth            = "hierarchy_depth"
	InventoryLocationFieldBarcode                   = "barcode"
	InventoryLocationFieldStorageCategoryId         = "storage_category_id"
	InventoryLocationFieldRemovalStrategy           = "removal_strategy"
	InventoryLocationFieldIsReplenishmentDestConfig = "is_replenishment_destination"
	InventoryLocationFieldIsSystemGenerated         = "is_system_generated"
	InventoryLocationFieldStatus                    = "status"
	InventoryLocationFieldDescription               = "description"
	InventoryLocationFieldOrgId                     = "org_id"

	InventoryLocationEdgeParentLocation  = "parent_location"
	InventoryLocationEdgeWarehouse       = "warehouse"
	InventoryLocationEdgeStorageCategory = "storage_category"
)

//go:embed inventory_location.json
var inventoryLocationSchemaJson string

func InventoryLocationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(inventoryLocationSchemaJson)
}

type InventoryLocation struct {
	basemodel.DynamicModelBase
}

func NewInventoryLocation() *InventoryLocation {
	return &InventoryLocation{basemodel.NewDynamicModel()}
}

func NewInventoryLocationFrom(src dmodel.DynamicFields) *InventoryLocation {
	return &InventoryLocation{basemodel.NewDynamicModel(src)}
}

func (this InventoryLocation) GetName() *string {
	return this.GetFieldData().GetString(InventoryLocationFieldName)
}

func (this *InventoryLocation) SetName(v *string) {
	this.GetFieldData().SetString(InventoryLocationFieldName, v)
}

// GetCanonicalName is the name of every ancestor and then this location's own, joined by
// ProductCategoryCanonicalSeparator: 'CMR / Zone A / Rack 01'. Derived from the tree by the domain
// service, which rewrites it for the whole subtree on a rename or re-parent.
func (this InventoryLocation) GetCanonicalName() *string {
	return this.GetFieldData().GetString(InventoryLocationFieldCanonicalName)
}

func (this *InventoryLocation) SetCanonicalName(v *string) {
	this.GetFieldData().SetString(InventoryLocationFieldCanonicalName, v)
}

func (this InventoryLocation) GetLocationType() *string {
	return this.GetFieldData().GetString(InventoryLocationFieldLocationType)
}

func (this *InventoryLocation) SetLocationType(v *string) {
	this.GetFieldData().SetString(InventoryLocationFieldLocationType, v)
}

// IsGroup reports a location that only groups others and holds no stock, such as a warehouse's root.
func (this InventoryLocation) IsGroup() bool {
	locationType := this.GetLocationType()
	return locationType != nil && *locationType == InventoryLocationTypeGroup
}

// GetWarehouseId returns nil when the location belongs to no warehouse, normal for vendor,
// customer, inventory-loss and shared transit locations. Derived, never client input: a warehouse's
// root carries its warehouse, every other location inherits its parent's.
func (this InventoryLocation) GetWarehouseId() *model.Id {
	return this.GetFieldData().GetModelId(InventoryLocationFieldWarehouseId)
}

func (this *InventoryLocation) SetWarehouseId(v *model.Id) {
	this.GetFieldData().SetModelId(InventoryLocationFieldWarehouseId, v)
}

func (this InventoryLocation) GetStorageCategoryId() *model.Id {
	return this.GetFieldData().GetModelId(InventoryLocationFieldStorageCategoryId)
}

func (this *InventoryLocation) SetStorageCategoryId(v *model.Id) {
	this.GetFieldData().SetModelId(InventoryLocationFieldStorageCategoryId, v)
}

func (this InventoryLocation) GetRemovalStrategy() *string {
	return this.GetFieldData().GetString(InventoryLocationFieldRemovalStrategy)
}

func (this *InventoryLocation) SetRemovalStrategy(v *string) {
	this.GetFieldData().SetString(InventoryLocationFieldRemovalStrategy, v)
}

func (this InventoryLocation) GetStatus() *string {
	return this.GetFieldData().GetString(InventoryLocationFieldStatus)
}

func (this *InventoryLocation) SetStatus(v *string) {
	this.GetFieldData().SetString(InventoryLocationFieldStatus, v)
}

// GetParentLocationId returns the parent in the location tree, or nil for a root location.
func (this InventoryLocation) GetParentLocationId() *model.Id {
	return this.GetFieldData().GetModelId(InventoryLocationFieldParentLocationId)
}

func (this *InventoryLocation) SetParentLocationId(v *model.Id) {
	this.GetFieldData().SetModelId(InventoryLocationFieldParentLocationId, v)
}

// GetHierarchyDepth is 0 for a root. Derived alongside canonical_name; the tree is the source of
// truth.
func (this InventoryLocation) GetHierarchyDepth() *int32 {
	return this.GetFieldData().GetInt32(InventoryLocationFieldHierarchyDepth)
}

func (this *InventoryLocation) SetHierarchyDepth(v *int32) {
	this.GetFieldData()[InventoryLocationFieldHierarchyDepth] = v
}

// GetIsSystemGenerated reports a location the warehouse created for itself; it cannot be
// restructured or archived while its warehouse still needs it.
func (this InventoryLocation) GetIsSystemGenerated() *bool {
	return this.GetFieldData().GetBool(InventoryLocationFieldIsSystemGenerated)
}

func (this *InventoryLocation) SetIsSystemGenerated(v *bool) {
	this.GetFieldData()[InventoryLocationFieldIsSystemGenerated] = v
}

func (this InventoryLocation) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

func (this InventoryLocation) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(InventoryLocationFieldOrgId)
}

func (this *InventoryLocation) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(InventoryLocationFieldOrgId, v)
}

func (this InventoryLocation) GetManagedByType() *string {
	return this.GetFieldData().GetString(basemodel.FieldManagedByType)
}

func (this InventoryLocation) GetManagedById() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldManagedById)
}

// IsManagedByWarehouse reports a location a warehouse created for itself and manages: its root,
// its Stock or one of its flow stops. It lives and goes with its warehouse.
func (this InventoryLocation) IsManagedByWarehouse() bool {
	managedBy := this.GetManagedByType()
	return managedBy != nil && *managedBy == InventoryLocationManagedByWarehouse
}

// IsWarehouseRoot reports a warehouse's root location: managed by the warehouse whose id is
// managed_by_id, with no parent. Its type is fixed (group).
func (this InventoryLocation) IsWarehouseRoot() bool {
	return this.IsManagedByWarehouse() && this.GetParentLocationId() == nil
}
