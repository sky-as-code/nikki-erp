package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesChannelFulfillmentMethodSchemaName = "sales_channel_fulfillment_method"

	SalesChannelFulfillmentMethodFieldId                  = "id"
	SalesChannelFulfillmentMethodFieldOrgId               = "org_id"
	SalesChannelFulfillmentMethodFieldSalesChannelId      = "sales_channel_id"
	SalesChannelFulfillmentMethodFieldFulfillmentMethodId = "fulfillment_method_id"

	SalesChannelFulfillmentMethodEdgeSalesChannel      = "sales_channel"
	SalesChannelFulfillmentMethodEdgeFulfillmentMethod = "fulfillment_method"
)

//go:embed sales_channel_fulfillment_method.json
var salesChannelFulfillmentMethodSchemaJson string

func SalesChannelFulfillmentMethodSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesChannelFulfillmentMethodSchemaJson)
}

// SalesChannelFulfillmentMethod records that one channel may use one fulfillment method. Like the
// payment mapping beside it, the row is the state and there is no enabled flag, so a channel with no
// mappings permits no method at all — a new method is off everywhere until somebody allows it. The
// channel's own default must be one of these rows: a default outside the allowed set would be a
// channel whose every order is refused by its own configuration.
type SalesChannelFulfillmentMethod struct {
	basemodel.DynamicModelBase
}

func NewSalesChannelFulfillmentMethod() *SalesChannelFulfillmentMethod {
	return &SalesChannelFulfillmentMethod{basemodel.NewDynamicModel()}
}

func NewSalesChannelFulfillmentMethodFrom(src dmodel.DynamicFields) *SalesChannelFulfillmentMethod {
	return &SalesChannelFulfillmentMethod{basemodel.NewDynamicModel(src)}
}

func (this SalesChannelFulfillmentMethod) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesChannelFulfillmentMethod) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesChannelFulfillmentMethod) GetSalesChannelId() *model.Id {
	return this.GetFieldData().GetModelId(SalesChannelFulfillmentMethodFieldSalesChannelId)
}

func (this *SalesChannelFulfillmentMethod) SetSalesChannelId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesChannelFulfillmentMethodFieldSalesChannelId, id)
}

func (this SalesChannelFulfillmentMethod) GetFulfillmentMethodId() *model.Id {
	return this.GetFieldData().GetModelId(SalesChannelFulfillmentMethodFieldFulfillmentMethodId)
}

func (this *SalesChannelFulfillmentMethod) SetFulfillmentMethodId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesChannelFulfillmentMethodFieldFulfillmentMethodId, id)
}
