package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// Warehouse Supply Relation declares which warehouse may resupply which. It is topology, not
// execution: creating one reserves no stock, creates no quant, and starts no transfer. It is
// independent of the warehouse hierarchy — parent_warehouse_id says which warehouse system a
// warehouse belongs to, this says who may restock it, and the two need not match. It has no status
// field: available for resupply planning is exactly "not archived".
const (
	WarehouseSupplyRelationSchemaName = "inventory_warehouse_supply_relation"

	WarehouseSupplyRelationFieldId                     = basemodel.FieldId
	WarehouseSupplyRelationFieldSourceWarehouseId      = "source_warehouse_id"
	WarehouseSupplyRelationFieldDestinationWarehouseId = "destination_warehouse_id"
	WarehouseSupplyRelationFieldPriority               = "priority"
	WarehouseSupplyRelationFieldIsDefault              = "is_default"
	WarehouseSupplyRelationFieldOrgId                  = "org_id"

	WarehouseSupplyRelationEdgeSourceWarehouse      = "source_warehouse"
	WarehouseSupplyRelationEdgeDestinationWarehouse = "destination_warehouse"
)

//go:embed warehouse_supply_relation.json
var warehouseSupplyRelationSchemaJson string

func WarehouseSupplyRelationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(warehouseSupplyRelationSchemaJson)
}

type WarehouseSupplyRelation struct {
	basemodel.DynamicModelBase
}

func NewWarehouseSupplyRelation() *WarehouseSupplyRelation {
	return &WarehouseSupplyRelation{basemodel.NewDynamicModel()}
}

func NewWarehouseSupplyRelationFrom(src dmodel.DynamicFields) *WarehouseSupplyRelation {
	return &WarehouseSupplyRelation{basemodel.NewDynamicModel(src)}
}

func (this WarehouseSupplyRelation) GetSourceWarehouseId() *model.Id {
	return this.GetFieldData().GetModelId(WarehouseSupplyRelationFieldSourceWarehouseId)
}

func (this *WarehouseSupplyRelation) SetSourceWarehouseId(v *model.Id) {
	this.GetFieldData().SetModelId(WarehouseSupplyRelationFieldSourceWarehouseId, v)
}

func (this WarehouseSupplyRelation) GetDestinationWarehouseId() *model.Id {
	return this.GetFieldData().GetModelId(WarehouseSupplyRelationFieldDestinationWarehouseId)
}

func (this *WarehouseSupplyRelation) SetDestinationWarehouseId(v *model.Id) {
	this.GetFieldData().SetModelId(WarehouseSupplyRelationFieldDestinationWarehouseId, v)
}

func (this WarehouseSupplyRelation) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(WarehouseSupplyRelationFieldOrgId)
}

func (this *WarehouseSupplyRelation) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(WarehouseSupplyRelationFieldOrgId, v)
}
