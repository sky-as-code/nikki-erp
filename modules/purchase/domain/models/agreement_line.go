package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	AgreementLineSchemaName = "purchase_agreement_line"

	AgreementLineFieldId                  = basemodel.FieldId
	AgreementLineFieldEtag                = basemodel.FieldEtag
	AgreementLineFieldOrgId               = basemodel.FieldOrgId
	AgreementLineFieldPurchaseAgreementId = "purchase_agreement_id"
	AgreementLineFieldSequence            = "sequence"
	AgreementLineFieldProductVariantId    = "product_variant_id"
	AgreementLineFieldUomId               = "uom_id"
	AgreementLineFieldQuantity            = "quantity"
	AgreementLineFieldUnitPrice           = "unit_price"
	AgreementLineFieldDescription         = "description"
	AgreementLineEdgeAgreement            = "agreement"
)

//go:embed agreement_line.json
var agreementLineSchemaJson string

func AgreementLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(agreementLineSchemaJson)
}

type AgreementLine struct {
	fields dmodel.DynamicFields
}

func NewAgreementLine() *AgreementLine {
	return &AgreementLine{fields: make(dmodel.DynamicFields)}
}

func NewAgreementLineFrom(src dmodel.DynamicFields) *AgreementLine {
	return &AgreementLine{fields: src}
}

func (this AgreementLine) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *AgreementLine) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this AgreementLine) GetId() *model.Id {
	return this.fields.GetModelId(AgreementLineFieldId)
}

func (this *AgreementLine) SetId(v *model.Id) {
	this.fields.SetModelId(AgreementLineFieldId, v)
}

func (this AgreementLine) GetEtag() *model.Etag {
	return this.fields.GetEtag(AgreementLineFieldEtag)
}

func (this *AgreementLine) SetEtag(v *model.Etag) {
	this.fields.SetEtag(AgreementLineFieldEtag, v)
}

func (this AgreementLine) GetOrgId() *model.Id {
	return this.fields.GetModelId(AgreementLineFieldOrgId)
}

func (this *AgreementLine) SetOrgId(v *model.Id) {
	this.fields.SetModelId(AgreementLineFieldOrgId, v)
}

func (this AgreementLine) GetPurchaseAgreementId() *model.Id {
	return this.fields.GetModelId(AgreementLineFieldPurchaseAgreementId)
}

func (this *AgreementLine) SetPurchaseAgreementId(v *model.Id) {
	this.fields.SetModelId(AgreementLineFieldPurchaseAgreementId, v)
}

func (this AgreementLine) GetSequence() *int32 {
	return this.fields.GetInt32(AgreementLineFieldSequence)
}

func (this *AgreementLine) SetSequence(v *int32) {
	this.fields.SetInt32(AgreementLineFieldSequence, v)
}

func (this AgreementLine) GetProductVariantId() *model.Id {
	return this.fields.GetModelId(AgreementLineFieldProductVariantId)
}

func (this *AgreementLine) SetProductVariantId(v *model.Id) {
	this.fields.SetModelId(AgreementLineFieldProductVariantId, v)
}

func (this AgreementLine) GetUomId() *model.Id {
	return this.fields.GetModelId(AgreementLineFieldUomId)
}

func (this *AgreementLine) SetUomId(v *model.Id) {
	this.fields.SetModelId(AgreementLineFieldUomId, v)
}

func (this AgreementLine) GetQuantity() *decimal.Decimal {
	return this.fields.GetDecimal(AgreementLineFieldQuantity)
}

func (this *AgreementLine) SetQuantity(v *decimal.Decimal) {
	this.fields.SetDecimal(AgreementLineFieldQuantity, v)
}

func (this AgreementLine) GetUnitPrice() *decimal.Decimal {
	return this.fields.GetDecimal(AgreementLineFieldUnitPrice)
}

func (this *AgreementLine) SetUnitPrice(v *decimal.Decimal) {
	this.fields.SetDecimal(AgreementLineFieldUnitPrice, v)
}

func (this AgreementLine) GetDescription() *string {
	return this.fields.GetString(AgreementLineFieldDescription)
}

func (this *AgreementLine) SetDescription(v *string) {
	this.fields.SetString(AgreementLineFieldDescription, v)
}
