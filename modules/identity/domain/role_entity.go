package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	RoleSchemaName = "authorize.role"

	RoleFieldId                = "id"
	RoleFieldName              = "name"
	RoleFieldDescription       = "description"
	RoleFieldOwnerUserId       = "owner_user_id"
	RoleFieldOwnerGroupId      = "owner_group_id"
	RoleFieldIsPrivate         = "is_private"
	RoleFieldIsRequestable     = "is_requestable"
	RoleFieldIsRequiredAttach  = "is_required_attachment"
	RoleFieldIsRequiredComment = "is_required_comment"
	RoleFieldOrgId             = "org_id"

	RoleEdgeRoleRequests   = "role_requests"
	RoleEdgeEntitlements   = "entitlements"
	RoleEdgeAssignedGroups = "assigned_groups"
	RoleEdgeAssignedUsers  = "assigned_users"
	RoleEdgeOwnerGroup     = "owner_group"
	RoleEdgeOwnerUser      = "owner_user"
)

func RoleSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RoleSchemaName).
		Label(model.LangJson{"en-US": "Role"}).
		TableName("authz_roles").
		PartialUnique(RoleFieldName, RoleFieldOrgId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(RoleFieldName).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldDescription).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(RoleFieldOwnerGroupId).
				Description(model.LangJson{"en-US": "One of the users in this group can approve grant requests for this role"}),
		).
		Field(
			basemodel.DefineFieldId(RoleFieldOwnerUserId).
				Description(model.LangJson{"en-US": "Only this user can approve grant requests for this role"}),
		).
		ExclusiveFields(RoleFieldOwnerGroupId, RoleFieldOwnerUserId).
		Field(
			dmodel.DefineField().Name(RoleFieldIsPrivate).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate().
				Default(false),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldIsRequestable).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldIsRequiredAttach).
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldIsRequiredComment).
				DataType(dmodel.FieldDataTypeBoolean()).
				Default(false),
		).
		Field(
			basemodel.DefineFieldId(RoleFieldOrgId).
				Description(model.LangJson{"en-US": "If specified, the role only accepts entitlements whose org_unit_id belongs to this organization. " +
					"Otherwise, the role only accepts entitlements with domain scope (org_unit_id is nil)",
				}),
		).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		EdgeFrom(
			dmodel.Edge(RoleEdgeRoleRequests).
				Label(model.LangJson{"en-US": "Grant requests"}).
				Existing(RoleRequestSchemaName, RoleReqEdgeRole),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeOwnerGroup).
				Label(model.LangJson{"en-US": "Owner group"}).
				ManyToOne(GroupSchemaName, dmodel.DynamicFields{
					RoleFieldOwnerGroupId: GroupFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeOwnerUser).
				Label(model.LangJson{"en-US": "Owner user"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleFieldOwnerUserId: UserFieldId,
				}),
		).
		EdgeFrom(
			dmodel.Edge(RoleEdgeEntitlements).
				Label(model.LangJson{"en-US": "Entitlements"}).
				Existing(EntitlementSchemaName, EntitlementEdgeRole),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeAssignedGroups).
				Label(model.LangJson{"en-US": "Assigned groups"}).
				ManyToMany(GroupSchemaName, RoleGroupAssignmentSchemaName, "role"),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeAssignedUsers).
				Label(model.LangJson{"en-US": "Assigned users"}).
				ManyToMany(UserSchemaName, RoleUserAssignmentSchemaName, "role"),
		)
}

type Role struct {
	basemodel.DynamicModelBase
}

func NewRole() *Role {
	return &Role{basemodel.NewDynamicModel()}
}

func NewRoleFrom(src dmodel.DynamicFields) *Role {
	return &Role{basemodel.NewDynamicModel(src)}
}

func (this Role) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *Role) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this Role) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(RoleFieldOrgId)
}

func (this *Role) SetOrgId(id *model.Id) {
	this.GetFieldData().SetModelId(RoleFieldOrgId, id)
}

func (this Role) IsPrivate() *bool {
	return this.GetFieldData().GetBool(RoleFieldIsPrivate)
}

func (this *Role) SetIsPrivate(v *bool) {
	this.GetFieldData().SetBool(RoleFieldIsPrivate, v)
}
