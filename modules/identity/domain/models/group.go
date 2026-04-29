package models

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
	GroupEdgeOwnRoles             = "own_roles"
	GroupEdgeUsers                = "users"
	GroupEdgeBenefitGrantRequests = "benefit_grant_requests"
)

const (
	GroupAuthScope = "org"

	GroupActionCreate      = "create"
	GroupActionDelete      = "delete"
	GroupActionUpdate      = "update"
	GroupActionView        = "view"
	GroupActionManageUsers = "manage_users"
)

const (
	GrpUsrRelSchemaName = "identity.group_user_rel"

	GrpUsrRelFieldId      = basemodel.FieldId
	GrpUsrRelFieldGroupId = "group_id"
	GrpUsrRelFieldUserId  = "user_id"
)

func GroupUserRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(GrpUsrRelSchemaName).
		TableName("ident_group_user_rel").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		CompositeUnique(GrpUsrRelFieldGroupId, GrpUsrRelFieldUserId).
		Field(
			basemodel.DefineFieldId(GrpUsrRelFieldGroupId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(GrpUsrRelFieldUserId).
				RequiredForCreate(),
		)
}

func GroupSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(GroupSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "User group"}).
		TableName("ident_groups").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(GroupFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(GroupFieldDesc).
				Label(model.LangJson{model.LanguageCodeEnUs: "Description"}).
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(GroupFieldOwnerId).
				RequiredForCreate().
				Description(model.LangJson{model.LanguageCodeEnUs: "User who owns the group, is notified when membership is updated and " +
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
				ManyToMany(RoleSchemaName, RoleGroupAssignmentSchemaName, "receiver_group").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(GroupEdgeOwnRoles).
				Label(model.LangJson{model.LanguageCodeEnUs: "Owned roles"}).
				Existing(RoleSchemaName, RoleEdgeOwnerGroup),
		).
		EdgeFrom(
			dmodel.Edge(GroupEdgeBenefitGrantRequests).
				Label(model.LangJson{model.LanguageCodeEnUs: "Grant requests for this group"}).
				Existing(RoleRequestSchemaName, RoleReqEdgeReceiverGroup),
		)
}

type Group struct {
	basemodel.DynamicModelBase
}

func NewGroup() *Group {
	return &Group{basemodel.NewDynamicModel()}
}

func NewGroupFrom(src dmodel.DynamicFields) *Group {
	return &Group{basemodel.NewDynamicModel(src)}
}

func (this Group) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *Group) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this Group) GetName() *string {
	return this.GetFieldData().GetString(GroupFieldName)
}

func (this *Group) SetName(v *string) {
	this.GetFieldData().SetString(GroupFieldName, v)
}

func (this Group) GetEtag() *model.Etag {
	return this.GetFieldData().GetEtag(basemodel.FieldEtag)
}

func (this *Group) SetEtag(v *model.Etag) {
	this.GetFieldData().SetEtag(basemodel.FieldEtag, v)
}
