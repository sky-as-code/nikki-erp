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
		TableName("inventory_attribute_groups").
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
	basemodel.DynamicModelBase
}

func NewAttributeGroup() *AttributeGroup {
	return &AttributeGroup{basemodel.NewDynamicModel()}
}

func NewAttributeGroupFrom(src dmodel.DynamicFields) *AttributeGroup {
	return &AttributeGroup{basemodel.NewDynamicModel(src)}
}

func (this AttributeGroup) GetName() *string {
	return this.GetFieldData().GetString(AttrGrpFieldName)
}

func (this *AttributeGroup) SetName(v *string) {
	this.GetFieldData().SetString(AttrGrpFieldName, v)
}

func (this AttributeGroup) GetIndex() *int64 {
	return this.GetFieldData().GetInt64(AttrGrpFieldIndex)
}

func (this *AttributeGroup) SetIndex(v *int64) {
	this.GetFieldData().SetInt64(AttrGrpFieldIndex, v)
}

func (this AttributeGroup) GetProductId() *model.Id {
	return this.GetFieldData().GetModelId(AttrGrpFieldProductId)
}

func (this *AttributeGroup) SetProductId(v *model.Id) {
	this.GetFieldData().SetModelId(AttrGrpFieldProductId, v)
}
