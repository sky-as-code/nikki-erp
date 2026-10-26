package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ConfigurationSchemaName = "purchase_configuration"

	ConfigurationFieldId                   = basemodel.FieldId
	ConfigurationFieldEtag                 = basemodel.FieldEtag
	ConfigurationFieldOrgId                = basemodel.FieldOrgId
	ConfigurationFieldApprovalMode         = "approval_mode"
	ConfigurationFieldApprovalThreshold    = "approval_threshold"
	ConfigurationFieldPoModificationPolicy = "po_modification_policy"
)

//go:embed configuration.json
var configurationSchemaJson string

func ConfigurationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(configurationSchemaJson)
}

type Configuration struct {
	fields dmodel.DynamicFields
}

func NewConfiguration() *Configuration {
	return &Configuration{fields: make(dmodel.DynamicFields)}
}

func NewConfigurationFrom(src dmodel.DynamicFields) *Configuration {
	return &Configuration{fields: src}
}

func (this Configuration) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Configuration) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Configuration) GetId() *model.Id {
	return this.fields.GetModelId(ConfigurationFieldId)
}

func (this *Configuration) SetId(v *model.Id) {
	this.fields.SetModelId(ConfigurationFieldId, v)
}

func (this Configuration) GetEtag() *model.Etag {
	return this.fields.GetEtag(ConfigurationFieldEtag)
}

func (this *Configuration) SetEtag(v *model.Etag) {
	this.fields.SetEtag(ConfigurationFieldEtag, v)
}

func (this Configuration) GetOrgId() *model.Id {
	return this.fields.GetModelId(ConfigurationFieldOrgId)
}

func (this *Configuration) SetOrgId(v *model.Id) {
	this.fields.SetModelId(ConfigurationFieldOrgId, v)
}

func (this Configuration) GetApprovalMode() *string {
	return this.fields.GetString(ConfigurationFieldApprovalMode)
}

func (this *Configuration) SetApprovalMode(v *string) {
	this.fields.SetString(ConfigurationFieldApprovalMode, v)
}

func (this Configuration) GetApprovalThreshold() *decimal.Decimal {
	return this.fields.GetDecimal(ConfigurationFieldApprovalThreshold)
}

func (this *Configuration) SetApprovalThreshold(v *decimal.Decimal) {
	this.fields.SetDecimal(ConfigurationFieldApprovalThreshold, v)
}

func (this Configuration) GetPoModificationPolicy() *string {
	return this.fields.GetString(ConfigurationFieldPoModificationPolicy)
}

func (this *Configuration) SetPoModificationPolicy(v *string) {
	this.fields.SetString(ConfigurationFieldPoModificationPolicy, v)
}
