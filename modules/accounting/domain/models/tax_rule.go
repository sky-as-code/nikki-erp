package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxRuleSchemaName = "accounting_tax_rule"

	TaxRuleFieldId               = basemodel.FieldId
	TaxRuleFieldOrgId            = basemodel.FieldOrgId
	TaxRuleFieldCode             = "code"
	TaxRuleFieldName             = "name"
	TaxRuleFieldJurisdictionId   = "jurisdiction_id"
	TaxRuleFieldPriority         = "priority"
	TaxRuleFieldStopProcessing   = "stop_processing"
	TaxRuleFieldEffectiveFrom    = "effective_from"
	TaxRuleFieldEffectiveTo      = "effective_to"
	TaxRuleFieldLegalReference   = "legal_reference"
	TaxRuleFieldVersionNo        = "version_no"
	TaxRuleFieldSupersedesRuleId = "supersedes_rule_id"
	TaxRuleFieldLifecycleStatus  = "lifecycle_status"
	TaxRuleFieldIsArchived       = basemodel.FieldIsArchived
)

//go:embed tax_rule.json
var taxRuleSchemaJson string

func TaxRuleSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxRuleSchemaJson)
}

// TaxRule decides which taxes apply to a transaction context. Rules evaluate in priority order and
// nothing else: precedence is an authored decision, never inferred from how specific a rule is.
type TaxRule struct {
	basemodel.DynamicModelBase
}

func NewTaxRule() *TaxRule {
	return &TaxRule{basemodel.NewDynamicModel()}
}

func NewTaxRuleFrom(src dmodel.DynamicFields) *TaxRule {
	return &TaxRule{basemodel.NewDynamicModel(src)}
}

func (this TaxRule) GetCode() *string {
	return this.GetFieldData().GetString(TaxRuleFieldCode)
}

func (this *TaxRule) SetCode(v *string) {
	this.GetFieldData().SetString(TaxRuleFieldCode, v)
}

func (this TaxRule) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(TaxRuleFieldName)
}

func (this *TaxRule) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(TaxRuleFieldName, v)
}

func (this TaxRule) GetJurisdictionId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRuleFieldJurisdictionId)
}

func (this *TaxRule) SetJurisdictionId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRuleFieldJurisdictionId, v)
}

func (this TaxRule) GetPriority() *int32 {
	return this.GetFieldData().GetInt32(TaxRuleFieldPriority)
}

func (this *TaxRule) SetPriority(v *int32) {
	this.GetFieldData().SetInt32(TaxRuleFieldPriority, v)
}

func (this TaxRule) GetStopProcessing() *bool {
	return this.GetFieldData().GetBool(TaxRuleFieldStopProcessing)
}

func (this *TaxRule) SetStopProcessing(v *bool) {
	this.GetFieldData().SetBool(TaxRuleFieldStopProcessing, v)
}

func (this TaxRule) GetEffectiveFrom() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxRuleFieldEffectiveFrom)
}

func (this *TaxRule) SetEffectiveFrom(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxRuleFieldEffectiveFrom, v)
}

func (this TaxRule) GetEffectiveTo() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxRuleFieldEffectiveTo)
}

func (this *TaxRule) SetEffectiveTo(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxRuleFieldEffectiveTo, v)
}

func (this TaxRule) GetLegalReference() *string {
	return this.GetFieldData().GetString(TaxRuleFieldLegalReference)
}

func (this *TaxRule) SetLegalReference(v *string) {
	this.GetFieldData().SetString(TaxRuleFieldLegalReference, v)
}

func (this TaxRule) GetVersionNo() *int32 {
	return this.GetFieldData().GetInt32(TaxRuleFieldVersionNo)
}

func (this *TaxRule) SetVersionNo(v *int32) {
	this.GetFieldData().SetInt32(TaxRuleFieldVersionNo, v)
}

func (this TaxRule) GetSupersedesRuleId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRuleFieldSupersedesRuleId)
}

func (this *TaxRule) SetSupersedesRuleId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRuleFieldSupersedesRuleId, v)
}

func (this TaxRule) GetLifecycleStatus() *string {
	return this.GetFieldData().GetString(TaxRuleFieldLifecycleStatus)
}

func (this *TaxRule) SetLifecycleStatus(v *string) {
	this.GetFieldData().SetString(TaxRuleFieldLifecycleStatus, v)
}
