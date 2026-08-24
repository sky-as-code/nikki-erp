package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesPointSchemaName = "sales_point"

	SalesPointFieldId                    = "id"
	SalesPointFieldOrgId                 = "org_id"
	SalesPointFieldSalesChannelId        = "sales_channel_id"
	SalesPointFieldName                  = "name"
	SalesPointFieldCode                  = "code"
	SalesPointFieldExternalReferenceId   = "external_reference_id"
	SalesPointFieldExternalReferenceType = "external_reference_type"
	SalesPointFieldStatus                = "status"

	SalesPointEdgeSalesChannel = "sales_channel"
)

// KioskReferenceType is the external_reference_type a vending kiosk's sales point carries.
//
// It names the owning module and resource so that external_reference_id is unambiguous: a bare
// ulid says nothing about which module to resolve it against, and more than one module may
// register points on the same channel.
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

// IsActive reports whether the point may take new orders.
//
// As with a channel, both gates are checked together and a nil status counts as inactive. Note
// this answers only the point's own state: an order also requires its channel to be active, which
// is a separate check the caller must not skip — a suspended channel does not cascade a status
// onto its points, because reactivating the channel would then have to guess which points were
// suspended in their own right.
func (this SalesPoint) IsActive() bool {
	status := this.GetStatus()
	if status == nil || SalesPointStatus(*status) != SalesPointStatusActive {
		return false
	}
	archived := this.GetIsArchived()
	return archived == nil || !*archived
}
