package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	GroupSchemaName = "identity.group"

	GroupFieldId      = basemodel.FieldId
	GroupFieldName    = "name"
	GroupFieldDesc    = "description"
	GroupFieldOwnerId = "owner_id"

	GroupEdgeOwner                = "owner"
	GroupEdgeRoles                = "roles"
	GroupEdgePrivateRole          = "private_role"
	GroupEdgeUsers                = "users"
	GroupEdgeBenefitGrantRequests = "benefit_grant_requests"
)

const (
	GrpUsrRelSchemaName = "identity.group_user_rel"

	GrpUsrRelFieldGroupId = "group_id"
	GrpUsrRelFieldUserId  = "user_id"
)

func GroupUserRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(GrpUsrRelSchemaName).
		TableName("ident_group_user_rel").
		ShouldBuildDb().
		Field(
			dmodel.DefineField().
				Name(GrpUsrRelFieldGroupId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().
				Name(GrpUsrRelFieldUserId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		)
}

func GroupSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(GroupSchemaName).
		Label(model.LangJson{"en-US": "User group"}).
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
		Field(
			dmodel.DefineField().
				Name(GroupFieldOwnerId).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate().
				Description(model.LangJson{"en-US": "User who owns the group, is notified when membership is updated and " +
					"is responsible for reviewing the membership periodically.",
				}),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(GroupEdgeOwner).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					GroupFieldOwnerId: UserFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(GroupEdgeUsers).
				ManyToMany(UserSchemaName, GrpUsrRelSchemaName, "group").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(GroupEdgeRoles).
				ManyToMany(RoleSchemaName, RoleAssignmentSchemaName, "receiver_group").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(GroupEdgePrivateRole).
				Label(model.LangJson{"en-US": "Private role"}).
				Existing(RoleSchemaName, RoleEdgeDedicatedGroup),
		).
		EdgeFrom(
			dmodel.Edge(GroupEdgeBenefitGrantRequests).
				Label(model.LangJson{"en-US": "Grant requests for this group"}).
				Existing(RoleRequestSchemaName, RoleReqEdgeReceiverGroup),
		)
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
