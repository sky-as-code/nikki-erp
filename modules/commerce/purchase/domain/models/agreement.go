package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	AgreementSchemaName = "purchase_agreement"

	AgreementFieldId            = basemodel.FieldId
	AgreementFieldEtag          = basemodel.FieldEtag
	AgreementFieldOrgId         = basemodel.FieldOrgId
	AgreementFieldIsArchived    = basemodel.FieldIsArchived
	AgreementFieldCode          = "code"
	AgreementFieldReference     = "reference"
	AgreementFieldAgreementType = "agreement_type"
	AgreementFieldStatus        = "status"
	AgreementFieldVendorId      = "vendor_id"
	AgreementFieldBuyerId       = "buyer_id"
	AgreementFieldCurrencyId    = "currency_id"
	AgreementFieldStartDate     = "start_date"
	AgreementFieldEndDate       = "end_date"
	AgreementFieldDescription   = "description"
)

//go:embed agreement.json
var agreementSchemaJson string

func AgreementSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(agreementSchemaJson)
}

type Agreement struct {
	fields dmodel.DynamicFields
}

func NewAgreement() *Agreement {
	return &Agreement{fields: make(dmodel.DynamicFields)}
}

func NewAgreementFrom(src dmodel.DynamicFields) *Agreement {
	return &Agreement{fields: src}
}

func (this Agreement) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Agreement) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Agreement) GetId() *model.Id {
	return this.fields.GetModelId(AgreementFieldId)
}

func (this *Agreement) SetId(v *model.Id) {
	this.fields.SetModelId(AgreementFieldId, v)
}

func (this Agreement) GetEtag() *model.Etag {
	return this.fields.GetEtag(AgreementFieldEtag)
}

func (this *Agreement) SetEtag(v *model.Etag) {
	this.fields.SetEtag(AgreementFieldEtag, v)
}

func (this Agreement) GetOrgId() *model.Id {
	return this.fields.GetModelId(AgreementFieldOrgId)
}

func (this *Agreement) SetOrgId(v *model.Id) {
	this.fields.SetModelId(AgreementFieldOrgId, v)
}

func (this Agreement) IsArchived() *bool {
	return this.fields.GetBool(AgreementFieldIsArchived)
}

func (this *Agreement) SetIsArchived(v *bool) {
	this.fields.SetBool(AgreementFieldIsArchived, v)
}

func (this Agreement) GetCode() *string {
	return this.fields.GetString(AgreementFieldCode)
}

func (this *Agreement) SetCode(v *string) {
	this.fields.SetString(AgreementFieldCode, v)
}

func (this Agreement) GetReference() *string {
	return this.fields.GetString(AgreementFieldReference)
}

func (this *Agreement) SetReference(v *string) {
	this.fields.SetString(AgreementFieldReference, v)
}

func (this Agreement) GetAgreementType() *string {
	return this.fields.GetString(AgreementFieldAgreementType)
}

func (this *Agreement) SetAgreementType(v *string) {
	this.fields.SetString(AgreementFieldAgreementType, v)
}

func (this Agreement) GetStatus() *string {
	return this.fields.GetString(AgreementFieldStatus)
}

func (this *Agreement) SetStatus(v *string) {
	this.fields.SetString(AgreementFieldStatus, v)
}

func (this Agreement) GetVendorId() *model.Id {
	return this.fields.GetModelId(AgreementFieldVendorId)
}

func (this *Agreement) SetVendorId(v *model.Id) {
	this.fields.SetModelId(AgreementFieldVendorId, v)
}

func (this Agreement) GetBuyerId() *model.Id {
	return this.fields.GetModelId(AgreementFieldBuyerId)
}

func (this *Agreement) SetBuyerId(v *model.Id) {
	this.fields.SetModelId(AgreementFieldBuyerId, v)
}

func (this Agreement) GetCurrencyId() *model.Id {
	return this.fields.GetModelId(AgreementFieldCurrencyId)
}

func (this *Agreement) SetCurrencyId(v *model.Id) {
	this.fields.SetModelId(AgreementFieldCurrencyId, v)
}

func (this Agreement) GetStartDate() *model.ModelDate {
	return this.fields.GetModelDate(AgreementFieldStartDate)
}

func (this *Agreement) SetStartDate(v *model.ModelDate) {
	this.fields.SetModelDate(AgreementFieldStartDate, v)
}

func (this Agreement) GetEndDate() *model.ModelDate {
	return this.fields.GetModelDate(AgreementFieldEndDate)
}

func (this *Agreement) SetEndDate(v *model.ModelDate) {
	this.fields.SetModelDate(AgreementFieldEndDate, v)
}

func (this Agreement) GetDescription() *string {
	return this.fields.GetString(AgreementFieldDescription)
}

func (this *Agreement) SetDescription(v *string) {
	this.fields.SetString(AgreementFieldDescription, v)
}
