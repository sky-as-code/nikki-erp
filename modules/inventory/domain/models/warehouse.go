package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// WarehouseRole says what a warehouse is for organisationally. It carries no behaviour: every role
// is configured and moved against identically, and only reporting and selection lists read it.
//
// A future Vending Machine requirement may add a role of its own. Nothing here needs to change for
// that to work, which is the point of keeping the roles behaviour-free.
const (
	WarehouseRoleCentral = "central"
	WarehouseRoleSub     = "sub"
	WarehouseRolePos     = "pos"
	WarehouseRoleOther   = "other"
)

// WarehouseFlow is how many stops goods make on the way in or out. It is policy: the Stock
// movement engine reads it to plan legs, and changing it never moves anything by itself.
const (
	WarehouseFlowOneStep   = "one_step"
	WarehouseFlowTwoStep   = "two_step"
	WarehouseFlowThreeStep = "three_step"
)

// WarehouseStatus is the operational state, and is deliberately not where archiving lives.
//
// Suspended means temporarily closed and reversible by Resume. Archived means withdrawn from the
// master data used for new work, and is carried by is_archived. There is no "archived" status: a
// warehouse is usable when status is Active AND is_archived is false, and code asking whether a
// warehouse is usable must test both.
const (
	WarehouseStatusActive    = "active"
	WarehouseStatusSuspended = "suspended"
)

const (
	WarehouseSchemaName = "inventory_warehouse"

	WarehouseFieldId                = basemodel.FieldId
	WarehouseFieldCode              = "code"
	WarehouseFieldName              = "name"
	WarehouseFieldWarehouseRole     = "warehouse_role"
	WarehouseFieldParentWarehouseId = "parent_warehouse_id"
	WarehouseFieldAddress           = "address"
	WarehouseFieldManagerUserId     = "manager_user_id"
	WarehouseFieldIncomingFlow      = "incoming_flow"
	WarehouseFieldOutgoingFlow      = "outgoing_flow"
	WarehouseFieldStatus            = "status"
	WarehouseFieldNotes             = "notes"
	WarehouseFieldOrgId             = "org_id"

	WarehouseEdgeParentWarehouse = "parent_warehouse"
)

//go:embed warehouse.json
var warehouseSchemaJson string

func WarehouseSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(warehouseSchemaJson)
}

type Warehouse struct {
	basemodel.DynamicModelBase
}

func NewWarehouse() *Warehouse {
	return &Warehouse{basemodel.NewDynamicModel()}
}

func NewWarehouseFrom(src dmodel.DynamicFields) *Warehouse {
	return &Warehouse{basemodel.NewDynamicModel(src)}
}

func (this Warehouse) GetCode() *string {
	return this.GetFieldData().GetString(WarehouseFieldCode)
}

func (this *Warehouse) SetCode(v *string) {
	this.GetFieldData().SetString(WarehouseFieldCode, v)
}

func (this Warehouse) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(WarehouseFieldName)
}

func (this *Warehouse) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(WarehouseFieldName, v)
}

func (this Warehouse) GetWarehouseRole() *string {
	return this.GetFieldData().GetString(WarehouseFieldWarehouseRole)
}

func (this *Warehouse) SetWarehouseRole(v *string) {
	this.GetFieldData().SetString(WarehouseFieldWarehouseRole, v)
}

// GetParentWarehouseId returns the parent in the warehouse hierarchy, or nil for a root warehouse.
//
// The relationship is organisational only. Stock at the parent is not stock at the child.
func (this Warehouse) GetParentWarehouseId() *model.Id {
	return this.GetFieldData().GetModelId(WarehouseFieldParentWarehouseId)
}

func (this *Warehouse) SetParentWarehouseId(v *model.Id) {
	this.GetFieldData().SetModelId(WarehouseFieldParentWarehouseId, v)
}

func (this Warehouse) GetIncomingFlow() *string {
	return this.GetFieldData().GetString(WarehouseFieldIncomingFlow)
}

func (this *Warehouse) SetIncomingFlow(v *string) {
	this.GetFieldData().SetString(WarehouseFieldIncomingFlow, v)
}

func (this Warehouse) GetOutgoingFlow() *string {
	return this.GetFieldData().GetString(WarehouseFieldOutgoingFlow)
}

func (this *Warehouse) SetOutgoingFlow(v *string) {
	this.GetFieldData().SetString(WarehouseFieldOutgoingFlow, v)
}

func (this Warehouse) GetStatus() *string {
	return this.GetFieldData().GetString(WarehouseFieldStatus)
}

func (this *Warehouse) SetStatus(v *string) {
	this.GetFieldData().SetString(WarehouseFieldStatus, v)
}

func (this Warehouse) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

func (this Warehouse) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(WarehouseFieldOrgId)
}

func (this *Warehouse) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(WarehouseFieldOrgId, v)
}
