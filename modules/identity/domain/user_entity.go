package domain

import (
	"go.bryk.io/pkg/errors"

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
	UserResourceCode = "identity_user"
	UserAuthScope    = "org"

	UserActionCreate      = "create"
	UserActionDelete      = "delete"
	UserActionUpdate      = "update"
	UserActionView        = "view"
	UserActionSetArchived = "set_archived"
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
	UserEdgeOwnRoles              = "own_roles"
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
			basemodel.DefineFieldId(UserFieldOrgUnitId).
				Label(model.LangJson{"en-US": "Organizational Unit"}),
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
				ManyToMany(RoleSchemaName, RoleUserAssignmentSchemaName, "receiver_user").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(UserEdgeOwnRoles).
				Label(model.LangJson{"en-US": "Owned roles"}).
				Existing(RoleSchemaName, RoleEdgeOwnerUser),
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
	basemodel.DynamicModelBase
}

func NewUser() *User {
	return &User{basemodel.NewDynamicModel()}
}

func NewUserFrom(src dmodel.DynamicFields) *User {
	return &User{basemodel.NewDynamicModel(src)}
}

func (this User) MustGetId() model.Id {
	v := this.GetFieldData().GetModelId(basemodel.FieldId)
	if v == nil {
		panic("id is nil")
	}
	return *v
}

func (this User) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *User) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this User) IsActive() bool {
	return this.MustIsArchived() == false && this.MustGetStatus() == UserStatusActive
}

func (this User) MustIsArchived() bool {
	val := this.IsArchived()
	if val == nil {
		panic(errors.New("is_archived is nil"))
	}
	return *val
}

func (this User) IsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

func (this *User) SetIsArchived(v *bool) {
	this.GetFieldData().SetBool(basemodel.FieldIsArchived, v)
}

func (this User) GetAvatarUrl() *string {
	return this.GetFieldData().GetString(UserFieldAvatarUrl)
}

func (this *User) SetAvatarUrl(v *string) {
	this.GetFieldData().SetString(UserFieldAvatarUrl, v)
}

func (this User) MustGetDisplayName() string {
	v := this.GetFieldData().GetString(UserFieldDisplayName)
	if v == nil {
		panic("display_name is nil")
	}
	return *v
}

func (this User) GetDisplayName() *string {
	return this.GetFieldData().GetString(UserFieldDisplayName)
}

func (this *User) SetDisplayName(v *string) {
	this.GetFieldData().SetString(UserFieldDisplayName, v)
}

func (this User) MustGetEmail() string {
	v := this.GetFieldData().GetString(UserFieldEmail)
	if v == nil {
		panic("email is nil")
	}
	return *v
}

func (this User) GetEmail() *string {
	return this.GetFieldData().GetString(UserFieldEmail)
}

func (this *User) SetEmail(v *string) {
	this.GetFieldData().SetString(UserFieldEmail, v)
}

func (this User) IsOwner() bool {
	val := this.GetFieldData().GetBool(UserFieldIsOwner)
	return val != nil && *val
}

func (this *User) SetIsOwner(v *bool) {
	this.GetFieldData().SetBool(UserFieldIsOwner, v)
}

func (this User) MustGetStatus() UserStatus {
	val := this.GetStatus()
	if val == nil {
		panic("status is nil")
	}
	return *val
}

func (this User) GetStatus() *UserStatus {
	s := this.GetFieldData().GetString(UserFieldStatus)
	if s == nil {
		return nil
	}
	st := UserStatus(*s)
	return &st
}

func (this *User) SetStatus(v *UserStatus) {
	if v == nil {
		this.GetFieldData().SetString(UserFieldStatus, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(UserFieldStatus, &s)
}

func (this User) GetOrgUnitId() *model.Id {
	return this.GetFieldData().GetModelId(UserFieldOrgUnitId)
}

func (this *User) SetOrgUnitId(v *model.Id) {
	this.GetFieldData().SetModelId(UserFieldOrgUnitId, v)
}
