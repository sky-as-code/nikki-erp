package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesFulfillmentTargetChangeSchemaName = "sales_fulfillment_target_change"

	SalesFulfillmentTargetChangeFieldId                    = "id"
	SalesFulfillmentTargetChangeFieldOrgId                 = "org_id"
	SalesFulfillmentTargetChangeFieldFulfillmentId         = "fulfillment_id"
	SalesFulfillmentTargetChangeFieldFromOutletId          = "from_outlet_id"
	SalesFulfillmentTargetChangeFieldToOutletId            = "to_outlet_id"
	SalesFulfillmentTargetChangeFieldInventoryOperationRef = "inventory_operation_ref"
	SalesFulfillmentTargetChangeFieldReason                = "reason"
	SalesFulfillmentTargetChangeFieldActorType             = "actor_type"
	SalesFulfillmentTargetChangeFieldActorId               = "actor_id"
	SalesFulfillmentTargetChangeFieldChangedAt             = "changed_at"
)

//go:embed sales_fulfillment_target_change.json
var salesFulfillmentTargetChangeSchemaJson string

func SalesFulfillmentTargetChangeSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesFulfillmentTargetChangeSchemaJson)
}

type SalesFulfillmentTargetChange struct {
	basemodel.DynamicModelBase
}

func NewSalesFulfillmentTargetChange() *SalesFulfillmentTargetChange {
	return &SalesFulfillmentTargetChange{basemodel.NewDynamicModel()}
}

func NewSalesFulfillmentTargetChangeFrom(src dmodel.DynamicFields) *SalesFulfillmentTargetChange {
	return &SalesFulfillmentTargetChange{basemodel.NewDynamicModel(src)}
}

func (this SalesFulfillmentTargetChange) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesFulfillmentTargetChange) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesFulfillmentTargetChange) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentTargetChangeFieldOrgId)
}

func (this *SalesFulfillmentTargetChange) SetOrgId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentTargetChangeFieldOrgId, value)
}

func (this SalesFulfillmentTargetChange) GetFulfillmentId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentTargetChangeFieldFulfillmentId)
}

func (this *SalesFulfillmentTargetChange) SetFulfillmentId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentTargetChangeFieldFulfillmentId, value)
}

func (this SalesFulfillmentTargetChange) GetFromOutletId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentTargetChangeFieldFromOutletId)
}

func (this *SalesFulfillmentTargetChange) SetFromOutletId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentTargetChangeFieldFromOutletId, value)
}

func (this SalesFulfillmentTargetChange) GetToOutletId() *model.Id {
	return this.GetFieldData().GetModelId(SalesFulfillmentTargetChangeFieldToOutletId)
}

func (this *SalesFulfillmentTargetChange) SetToOutletId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesFulfillmentTargetChangeFieldToOutletId, value)
}

func (this SalesFulfillmentTargetChange) GetInventoryOperationRef() *string {
	return this.GetFieldData().GetString(SalesFulfillmentTargetChangeFieldInventoryOperationRef)
}

func (this *SalesFulfillmentTargetChange) SetInventoryOperationRef(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentTargetChangeFieldInventoryOperationRef, value)
}

func (this SalesFulfillmentTargetChange) GetReason() *string {
	return this.GetFieldData().GetString(SalesFulfillmentTargetChangeFieldReason)
}

func (this *SalesFulfillmentTargetChange) SetReason(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentTargetChangeFieldReason, value)
}

func (this SalesFulfillmentTargetChange) GetActorType() *string {
	return this.GetFieldData().GetString(SalesFulfillmentTargetChangeFieldActorType)
}

func (this *SalesFulfillmentTargetChange) SetActorType(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentTargetChangeFieldActorType, value)
}

func (this SalesFulfillmentTargetChange) GetActorId() *string {
	return this.GetFieldData().GetString(SalesFulfillmentTargetChangeFieldActorId)
}

func (this *SalesFulfillmentTargetChange) SetActorId(value *string) {
	this.GetFieldData().SetString(SalesFulfillmentTargetChangeFieldActorId, value)
}

func (this SalesFulfillmentTargetChange) GetChangedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesFulfillmentTargetChangeFieldChangedAt)
}

func (this *SalesFulfillmentTargetChange) SetChangedAt(value *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesFulfillmentTargetChangeFieldChangedAt, value)
}
