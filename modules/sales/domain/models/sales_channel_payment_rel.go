package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesChannelPaymentRelSchemaName = "sales_channel_payment_rel"

	SalesChannelPaymentRelFieldId              = "id"
	SalesChannelPaymentRelFieldOrgId           = "org_id"
	SalesChannelPaymentRelFieldSalesChannelId  = "sales_channel_id"
	SalesChannelPaymentRelFieldPaymentMethodId = "payment_method_id"
)

//go:embed sales_channel_payment_rel.json
var salesChannelPaymentRelSchemaJson string

func SalesChannelPaymentRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesChannelPaymentRelSchemaJson)
}

// SalesChannelPaymentRel records that one channel accepts one payment method.
//
// It carries no enabled flag on purpose (CR §27): the row IS the state. A boolean would create a
// second way to express "not enabled" — a row set to false versus no row at all — and every reader
// would then have to handle both, with nothing making them agree.
//
// The consequence is the default-deny rule (CR §76): a channel with no mappings accepts no payment
// method, never all of them. A new method appearing in paymentinvoice is therefore off everywhere
// until somebody enables it (CR §77), which is the correct default for something that moves money.
type SalesChannelPaymentRel struct {
	basemodel.DynamicModelBase
}

func NewSalesChannelPaymentRel() *SalesChannelPaymentRel {
	return &SalesChannelPaymentRel{basemodel.NewDynamicModel()}
}

func NewSalesChannelPaymentRelFrom(src dmodel.DynamicFields) *SalesChannelPaymentRel {
	return &SalesChannelPaymentRel{basemodel.NewDynamicModel(src)}
}

func (this SalesChannelPaymentRel) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesChannelPaymentRel) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesChannelPaymentRel) GetSalesChannelId() *model.Id {
	return this.GetFieldData().GetModelId(SalesChannelPaymentRelFieldSalesChannelId)
}

func (this *SalesChannelPaymentRel) SetSalesChannelId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesChannelPaymentRelFieldSalesChannelId, id)
}

func (this SalesChannelPaymentRel) GetPaymentMethodId() *model.Id {
	return this.GetFieldData().GetModelId(SalesChannelPaymentRelFieldPaymentMethodId)
}

func (this *SalesChannelPaymentRel) SetPaymentMethodId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesChannelPaymentRelFieldPaymentMethodId, id)
}
