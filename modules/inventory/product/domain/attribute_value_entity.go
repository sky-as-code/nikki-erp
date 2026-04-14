package domain

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	AttributeValueSchemaName = "inventory.attribute_value"

	AttrValFieldId           = basemodel.FieldId
	AttrValFieldAttributeId  = "attribute_id"
	AttrValFieldValueText    = "value_text"
	AttrValFieldValueInteger = "value_integer"
	AttrValFieldValueDecimal = "value_decimal"
	AttrValFieldValueBool    = "value_bool"
	AttrValFieldValueRef     = "value_ref"
	AttrValFieldVariantCount = "variant_count"

	AttrValEdgeAttribute = "attribute"
	AttrValEdgeVariants  = "variants"
)

func AttributeValueSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(AttributeValueSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Value"}).
		TableName("inventory_attribute_values").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(AttrValFieldAttributeId).
				RequiredForCreate(),
		).
		ExclusiveFields(AttrValFieldValueText, AttrValFieldValueDecimal, AttrValFieldValueInteger, AttrValFieldValueBool, AttrValFieldValueRef).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueText).
				Label(model.LangJson{model.LanguageCodeEnUs: "Text Value"}).
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueDecimal).
				Label(model.LangJson{model.LanguageCodeEnUs: "Decimal Value"}).
				DataType(dmodel.FieldDataTypeDecimal("0", fmt.Sprint(model.MODEL_RULE_CURRENCY_MAX), model.MODEL_RULE_CURRENCY_SCALE)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueInteger).
				Label(model.LangJson{model.LanguageCodeEnUs: "Decimal Value"}).
				DataType(dmodel.FieldDataTypeInt64(0, math.MaxInt64)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueBool).
				Label(model.LangJson{model.LanguageCodeEnUs: "Boolean Value"}).
				DataType(dmodel.FieldDataTypeBoolean()),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueRef).
				Label(model.LangJson{model.LanguageCodeEnUs: "Reference Value"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldVariantCount).
				Label(model.LangJson{model.LanguageCodeEnUs: "Variant Count"}).
				DataType(dmodel.FieldDataTypeInt64(0, math.MaxInt16)).
				Default(0),
		).
		EdgeTo(
			dmodel.Edge(AttrValEdgeAttribute).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attribute"}).
				ManyToOne(AttributeSchemaName, dmodel.DynamicFields{
					AttrValFieldAttributeId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(AttrValEdgeVariants).
				Label(model.LangJson{model.LanguageCodeEnUs: "Variants"}).
				ManyToMany(VariantSchemaName, VarAttrValRelSchemaName, "attribute_value").
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type AttributeValue struct {
	basemodel.DynamicModelBase
}

func NewAttributeValue() *AttributeValue {
	return &AttributeValue{basemodel.NewDynamicModel()}
}

func NewAttributeValueFrom(src dmodel.DynamicFields) *AttributeValue {
	return &AttributeValue{basemodel.NewDynamicModel(src)}
}

func (this AttributeValue) GetAttributeId() *model.Id {
	return this.GetFieldData().GetModelId(AttrValFieldAttributeId)
}

func (this *AttributeValue) SetAttributeId(v *model.Id) {
	this.GetFieldData().SetModelId(AttrValFieldAttributeId, v)
}

func (this AttributeValue) GetValueText() *model.LangJson {
	v := this.GetFieldData().GetAny(AttrValFieldValueText)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *AttributeValue) SetValueText(v *model.LangJson) {
	if v == nil {
		this.GetFieldData().SetAny(AttrValFieldValueText, nil)
		return
	}
	this.GetFieldData().SetAny(AttrValFieldValueText, *v)
}

func (this AttributeValue) GetValueDecimal() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(AttrValFieldValueDecimal)
}

func (this *AttributeValue) SetValueDecimal(v *string) {
	this.GetFieldData().SetDecimalStr(AttrValFieldValueDecimal, v)
}

func (this AttributeValue) GetValueInteger() *int64 {
	return this.GetFieldData().GetInt64(AttrValFieldValueInteger)
}

func (this *AttributeValue) SetValueInteger(v *int64) {
	this.GetFieldData().SetInt64(AttrValFieldValueInteger, v)
}

func (this AttributeValue) GetValueBool() *bool {
	return this.GetFieldData().GetBool(AttrValFieldValueBool)
}

func (this *AttributeValue) SetValueBool(v *bool) {
	this.GetFieldData().SetBool(AttrValFieldValueBool, v)
}

func (this AttributeValue) GetValueRef() *string {
	return this.GetFieldData().GetString(AttrValFieldValueRef)
}

func (this *AttributeValue) SetValueRef(v *string) {
	this.GetFieldData().SetString(AttrValFieldValueRef, v)
}

func (this AttributeValue) GetVariantCount() *int64 {
	return this.GetFieldData().GetInt64(AttrValFieldVariantCount)
}

func (this *AttributeValue) SetVariantCount(v *int64) {
	this.GetFieldData().SetInt64(AttrValFieldVariantCount, v)
}
