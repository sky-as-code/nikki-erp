package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxRoundingPolicySchemaName = "accounting_tax_rounding_policy"

	TaxRoundingPolicyFieldId                 = basemodel.FieldId
	TaxRoundingPolicyFieldOrgId              = basemodel.FieldOrgId
	TaxRoundingPolicyFieldCode               = "code"
	TaxRoundingPolicyFieldName               = "name"
	TaxRoundingPolicyFieldJurisdictionId     = "jurisdiction_id"
	TaxRoundingPolicyFieldCurrencyCode       = "currency_code"
	TaxRoundingPolicyFieldRoundingScope      = "rounding_scope"
	TaxRoundingPolicyFieldRoundingMethod     = "rounding_method"
	TaxRoundingPolicyFieldRoundingIncrement  = "rounding_increment"
	TaxRoundingPolicyFieldPrecision          = "precision"
	TaxRoundingPolicyFieldVersionNo          = "version_no"
	TaxRoundingPolicyFieldEffectiveFrom      = "effective_from"
	TaxRoundingPolicyFieldEffectiveTo        = "effective_to"
	TaxRoundingPolicyFieldSupersedesPolicyId = "supersedes_policy_id"
	TaxRoundingPolicyFieldLifecycleStatus    = "lifecycle_status"
	TaxRoundingPolicyFieldIsArchived         = basemodel.FieldIsArchived
)

//go:embed tax_rounding_policy.json
var taxRoundingPolicySchemaJson string

func TaxRoundingPolicySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxRoundingPolicySchemaJson)
}

// TaxRoundingPolicy decides how a computed tax amount is rounded, and at which scope.
//
// Effective-dated and versioned like every other tax configuration, because a rounding rule is
// as much a legal choice as a rate: changing it retroactively would alter what past documents
// should have charged.
type TaxRoundingPolicy struct {
	basemodel.DynamicModelBase
}

func NewTaxRoundingPolicy() *TaxRoundingPolicy {
	return &TaxRoundingPolicy{basemodel.NewDynamicModel()}
}

func NewTaxRoundingPolicyFrom(src dmodel.DynamicFields) *TaxRoundingPolicy {
	return &TaxRoundingPolicy{basemodel.NewDynamicModel(src)}
}

func (this TaxRoundingPolicy) GetCode() *string {
	return this.GetFieldData().GetString(TaxRoundingPolicyFieldCode)
}

func (this *TaxRoundingPolicy) SetCode(v *string) {
	this.GetFieldData().SetString(TaxRoundingPolicyFieldCode, v)
}

func (this TaxRoundingPolicy) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(TaxRoundingPolicyFieldName)
}

func (this *TaxRoundingPolicy) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(TaxRoundingPolicyFieldName, v)
}

func (this TaxRoundingPolicy) GetJurisdictionId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRoundingPolicyFieldJurisdictionId)
}

func (this *TaxRoundingPolicy) SetJurisdictionId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRoundingPolicyFieldJurisdictionId, v)
}

func (this TaxRoundingPolicy) GetCurrencyCode() *string {
	return this.GetFieldData().GetString(TaxRoundingPolicyFieldCurrencyCode)
}

func (this *TaxRoundingPolicy) SetCurrencyCode(v *string) {
	this.GetFieldData().SetString(TaxRoundingPolicyFieldCurrencyCode, v)
}

func (this TaxRoundingPolicy) GetRoundingScope() *string {
	return this.GetFieldData().GetString(TaxRoundingPolicyFieldRoundingScope)
}

func (this *TaxRoundingPolicy) SetRoundingScope(v *string) {
	this.GetFieldData().SetString(TaxRoundingPolicyFieldRoundingScope, v)
}

func (this TaxRoundingPolicy) GetRoundingMethod() *string {
	return this.GetFieldData().GetString(TaxRoundingPolicyFieldRoundingMethod)
}

func (this *TaxRoundingPolicy) SetRoundingMethod(v *string) {
	this.GetFieldData().SetString(TaxRoundingPolicyFieldRoundingMethod, v)
}

func (this TaxRoundingPolicy) GetRoundingIncrement() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(TaxRoundingPolicyFieldRoundingIncrement)
}

func (this *TaxRoundingPolicy) SetRoundingIncrement(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(TaxRoundingPolicyFieldRoundingIncrement, v)
}

func (this TaxRoundingPolicy) GetPrecision() *int32 {
	return this.GetFieldData().GetInt32(TaxRoundingPolicyFieldPrecision)
}

func (this *TaxRoundingPolicy) SetPrecision(v *int32) {
	this.GetFieldData().SetInt32(TaxRoundingPolicyFieldPrecision, v)
}

func (this TaxRoundingPolicy) GetVersionNo() *int32 {
	return this.GetFieldData().GetInt32(TaxRoundingPolicyFieldVersionNo)
}

func (this *TaxRoundingPolicy) SetVersionNo(v *int32) {
	this.GetFieldData().SetInt32(TaxRoundingPolicyFieldVersionNo, v)
}

func (this TaxRoundingPolicy) GetEffectiveFrom() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxRoundingPolicyFieldEffectiveFrom)
}

func (this *TaxRoundingPolicy) SetEffectiveFrom(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxRoundingPolicyFieldEffectiveFrom, v)
}

func (this TaxRoundingPolicy) GetEffectiveTo() *model.ModelDate {
	return this.GetFieldData().GetModelDate(TaxRoundingPolicyFieldEffectiveTo)
}

func (this *TaxRoundingPolicy) SetEffectiveTo(v *model.ModelDate) {
	this.GetFieldData().SetModelDate(TaxRoundingPolicyFieldEffectiveTo, v)
}

func (this TaxRoundingPolicy) GetSupersedesPolicyId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRoundingPolicyFieldSupersedesPolicyId)
}

func (this *TaxRoundingPolicy) SetSupersedesPolicyId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRoundingPolicyFieldSupersedesPolicyId, v)
}

func (this TaxRoundingPolicy) GetLifecycleStatus() *string {
	return this.GetFieldData().GetString(TaxRoundingPolicyFieldLifecycleStatus)
}

func (this *TaxRoundingPolicy) SetLifecycleStatus(v *string) {
	this.GetFieldData().SetString(TaxRoundingPolicyFieldLifecycleStatus, v)
}
