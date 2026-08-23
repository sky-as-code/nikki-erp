package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesChannelSchemaName = "sales_channel"

	SalesChannelFieldId              = "id"
	SalesChannelFieldOrgId           = "org_id"
	SalesChannelFieldCode            = "code"
	SalesChannelFieldName            = "name"
	SalesChannelFieldDescription     = "description"
	SalesChannelFieldManagedByModule = "managed_by_module"
	SalesChannelFieldStatus          = "status"
	SalesChannelFieldIsSystem        = "is_system"

	SalesChannelEdgeSalesPoints = "sales_points"
)

// VendingChannelCode is the published integration code of the vending machine channel.
//
// It is a contract, not a configuration value: CR §41 freezes it against renaming, case changes
// and reuse. Vending code names the channel by this string and lets Sales resolve the id, so that
// no database id ever appears in a kiosk's configuration (CR §42).
const VendingChannelCode = "vdmc"

// BackOfficeChannelCode is the channel back-office sales are recorded against.
//
// It exists so that every order has a real sales point (D-19). A back-office sale has no physical
// store, but modelling that as a null sales point would leak "unknown origin" into every
// downstream query; a seeded logical point keeps the invariant true instead.
const BackOfficeChannelCode = "bo"

//go:embed sales_channel.json
var salesChannelSchemaJson string

func SalesChannelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesChannelSchemaJson)
}

type SalesChannel struct {
	basemodel.DynamicModelBase
}

func NewSalesChannel() *SalesChannel {
	return &SalesChannel{basemodel.NewDynamicModel()}
}

func NewSalesChannelFrom(src dmodel.DynamicFields) *SalesChannel {
	return &SalesChannel{basemodel.NewDynamicModel(src)}
}

func (this SalesChannel) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesChannel) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesChannel) GetCode() *string {
	return this.GetFieldData().GetString(SalesChannelFieldCode)
}

func (this *SalesChannel) SetCode(code *string) {
	this.GetFieldData().SetString(SalesChannelFieldCode, code)
}

func (this SalesChannel) GetName() *string {
	return this.GetFieldData().GetString(SalesChannelFieldName)
}

func (this *SalesChannel) SetName(name *string) {
	this.GetFieldData().SetString(SalesChannelFieldName, name)
}

func (this SalesChannel) GetDescription() *string {
	return this.GetFieldData().GetString(SalesChannelFieldDescription)
}

func (this *SalesChannel) SetDescription(description *string) {
	this.GetFieldData().SetString(SalesChannelFieldDescription, description)
}

func (this SalesChannel) GetManagedByModule() *string {
	return this.GetFieldData().GetString(SalesChannelFieldManagedByModule)
}

func (this *SalesChannel) SetManagedByModule(moduleName *string) {
	this.GetFieldData().SetString(SalesChannelFieldManagedByModule, moduleName)
}

func (this SalesChannel) GetStatus() *string {
	return this.GetFieldData().GetString(SalesChannelFieldStatus)
}

func (this *SalesChannel) SetStatus(status *string) {
	this.GetFieldData().SetString(SalesChannelFieldStatus, status)
}

func (this SalesChannel) GetIsSystem() *bool {
	return this.GetFieldData().GetBool(SalesChannelFieldIsSystem)
}

func (this *SalesChannel) SetIsSystem(isSystem *bool) {
	this.GetFieldData().SetBool(SalesChannelFieldIsSystem, isSystem)
}

func (this SalesChannel) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

// IsActive reports whether the channel may take new business.
//
// Both gates are checked together because callers always need both, and a caller that remembered
// only one would let a suspended channel keep selling or an archived one accept a new sales point.
// A nil status counts as inactive: an absent value is not a licence to trade.
func (this SalesChannel) IsActive() bool {
	status := this.GetStatus()
	if status == nil || SalesChannelStatus(*status) != SalesChannelStatusActive {
		return false
	}
	archived := this.GetIsArchived()
	return archived == nil || !*archived
}
