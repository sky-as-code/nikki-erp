package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesPointSchemaName = "sales_point"

	SalesPointFieldId                         = "id"
	SalesPointFieldOrgId                      = "org_id"
	SalesPointFieldSalesChannelId             = "sales_channel_id"
	SalesPointFieldName                       = "name"
	SalesPointFieldCode                       = "code"
	SalesPointFieldExternalReferenceId        = "external_reference_id"
	SalesPointFieldExternalReferenceType      = "external_reference_type"
	SalesPointFieldStatus                     = "status"
	SalesPointFieldDefaultFulfillmentMethodId = "default_fulfillment_method_id"
	SalesPointFieldFulfillmentEnabled         = "fulfillment_enabled"
	SalesPointFieldInventoryLocationId        = "inventory_location_id"

	SalesPointEdgeSalesChannel = "sales_channel"
)

// KioskReferenceType is the external_reference_type a vending kiosk's sales point carries. It names
// the owning module and resource so external_reference_id is unambiguous: a bare ulid says nothing
// about which module to resolve it against.
const KioskReferenceType = "vending_machine.kiosk"

//go:embed sales_point.json
var salesPointSchemaJson string

func SalesPointSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPointSchemaJson)
}

type SalesPoint struct {
	basemodel.DynamicModelBase
}

func NewSalesPoint() *SalesPoint {
	return &SalesPoint{basemodel.NewDynamicModel()}
}

func NewSalesPointFrom(src dmodel.DynamicFields) *SalesPoint {
	return &SalesPoint{basemodel.NewDynamicModel(src)}
}

func (this SalesPoint) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesPoint) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesPoint) GetSalesChannelId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPointFieldSalesChannelId)
}

func (this *SalesPoint) SetSalesChannelId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesPointFieldSalesChannelId, id)
}

func (this SalesPoint) GetName() *string {
	return this.GetFieldData().GetString(SalesPointFieldName)
}

func (this *SalesPoint) SetName(name *string) {
	this.GetFieldData().SetString(SalesPointFieldName, name)
}

func (this SalesPoint) GetCode() *string {
	return this.GetFieldData().GetString(SalesPointFieldCode)
}

func (this *SalesPoint) SetCode(code *string) {
	this.GetFieldData().SetString(SalesPointFieldCode, code)
}

func (this SalesPoint) GetExternalReferenceId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPointFieldExternalReferenceId)
}

func (this *SalesPoint) SetExternalReferenceId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesPointFieldExternalReferenceId, id)
}

func (this SalesPoint) GetExternalReferenceType() *string {
	return this.GetFieldData().GetString(SalesPointFieldExternalReferenceType)
}

func (this *SalesPoint) SetExternalReferenceType(refType *string) {
	this.GetFieldData().SetString(SalesPointFieldExternalReferenceType, refType)
}

func (this SalesPoint) GetStatus() *string {
	return this.GetFieldData().GetString(SalesPointFieldStatus)
}

func (this *SalesPoint) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesPointFieldStatus, status)
}

func (this SalesPoint) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

// IsActive reports whether the point may take new orders; a nil status counts as inactive. It
// answers only the point's own state — an order also requires its channel to be active, a separate
// check the caller must not skip, because a suspended channel does not cascade onto its points.
func (this SalesPoint) IsActive() bool {
	status := this.GetStatus()
	if status == nil || SalesPointStatus(*status) != SalesPointStatusActive {
		return false
	}
	archived := this.GetIsArchived()
	return archived == nil || !*archived
}

func (this SalesPoint) GetDefaultFulfillmentMethodId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPointFieldDefaultFulfillmentMethodId)
}

func (this *SalesPoint) SetDefaultFulfillmentMethodId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesPointFieldDefaultFulfillmentMethodId, id)
}

func (this SalesPoint) GetFulfillmentEnabled() *bool {
	return this.GetFieldData().GetBool(SalesPointFieldFulfillmentEnabled)
}

func (this *SalesPoint) SetFulfillmentEnabled(enabled *bool) {
	this.GetFieldData().SetBool(SalesPointFieldFulfillmentEnabled, enabled)
}

func (this SalesPoint) GetInventoryLocationId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPointFieldInventoryLocationId)
}

func (this *SalesPoint) SetInventoryLocationId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesPointFieldInventoryLocationId, id)
}

// CanFulfill reports whether this point may be chosen as a fulfillment target. It asks three
// questions together because a caller remembering only one would offer the customer a kiosk that
// cannot serve them: the point must be sellable at all, flagged as a fulfillment location, and
// backed by an Inventory location — a target with no location has no stock to reserve against.
func (this SalesPoint) CanFulfill() bool {
	if !this.IsActive() {
		return false
	}
	enabled := this.GetFulfillmentEnabled()
	if enabled == nil || !*enabled {
		return false
	}
	return this.GetInventoryLocationId() != nil
}
