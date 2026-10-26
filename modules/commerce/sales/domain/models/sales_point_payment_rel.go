package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesPointPaymentRelSchemaName = "sales_point_payment_rel"

	SalesPointPaymentRelFieldId               = "id"
	SalesPointPaymentRelFieldOrgId            = "org_id"
	SalesPointPaymentRelFieldSalesPointId     = "sales_point_id"
	SalesPointPaymentRelFieldPaymentMethodId  = "payment_method_id"
	SalesPointPaymentRelFieldPaymentProfileId = "payment_profile_id"
)

//go:embed sales_point_payment_rel.json
var salesPointPaymentRelSchemaJson string

func SalesPointPaymentRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesPointPaymentRelSchemaJson)
}

type SalesPointPaymentRel struct {
	basemodel.DynamicModelBase
}

func NewSalesPointPaymentRel() *SalesPointPaymentRel {
	return &SalesPointPaymentRel{basemodel.NewDynamicModel()}
}

func NewSalesPointPaymentRelFrom(src dmodel.DynamicFields) *SalesPointPaymentRel {
	return &SalesPointPaymentRel{basemodel.NewDynamicModel(src)}
}

func (this SalesPointPaymentRel) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesPointPaymentRel) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesPointPaymentRel) GetSalesPointId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPointPaymentRelFieldSalesPointId)
}

func (this *SalesPointPaymentRel) SetSalesPointId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesPointPaymentRelFieldSalesPointId, id)
}

func (this SalesPointPaymentRel) GetPaymentMethodId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPointPaymentRelFieldPaymentMethodId)
}

func (this *SalesPointPaymentRel) SetPaymentMethodId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesPointPaymentRelFieldPaymentMethodId, id)
}

func (this SalesPointPaymentRel) GetPaymentProfileId() *model.Id {
	return this.GetFieldData().GetModelId(SalesPointPaymentRelFieldPaymentProfileId)
}

func (this *SalesPointPaymentRel) SetPaymentProfileId(id *model.Id) {
	this.GetFieldData().SetModelId(SalesPointPaymentRelFieldPaymentProfileId, id)
}
