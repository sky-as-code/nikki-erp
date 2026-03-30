package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	HierarchyLevelSchemaName = "identity.hierarchy_level"
	HierFieldName            = "name"
	HierFieldParentId        = "parent_id"
	HierFieldOrgId           = "org_id"
	HierFieldUsers           = "users"

	HierEdgeChildren = "children"
	HierEdgeParent   = "parent"
	HierEdgeOrg      = "org"
)

func HierarchyLevelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(HierarchyLevelSchemaName).
		Label(model.LangJson{"en-US": "Hierarchy Level"}).
		TableName("ident_hierarchy_levels").
		PartialUnique(HierFieldName, HierFieldOrgId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(HierFieldName).
				Label(model.LangJson{"en-US": "Name"}).
				DataType(dmodel.FieldDataTypeString(1, 50)).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(HierFieldParentId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().
				Name(HierFieldOrgId).
				Label(model.LangJson{"en-US": "Organization"}).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		EdgeTo(
			dmodel.Edge(HierEdgeParent).
				Label(model.LangJson{"en-US": "Parent Level"}).
				ManyToOne(HierarchyLevelSchemaName, dmodel.DynamicFields{
					HierFieldParentId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(HierEdgeOrg).
				Label(model.LangJson{"en-US": "Org"}).
				ManyToOne(OrganizationSchemaName, dmodel.DynamicFields{
					HierFieldOrgId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(HierFieldUsers).
				Label(model.LangJson{"en-US": "Users"}).
				Existing(UserSchemaName, UserEdgeHierarchy),
		).
		EdgeFrom(
			dmodel.Edge(HierEdgeChildren).
				Label(model.LangJson{"en-US": "Children Levels"}).
				Existing(HierarchyLevelSchemaName, HierEdgeParent),
		)
}

type HierarchyLevel struct {
	fields dmodel.DynamicFields
}

func NewHierarchyLevel() *HierarchyLevel {
	return &HierarchyLevel{fields: make(dmodel.DynamicFields)}
}

func NewHierarchyLevelFrom(src dmodel.DynamicFields) *HierarchyLevel {
	return &HierarchyLevel{fields: src}
}

func (this HierarchyLevel) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *HierarchyLevel) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this HierarchyLevel) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *HierarchyLevel) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this HierarchyLevel) GetName() *string {
	return this.fields.GetString(HierFieldName)
}

func (this *HierarchyLevel) SetName(v *string) {
	this.fields.SetString(HierFieldName, v)
}

func (this HierarchyLevel) GetOrgId() *model.Id {
	return this.fields.GetModelId(HierFieldOrgId)
}

func (this *HierarchyLevel) SetOrgId(v *model.Id) {
	this.fields.SetModelId(HierFieldOrgId, v)
}

func (this HierarchyLevel) GetParentId() *model.Id {
	return this.fields.GetModelId(HierFieldParentId)
}

func (this *HierarchyLevel) SetParentId(v *model.Id) {
	this.fields.SetModelId(HierFieldParentId, v)
}

func (this HierarchyLevel) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *HierarchyLevel) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}
