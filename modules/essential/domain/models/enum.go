package models

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	EnumSchemaName = "essential_enum"

	EnumFieldId    = basemodel.FieldId
	EnumFieldEtag  = basemodel.FieldEtag
	EnumFieldLabel = "label"
	EnumFieldValue = "value"
	EnumFieldType  = "type"
)

func EnumSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(EnumSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", EnumSchemaName)).
		TableName("essential_enums").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(EnumFieldLabel).
				Label(model.NewLangJsonRefSf("fields.%s", EnumFieldLabel)).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(EnumFieldValue).
				Label(model.NewLangJsonRefSf("fields.%s", EnumFieldValue)).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(EnumFieldType).
				Label(model.NewLangJsonRefSf("fields.%s", EnumFieldType)).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type Enum struct {
	fields dmodel.DynamicFields
}

func NewEnum() *Enum {
	return &Enum{fields: make(dmodel.DynamicFields)}
}

func NewEnumFrom(src dmodel.DynamicFields) *Enum {
	return &Enum{fields: src}
}

func (this Enum) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Enum) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Enum) GetId() *model.Id {
	return this.fields.GetModelId(EnumFieldId)
}

func (this *Enum) SetId(v *model.Id) {
	this.fields.SetModelId(EnumFieldId, v)
}

func (this Enum) GetEtag() *model.Etag {
	return this.fields.GetEtag(EnumFieldEtag)
}

func (this *Enum) SetEtag(v *model.Etag) {
	this.fields.SetEtag(EnumFieldEtag, v)
}

func (this Enum) GetLabel() *model.LangJson {
	v := this.fields.GetAny(EnumFieldLabel)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *Enum) SetLabel(v *model.LangJson) {
	if v == nil {
		this.fields.SetAny(EnumFieldLabel, nil)
		return
	}
	this.fields.SetAny(EnumFieldLabel, *v)
}

func (this Enum) GetValue() *string {
	return this.fields.GetString(EnumFieldValue)
}

func (this *Enum) SetValue(v *string) {
	this.fields.SetString(EnumFieldValue, v)
}

func (this Enum) GetType() *string {
	return this.fields.GetString(EnumFieldType)
}

func (this *Enum) SetType(v *string) {
	this.fields.SetString(EnumFieldType, v)
}
