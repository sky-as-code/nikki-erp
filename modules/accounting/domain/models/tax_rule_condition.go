package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TaxRuleConditionSchemaName = "accounting_tax_rule_condition"

	TaxRuleConditionFieldId                = basemodel.FieldId
	TaxRuleConditionFieldOrgId             = basemodel.FieldOrgId
	TaxRuleConditionFieldTaxRuleId         = "tax_rule_id"
	TaxRuleConditionFieldFieldKey          = "field_key"
	TaxRuleConditionFieldOperator          = "operator"
	TaxRuleConditionFieldValue             = "value"
	TaxRuleConditionFieldValueCurrencyCode = "value_currency_code"
	TaxRuleConditionFieldSequence          = "sequence"
	TaxRuleConditionFieldIsArchived        = basemodel.FieldIsArchived
)

//go:embed tax_rule_condition.json
var taxRuleConditionSchemaJson string

func TaxRuleConditionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(taxRuleConditionSchemaJson)
}

// TaxRuleCondition is one typed predicate on the tax context. Conditions within a rule are ANDed;
// OR is expressed by writing a second rule. Field keys come from a whitelist and values are typed
// JSON; user-supplied SQL or expressions are forbidden.
type TaxRuleCondition struct {
	basemodel.DynamicModelBase
}

func NewTaxRuleCondition() *TaxRuleCondition {
	return &TaxRuleCondition{basemodel.NewDynamicModel()}
}

func NewTaxRuleConditionFrom(src dmodel.DynamicFields) *TaxRuleCondition {
	return &TaxRuleCondition{basemodel.NewDynamicModel(src)}
}

func (this TaxRuleCondition) GetTaxRuleId() *model.Id {
	return this.GetFieldData().GetModelId(TaxRuleConditionFieldTaxRuleId)
}

func (this *TaxRuleCondition) SetTaxRuleId(v *model.Id) {
	this.GetFieldData().SetModelId(TaxRuleConditionFieldTaxRuleId, v)
}

func (this TaxRuleCondition) GetFieldKey() *string {
	return this.GetFieldData().GetString(TaxRuleConditionFieldFieldKey)
}

func (this *TaxRuleCondition) SetFieldKey(v *string) {
	this.GetFieldData().SetString(TaxRuleConditionFieldFieldKey, v)
}

func (this TaxRuleCondition) GetOperator() *string {
	return this.GetFieldData().GetString(TaxRuleConditionFieldOperator)
}

func (this *TaxRuleCondition) SetOperator(v *string) {
	this.GetFieldData().SetString(TaxRuleConditionFieldOperator, v)
}

func (this TaxRuleCondition) GetValue() any {
	return this.GetFieldData().GetAny(TaxRuleConditionFieldValue)
}

func (this *TaxRuleCondition) SetValue(v any) {
	this.GetFieldData().SetAny(TaxRuleConditionFieldValue, v)
}

func (this TaxRuleCondition) GetValueCurrencyCode() *string {
	return this.GetFieldData().GetString(TaxRuleConditionFieldValueCurrencyCode)
}

func (this *TaxRuleCondition) SetValueCurrencyCode(v *string) {
	this.GetFieldData().SetString(TaxRuleConditionFieldValueCurrencyCode, v)
}

func (this TaxRuleCondition) GetSequence() *int32 {
	return this.GetFieldData().GetInt32(TaxRuleConditionFieldSequence)
}

func (this *TaxRuleCondition) SetSequence(v *int32) {
	this.GetFieldData().SetInt32(TaxRuleConditionFieldSequence, v)
}
