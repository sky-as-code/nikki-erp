package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// InventoryLocationUsage is the general logistics behaviour of a location, which decides whether
// stock sitting in it is company-owned. Only Internal holds real inventory; the rest are the
// counterparties and virtual locations that give every movement an opposite side, so that a
// receipt, a delivery, an adjustment and a scrap are all the same operation with different
// endpoints. See BR §4.2 and the change request §8.3.
//
// Scrap is kept distinct from InventoryLoss even though the change request's canonical list names
// only the latter: the list is a minimum ("tối thiểu") and explicitly extensible, and the two are
// different destinations here. An adjustment balances against inventory loss while a write-off
// moves to scrap, and the scrap lifecycle rejects a location of the wrong usage precisely so that
// scrapping cannot silently move usable stock somewhere it will be counted again.
const (
	InventoryLocationUsageInternal      = "internal"
	InventoryLocationUsageCustomer      = "customer"
	InventoryLocationUsageVendor        = "vendor"
	InventoryLocationUsageInventoryLoss = "inventory_loss"
	InventoryLocationUsageScrap         = "scrap"
	InventoryLocationUsageTransit       = "transit"
	InventoryLocationUsageVirtual       = "virtual"
)

// InventoryLocationPurpose is what a location is used for inside its warehouse, and is orthogonal
// to usage: WH/Input is Internal by usage and Receiving by purpose. It is null for a location
// outside any warehouse.
//
// These values are why Zone, Dock, Receiving Area and Picking Area are not separate resources.
const (
	InventoryLocationPurposeStorage   = "storage"
	InventoryLocationPurposeReceiving = "receiving"
	InventoryLocationPurposeQuality   = "quality"
	InventoryLocationPurposePicking   = "picking"
	InventoryLocationPurposePacking   = "packing"
	InventoryLocationPurposeOutput    = "output"
	InventoryLocationPurposeOther     = "other"
)

// InventoryLocationRemovalStrategy is the order goods should be taken from a location's subtree.
// It is configuration: the Stock reservation engine applies it, Warehouse only records it.
const (
	InventoryLocationRemovalStrategyFifo          = "fifo"
	InventoryLocationRemovalStrategyLifo          = "lifo"
	InventoryLocationRemovalStrategyFefo          = "fefo"
	InventoryLocationRemovalStrategyClosest       = "closest"
	InventoryLocationRemovalStrategyLeastPackages = "least_packages"
)

// InventoryLocationStatus is the operational state, separate from is_archived.
//
// Suspended is allowed while the location still holds stock — locking a location that holds goods
// is exactly what it is for — whereas archiving one that does is refused. Do not copy the guard
// from one into the other.
const (
	InventoryLocationStatusActive    = "active"
	InventoryLocationStatusSuspended = "suspended"
)

const (
	InventoryLocationSchemaName = "inventory_location"

	InventoryLocationFieldId                        = basemodel.FieldId
	InventoryLocationFieldCode                      = "code"
	InventoryLocationFieldName                      = "name"
	InventoryLocationFieldLocationUsage             = "location_usage"
	InventoryLocationFieldPurpose                   = "purpose"
	InventoryLocationFieldWarehouseId               = "warehouse_id"
	InventoryLocationFieldParentLocationId          = "parent_location_id"
	InventoryLocationFieldCompletePath              = "complete_path"
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

func (this InventoryLocation) GetCode() *string {
	return this.GetFieldData().GetString(InventoryLocationFieldCode)
}

func (this *InventoryLocation) SetCode(v *string) {
	this.GetFieldData().SetString(InventoryLocationFieldCode, v)
}

func (this InventoryLocation) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(InventoryLocationFieldName)
}

func (this *InventoryLocation) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(InventoryLocationFieldName, v)
}

func (this InventoryLocation) GetLocationUsage() *string {
	return this.GetFieldData().GetString(InventoryLocationFieldLocationUsage)
}

func (this *InventoryLocation) SetLocationUsage(v *string) {
	this.GetFieldData().SetString(InventoryLocationFieldLocationUsage, v)
}

func (this InventoryLocation) GetPurpose() *string {
	return this.GetFieldData().GetString(InventoryLocationFieldPurpose)
}

func (this *InventoryLocation) SetPurpose(v *string) {
	this.GetFieldData().SetString(InventoryLocationFieldPurpose, v)
}

// GetWarehouseId returns the owning warehouse, or nil when the location belongs to none — which is
// normal for vendor, customer, inventory-loss and shared transit locations.
func (this InventoryLocation) GetWarehouseId() *model.Id {
	return this.GetFieldData().GetModelId(InventoryLocationFieldWarehouseId)
}

func (this *InventoryLocation) SetWarehouseId(v *model.Id) {
	this.GetFieldData().SetModelId(InventoryLocationFieldWarehouseId, v)
}

func (this InventoryLocation) GetCompletePath() *string {
	return this.GetFieldData().GetString(InventoryLocationFieldCompletePath)
}

func (this *InventoryLocation) SetCompletePath(v *string) {
	this.GetFieldData().SetString(InventoryLocationFieldCompletePath, v)
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

// GetHierarchyDepth returns how deep the location sits, 0 for a root. Derived alongside
// complete_path; the tree itself is the source of truth.
func (this InventoryLocation) GetHierarchyDepth() *int32 {
	return this.GetFieldData().GetInt32(InventoryLocationFieldHierarchyDepth)
}

func (this *InventoryLocation) SetHierarchyDepth(v *int32) {
	this.GetFieldData()[InventoryLocationFieldHierarchyDepth] = v
}

// GetIsSystemGenerated reports whether the warehouse created this location for itself. Such a
// location is protected from being restructured or archived while its warehouse still needs it.
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
