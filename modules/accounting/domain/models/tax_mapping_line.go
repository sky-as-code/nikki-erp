package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxMappingLineSchemaName = "accounting_tax_mapping_line"

	TaxMappingLineFieldId           = basemodel.FieldId
	TaxMappingLineFieldOrgId        = basemodel.FieldOrgId
	TaxMappingLineFieldTaxMappingId = "tax_mapping_id"
	TaxMappingLineFieldSourceTaxId  = "source_tax_id"
	TaxMappingLineFieldTargetTaxId  = "target_tax_id"
	TaxMappingLineFieldSequence     = "sequence"
	TaxMappingLineFieldIsArchived   = basemodel.FieldIsArchived
)

//go:embed tax_mapping_line.json
var taxMappingLineSchemaJson string

func TaxMappingLineSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxMappingLineSchemaJson)
}

// TaxMappingLine is one source-tax to target-tax substitution within a mapping.
type TaxMappingLine struct {
	basemodel.DynamicModelBase
}

func NewTaxMappingLine() *TaxMappingLine {
	return &TaxMappingLine{basemodel.NewDynamicModel()}
}

func NewTaxMappingLineFrom(src dmodel.DynamicFields) *TaxMappingLine {
	return &TaxMappingLine{basemodel.NewDynamicModel(src)}
}

func (this TaxMappingLine) GetTaxMappingId() *model.Id {
	return this.GetFieldData().GetModelId(TaxMappingLineFieldTaxMappingId)
}

func (this *TaxMappingLine) SetTaxMappingId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxMappingLineFieldTaxMappingId, v)
}

func (this TaxMappingLine) GetSourceTaxId() *model.Id {
	return this.GetFieldData().GetModelId(TaxMappingLineFieldSourceTaxId)
}

func (this *TaxMappingLine) SetSourceTaxId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxMappingLineFieldSourceTaxId, v)
}

func (this TaxMappingLine) GetTargetTaxId() *model.Id {
	return this.GetFieldData().GetModelId(TaxMappingLineFieldTargetTaxId)
}

func (this *TaxMappingLine) SetTargetTaxId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxMappingLineFieldTargetTaxId, v)
}

func (this TaxMappingLine) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(TaxMappingLineFieldSequence)
}

func (this *TaxMappingLine) SetSequence(v *int32) {
	this.GetFieldData().SetInt32(TaxMappingLineFieldSequence, v)
}
