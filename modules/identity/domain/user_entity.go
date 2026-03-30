package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type UserStatus string

const (
	UserStatusDraft      = UserStatus("draft")
	UserStatusInvited    = UserStatus("invited")
	UserStatusActive     = UserStatus("active")
	UserStatusSuspended  = UserStatus("suspended")
	UserStatusTerminated = UserStatus("terminated")
)

func (this UserStatus) String() string {
	return string(this)
}

func WrapUserStatus(s string) *UserStatus {
	st := UserStatus(s)
	return &st
}

const (
	UserSchemaName       = "identity.user"
	UserFieldAvatarUrl   = "avatar_url"
	UserFieldDisplayName = "display_name"
	UserFieldEmail       = "email"
	UserFieldId          = basemodel.FieldId
	UserFieldIsOwner     = "is_owner"
	UserFieldIsLocked    = "is_locked"
	UserFieldHierarchyId = "hierarchy_id"
	UserFieldStatus      = "status"

	UserEdgeGroups    = "groups"
	UserEdgeOrgs      = "orgs"
	UserEdgeHierarchy = "hierarchy"
)

const (
	UsrGrpRelSchemaName   = "identity.user_group_rel"
	UsrGrpRelFieldUserId  = "user_id"
	UsrGrpRelFieldGroupId = "group_id"
)

func UserGroupRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(UsrGrpRelSchemaName).
		TableName("ident_user_group_rel").
		ShouldBuildDb().
		Field(
			dmodel.DefineField().
				Name(UsrGrpRelFieldUserId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().
				Name(UsrGrpRelFieldGroupId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		)
}

func UserSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(UserSchemaName).
		Label(model.LangJson{"en-US": "User"}).
		TableName("ident_users").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(UserFieldAvatarUrl).
				Label(model.LangJson{"en-US": "Avatar URL"}).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Field(
			dmodel.DefineField().
				Name(UserFieldDisplayName).
				Label(model.LangJson{"en-US": "Display Name"}).
				DataType(dmodel.FieldDataTypeString(3, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(UserFieldEmail).
				Label(model.LangJson{"en-US": "Email"}).
				DataType(dmodel.FieldDataTypeEmail()).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(UserFieldStatus).
				Label(model.LangJson{"en-US": "Status"}).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(UserStatusDraft), string(UserStatusInvited), string(UserStatusActive), string(UserStatusSuspended), string(UserStatusTerminated),
				})).
				RequiredForCreate().
				Default(string(UserStatusDraft)),
		).
		Field(
			dmodel.DefineField().
				Name(UserFieldIsOwner).
				Label(model.LangJson{"en-US": "Is Owner"}).
				DataType(dmodel.FieldDataTypeBoolean()).
				Unique(), // Only one owner per deployment
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(UserFieldHierarchyId).
				Label(model.LangJson{"en-US": "Hierarchy Level"}).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		EdgeTo(
			dmodel.Edge(UserEdgeHierarchy).
				ManyToOne(HierarchyLevelSchemaName, dmodel.DynamicFields{
					UserFieldHierarchyId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeSetNull),
		).
		EdgeTo(
			dmodel.Edge(UserEdgeGroups).
				ManyToMany(GroupSchemaName, UsrGrpRelSchemaName, "user").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserEdgeOrgs).
				ManyToMany(OrganizationSchemaName, UsrOrgRelSchemaName, "user").
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type User struct {
	fields dmodel.DynamicFields
}

func NewUser() *User {
	return &User{fields: make(dmodel.DynamicFields)}
}

func NewUserFrom(src dmodel.DynamicFields) *User {
	return &User{fields: src}
}

func (this User) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *User) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this User) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *User) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this User) IsArchived() bool {
	val := this.fields.GetBool(basemodel.FieldIsArchived)
	if val == nil {
		return false
	}
	return *val
}

func (this *User) SetIsArchived(v *bool) {
	this.fields.SetBool(basemodel.FieldIsArchived, v)
}

func (this User) GetAvatarUrl() *string {
	return this.fields.GetString(UserFieldAvatarUrl)
}

func (this *User) SetAvatarUrl(v *string) {
	this.fields.SetString(UserFieldAvatarUrl, v)
}

func (this User) GetDisplayName() *string {
	return this.fields.GetString(UserFieldDisplayName)
}

func (this *User) SetDisplayName(v *string) {
	this.fields.SetString(UserFieldDisplayName, v)
}

func (this User) GetEmail() *string {
	return this.fields.GetString(UserFieldEmail)
}

func (this *User) SetEmail(v *string) {
	this.fields.SetString(UserFieldEmail, v)
}

func (this User) IsOwner() bool {
	val := this.fields.GetBool(UserFieldIsOwner)
	if val == nil {
		return false
	}
	return *val
}

func (this *User) SetIsOwner(v *bool) {
	this.fields.SetBool(UserFieldIsOwner, v)
}

func (this User) GetStatus() *UserStatus {
	s := this.fields.GetString(UserFieldStatus)
	if s == nil {
		return nil
	}
	st := UserStatus(*s)
	return &st
}

func (this *User) SetStatus(v *UserStatus) {
	if v == nil {
		this.fields.SetString(UserFieldStatus, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(UserFieldStatus, &s)
}

func (this User) GetHierarchyId() *model.Id {
	return this.fields.GetModelId(UserFieldHierarchyId)
}

func (this *User) SetHierarchyId(v *model.Id) {
	this.fields.SetModelId(UserFieldHierarchyId, v)
}
