package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxMappingSchemaName = "accounting_tax_mapping"

	TaxMappingFieldId                  = basemodel.FieldId
	TaxMappingFieldOrgId               = basemodel.FieldOrgId
	TaxMappingFieldCode                = "code"
	TaxMappingFieldName                = "name"
	TaxMappingFieldVersionNo           = "version_no"
	TaxMappingFieldPriority            = "priority"
	TaxMappingFieldEffectiveFrom       = "effective_from"
	TaxMappingFieldEffectiveTo         = "effective_to"
	TaxMappingFieldSupersedesMappingId = "supersedes_mapping_id"
	TaxMappingFieldLifecycleStatus     = "lifecycle_status"
	TaxMappingFieldIsArchived          = basemodel.FieldIsArchived
)

//go:embed tax_mapping.json
var taxMappingSchemaJson string

func TaxMappingSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxMappingSchemaJson)
}

// TaxMapping substitutes one tax for another in a given context. It is never an independent
// determination engine: a mapping applies only when a rule result says apply_mapping.
type TaxMapping struct {
	basemodel.DynamicModelBase
}

func NewTaxMapping() *TaxMapping {
	return &TaxMapping{basemodel.NewDynamicModel()}
}

func NewTaxMappingFrom(src dmodel.DynamicFields) *TaxMapping {
	return &TaxMapping{basemodel.NewDynamicModel(src)}
}

func (this TaxMapping) GetCode() *string {
	return this.GetFieldData().GetString(TaxMappingFieldCode)
}

func (this *TaxMapping) SetCode(v *string) {
	this.GetFieldData().SetString(TaxMappingFieldCode, v)
}

func (this TaxMapping) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(TaxMappingFieldName)
}

func (this *TaxMapping) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(TaxMappingFieldName, v)
}

func (this TaxMapping) GetVersionNo() *int32 {
	return this.GetFieldData().GetInt32(TaxMappingFieldVersionNo)
}

func (this *TaxMapping) SetVersionNo(v *int32) {
	this.GetFieldData().SetInt32(TaxMappingFieldVersionNo, v)
}

func (this TaxMapping) GetPriority() *int32 {
	return this.GetFieldData().GetInt32(TaxMappingFieldPriority)
}

func (this *TaxMapping) SetPriority(v *int32) {
	this.GetFieldData().SetInt32(TaxMappingFieldPriority, v)
}

func (this TaxMapping) GetEffectiveFrom() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxMappingFieldEffectiveFrom)
}

func (this *TaxMapping) SetEffectiveFrom(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxMappingFieldEffectiveFrom, v)
}

func (this TaxMapping) GetEffectiveTo() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxMappingFieldEffectiveTo)
}

func (this *TaxMapping) SetEffectiveTo(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxMappingFieldEffectiveTo, v)
}

func (this TaxMapping) GetSupersedesMappingId() *model.Id {
	return this.GetFieldData().GetModelId(TaxMappingFieldSupersedesMappingId)
}

func (this *TaxMapping) SetSupersedesMappingId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxMappingFieldSupersedesMappingId, v)
}

func (this TaxMapping) GetLifecycleStatus() *string {
	return this.GetFieldData().GetString(TaxMappingFieldLifecycleStatus)
}

func (this *TaxMapping) SetLifecycleStatus(v *string) {
	this.GetFieldData().SetString(TaxMappingFieldLifecycleStatus, v)
}
