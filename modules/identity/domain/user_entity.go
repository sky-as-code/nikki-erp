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

const (
	UserSchemaName = "identity.user"

	UserFieldId          = basemodel.FieldId
	UserFieldAvatarUrl   = "avatar_url"
	UserFieldDisplayName = "display_name"
	UserFieldEmail       = "email"
	UserFieldIsOwner     = "is_owner"
	UserFieldIsLocked    = "is_locked"
	UserFieldOrgUnitId   = "org_unit_id"
	UserFieldStatus      = "status"

	UserEdgeGroups                = "groups"
	UserEdgeOrgs                  = "orgs"
	UserEdgeOrgUnit               = "org_unit"
	UserEdgeRoles                 = "roles"
	UserEdgePrivateRole           = "private_role"
	UserEdgeBenefitRoleRequests   = "benefit_role_requests"
	UserEdgeCreatedRoleRequests   = "created_role_requests"
	UserEdgeRespondedRoleRequests = "responded_role_requests"
)

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
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
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
		Field(
			dmodel.DefineField().
				Name(UserFieldOrgUnitId).
				Label(model.LangJson{"en-US": "Organizational Unit"}).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(UserEdgeOrgUnit).
				ManyToOne(OrganizationalUnitSchemaName, dmodel.DynamicFields{
					UserFieldOrgUnitId: basemodel.FieldId,
				}).
				OnDelete(dmodel.RelationCascadeSetNull),
		).
		EdgeTo(
			dmodel.Edge(UserEdgeGroups).
				ManyToMany(GroupSchemaName, GrpUsrRelSchemaName, "user").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserEdgeOrgs).
				ManyToMany(OrganizationSchemaName, OrgUsrRelSchemaName, "user").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserEdgeRoles).
				ManyToMany(RoleSchemaName, RoleAssignmentSchemaName, "receiver_user").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(UserEdgePrivateRole).
				Label(model.LangJson{"en-US": "Private role"}).
				Existing(RoleSchemaName, RoleEdgeDedicatedUser),
		).
		EdgeFrom(
			dmodel.Edge(UserEdgeBenefitRoleRequests).
				Label(model.LangJson{"en-US": "Grant requests for me"}).
				Existing(RoleRequestSchemaName, RoleReqEdgeReceiverUser),
		).
		EdgeFrom(
			dmodel.Edge(UserEdgeCreatedRoleRequests).
				Label(model.LangJson{"en-US": "Grant requests created by me"}).
				Existing(RoleRequestSchemaName, RoleReqEdgeRequestor),
		).
		EdgeFrom(
			dmodel.Edge(UserEdgeRespondedRoleRequests).
				Label(model.LangJson{"en-US": "Grant requests responded by me"}).
				Existing(RoleRequestSchemaName, RoleReqEdgeResponder),
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

func (this User) GetOrgUnitId() *model.Id {
	return this.fields.GetModelId(UserFieldOrgUnitId)
}

func (this *User) SetOrgUnitId(v *model.Id) {
	this.fields.SetModelId(UserFieldOrgUnitId, v)
}
