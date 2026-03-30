package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	GroupSchemaName = "identity.group"
	GroupFieldName  = "name"
	GroupFieldDesc  = "description"
	GroupFieldUsers = "users"
)

func GroupSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(GroupSchemaName).
		Label(model.LangJson{"en-US": "User Group"}).
		TableName("ident_groups").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(GroupFieldName).
				Label(model.LangJson{"en-US": "Name"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(GroupFieldDesc).
				Label(model.LangJson{"en-US": "Description"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		EdgeTo(
			dmodel.Edge(GroupFieldUsers).
				ManyToMany(UserSchemaName, UsrGrpRelSchemaName, "group").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type Group struct {
	fields dmodel.DynamicFields
}

func NewGroup() *Group {
	return &Group{fields: make(dmodel.DynamicFields)}
}

func NewGroupFrom(src dmodel.DynamicFields) *Group {
	return &Group{fields: src}
}

func (this Group) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Group) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Group) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *Group) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this Group) GetName() *string {
	return this.fields.GetString(GroupFieldName)
}

func (this *Group) SetName(v *string) {
	this.fields.SetString(GroupFieldName, v)
}

func (this Group) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *Group) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}
