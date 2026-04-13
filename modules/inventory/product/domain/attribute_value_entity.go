package domain

import (
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
	AttrValFieldValueNumber  = "value_number"
	AttrValFieldValueBool    = "value_bool"
	AttrValFieldValueRef     = "value_ref"
	AttrValFieldVariantCount = "variant_count"

	AttrValEdgeAttribute = "attribute"
	AttrValEdgeVariants  = "variants"
)

func AttributeValueSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(AttributeValueSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Value"}).
		TableName("invent_attribute_values").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldAttributeId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attribute"}).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueText).
				Label(model.LangJson{model.LanguageCodeEnUs: "Text Value"}).
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(AttrValFieldValueNumber).
				Label(model.LangJson{model.LanguageCodeEnUs: "Number Value"}).
				DataType(dmodel.FieldDataTypeDecimal("0", "9999999999.9999", 4)),
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
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
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
	fields dmodel.DynamicFields
}

func NewAttributeValue() *AttributeValue {
	return &AttributeValue{fields: make(dmodel.DynamicFields)}
}

func NewAttributeValueFrom(src dmodel.DynamicFields) *AttributeValue {
	return &AttributeValue{fields: src}
}

func (this AttributeValue) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *AttributeValue) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this AttributeValue) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *AttributeValue) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this AttributeValue) GetAttributeId() *model.Id {
	return this.fields.GetModelId(AttrValFieldAttributeId)
}

func (this *AttributeValue) SetAttributeId(v *model.Id) {
	this.fields.SetModelId(AttrValFieldAttributeId, v)
}

func (this AttributeValue) GetValueText() *model.LangJson {
	v := this.fields.GetAny(AttrValFieldValueText)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *AttributeValue) SetValueText(v *model.LangJson) {
	if v == nil {
		this.fields.SetAny(AttrValFieldValueText, nil)
		return
	}
	this.fields.SetAny(AttrValFieldValueText, *v)
}

func (this AttributeValue) GetValueNumber() *decimal.Decimal {
	v := this.fields.GetAny(AttrValFieldValueNumber)
	if v == nil {
		return nil
	}
	d := v.(decimal.Decimal)
	return &d
}

func (this *AttributeValue) SetValueNumber(v *decimal.Decimal) {
	if v == nil {
		this.fields.SetAny(AttrValFieldValueNumber, nil)
		return
	}
	this.fields.SetAny(AttrValFieldValueNumber, *v)
}

func (this AttributeValue) GetValueBool() *bool {
	return this.fields.GetBool(AttrValFieldValueBool)
}

func (this *AttributeValue) SetValueBool(v *bool) {
	this.fields.SetBool(AttrValFieldValueBool, v)
}

func (this AttributeValue) GetValueRef() *string {
	return this.fields.GetString(AttrValFieldValueRef)
}

func (this *AttributeValue) SetValueRef(v *string) {
	this.fields.SetString(AttrValFieldValueRef, v)
}

func (this AttributeValue) GetVariantCount() *int64 {
	return this.fields.GetInt64(AttrValFieldVariantCount)
}

func (this *AttributeValue) SetVariantCount(v *int64) {
	this.fields.SetInt64(AttrValFieldVariantCount, v)
}

func (this AttributeValue) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *AttributeValue) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}
