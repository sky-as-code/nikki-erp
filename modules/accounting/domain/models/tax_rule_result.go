package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxRuleResultSchemaName = "accounting_tax_rule_result"

	TaxRuleResultFieldId           = basemodel.FieldId
	TaxRuleResultFieldOrgId        = basemodel.FieldOrgId
	TaxRuleResultFieldTaxRuleId    = "tax_rule_id"
	TaxRuleResultFieldAction       = "action"
	TaxRuleResultFieldTaxId        = "tax_id"
	TaxRuleResultFieldTaxMappingId = "tax_mapping_id"
	TaxRuleResultFieldTaxTreatment = "tax_treatment"
	TaxRuleResultFieldSequence     = "sequence"
	TaxRuleResultFieldIsArchived   = basemodel.FieldIsArchived
)

//go:embed tax_rule_result.json
var taxRuleResultSchemaJson string

func TaxRuleResultSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxRuleResultSchemaJson)
}

// TaxRuleResult is what a matching rule does to the candidate tax set.
type TaxRuleResult struct {
	basemodel.DynamicModelBase
}

func NewTaxRuleResult() *TaxRuleResult {
	return &TaxRuleResult{basemodel.NewDynamicModel()}
}

func NewTaxRuleResultFrom(src dmodel.DynamicFields) *TaxRuleResult {
	return &TaxRuleResult{basemodel.NewDynamicModel(src)}
}

func (this TaxRuleResult) GetTaxRuleId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRuleResultFieldTaxRuleId)
}

func (this *TaxRuleResult) SetTaxRuleId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRuleResultFieldTaxRuleId, v)
}

func (this TaxRuleResult) GetAction() *string {
	return this.GetFieldData().GetString(TaxRuleResultFieldAction)
}

func (this *TaxRuleResult) SetAction(v *string) {
	this.GetFieldData().SetString(TaxRuleResultFieldAction, v)
}

func (this TaxRuleResult) GetTaxId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRuleResultFieldTaxId)
}

func (this *TaxRuleResult) SetTaxId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRuleResultFieldTaxId, v)
}

func (this TaxRuleResult) GetTaxMappingId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRuleResultFieldTaxMappingId)
}

func (this *TaxRuleResult) SetTaxMappingId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRuleResultFieldTaxMappingId, v)
}

func (this TaxRuleResult) GetTaxTreatment() *string {
	return this.GetFieldData().GetString(TaxRuleResultFieldTaxTreatment)
}

func (this *TaxRuleResult) SetTaxTreatment(v *string) {
	this.GetFieldData().SetString(TaxRuleResultFieldTaxTreatment, v)
}

func (this TaxRuleResult) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(TaxRuleResultFieldSequence)
}

func (this *TaxRuleResult) SetSequence(v *int32) {
	this.GetFieldData().SetInt32(TaxRuleResultFieldSequence, v)
}
