package models

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TagSchemaName = "essential_tag"

	TagFieldId    = basemodel.FieldId
	TagFieldEtag  = basemodel.FieldEtag
	TagFieldLabel = "label"
	TagFieldType  = "type"
)

func TagSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TagSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", TagSchemaName)).
		TableName("essential_tags").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(TagFieldLabel).
				Label(model.NewLangJsonRefSf("fields.%s", TagFieldLabel)).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(TagFieldType).
				Label(model.NewLangJsonRefSf("fields.%s", TagFieldType)).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type Tag struct {
	fields dmodel.DynamicFields
}

func NewTag() *Tag {
	return &Tag{fields: make(dmodel.DynamicFields)}
}

func NewTagFrom(src dmodel.DynamicFields) *Tag {
	return &Tag{fields: src}
}

func (this Tag) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Tag) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Tag) GetId() *model.Id {
	return this.fields.GetModelId(TagFieldId)
}

func (this *Tag) SetId(v *model.Id) {
	this.fields.SetModelId(TagFieldId, v)
}

func (this Tag) GetEtag() *model.Etag {
	return this.fields.GetEtag(TagFieldEtag)
}

func (this *Tag) SetEtag(v *model.Etag) {
	this.fields.SetEtag(TagFieldEtag, v)
}

func (this Tag) GetLabel() *model.LangJson {
	v := this.fields.GetAny(TagFieldLabel)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *Tag) SetLabel(v *model.LangJson) {
	if v == nil {
		this.fields.SetAny(TagFieldLabel, nil)
		return
	}
	this.fields.SetAny(TagFieldLabel, *v)
}

func (this Tag) GetType() *string {
	return this.fields.GetString(TagFieldType)
}

func (this *Tag) SetType(v *string) {
	this.fields.SetString(TagFieldType, v)
}
