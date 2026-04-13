package domain

import (
	"math"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	AttributeGroupSchemaName = "inventory.attribute_group"

	AttrGrpFieldId        = basemodel.FieldId
	AttrGrpFieldName      = "name"
	AttrGrpFieldIndex     = "index"
	AttrGrpFieldProductId = "product_id"

	AttrGrpEdgeProduct    = "product"
	AttrGrpEdgeAttributes = "attributes"
)

func AttributeGroupSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(AttributeGroupSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Attribute Group"}).
		TableName("invent_attribute_groups").
		CompositeUnique(AttrGrpFieldName, AttrGrpFieldProductId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(AttrGrpFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_SHORT_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(AttrGrpFieldIndex).
				Label(model.LangJson{model.LanguageCodeEnUs: "Index"}).
				DataType(dmodel.FieldDataTypeInt64(0, math.MaxInt16)).
				Default(0),
		).
		Field(
			dmodel.DefineField().
				Name(AttrGrpFieldProductId).
				Label(model.LangJson{model.LanguageCodeEnUs: "Product"}).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(AttrGrpEdgeProduct).
				Label(model.LangJson{model.LanguageCodeEnUs: "Product"}).
				ManyToOne(ProductSchemaName, dmodel.DynamicFields{
					AttrGrpFieldProductId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(AttrGrpEdgeAttributes).
				Label(model.LangJson{model.LanguageCodeEnUs: "Attributes"}).
				Existing(AttributeSchemaName, AttrEdgeGroup),
		)
}

type AttributeGroup struct {
	fields dmodel.DynamicFields
}

func NewAttributeGroup() *AttributeGroup {
	return &AttributeGroup{fields: make(dmodel.DynamicFields)}
}

func NewAttributeGroupFrom(src dmodel.DynamicFields) *AttributeGroup {
	return &AttributeGroup{fields: src}
}

func (this AttributeGroup) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *AttributeGroup) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this AttributeGroup) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *AttributeGroup) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this AttributeGroup) GetName() *string {
	return this.fields.GetString(AttrGrpFieldName)
}

func (this *AttributeGroup) SetName(v *string) {
	this.fields.SetString(AttrGrpFieldName, v)
}

func (this AttributeGroup) GetIndex() *int64 {
	return this.fields.GetInt64(AttrGrpFieldIndex)
}

func (this *AttributeGroup) SetIndex(v *int64) {
	this.fields.SetInt64(AttrGrpFieldIndex, v)
}

func (this AttributeGroup) GetProductId() *model.Id {
	return this.fields.GetModelId(AttrGrpFieldProductId)
}

func (this *AttributeGroup) SetProductId(v *model.Id) {
	this.fields.SetModelId(AttrGrpFieldProductId, v)
}

func (this AttributeGroup) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *AttributeGroup) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}
