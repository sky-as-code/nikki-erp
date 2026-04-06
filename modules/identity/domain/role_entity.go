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
	RoleFieldDedicatedUserId   = "dedicated_user_id"
	RoleFieldDedicatedGroupId  = "dedicated_group_id"
	RoleFieldOwnerUserId       = "owner_user_id"
	RoleFieldOwnerGroupId      = "owner_group_id"
	RoleFieldIsRequestable     = "is_requestable"
	RoleFieldIsRequiredAttach  = "is_required_attachment"
	RoleFieldIsRequiredComment = "is_required_comment"
	RoleFieldOrgId             = "org_id"

	RoleEdgeRoleRequests   = "role_requests"
	RoleEdgeEntitlements   = "entitlements"
	RoleEdgeAssignedGroups = "assigned_groups"
	RoleEdgeAssignedUsers  = "assigned_users"
	RoleEdgeDedicatedGroup = "dedicated_group"
	RoleEdgeDedicatedUser  = "dedicated_user"
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
			dmodel.DefineField().Name(RoleFieldDedicatedGroupId).
				DataType(dmodel.FieldDataTypeUlid()).
				Unique().
				Description(model.LangJson{"en-US": "This role is implicit (hidden) role belonging to this group"}),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldDedicatedUserId).
				DataType(dmodel.FieldDataTypeUlid()).
				Unique().
				Description(model.LangJson{"en-US": "This role is implicit (hidden) role belonging to this user"}),
		).
		ExclusiveFields(RoleFieldDedicatedGroupId, RoleFieldDedicatedUserId).
		Field(
			dmodel.DefineField().Name(RoleFieldOwnerGroupId).
				DataType(dmodel.FieldDataTypeUlid()).
				Description(model.LangJson{"en-US": "One of the users in this group can approve grant requests for this role"}),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldOwnerUserId).
				DataType(dmodel.FieldDataTypeUlid()).
				Description(model.LangJson{"en-US": "Only this user can approve grant requests for this role"}),
		).
		ExclusiveFields(RoleFieldOwnerGroupId, RoleFieldOwnerUserId).
		Field(
			dmodel.DefineField().Name(RoleFieldIsRequestable).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldIsRequiredAttach).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldIsRequiredComment).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleFieldOrgId).
				DataType(dmodel.FieldDataTypeUlid()).
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
			dmodel.Edge(RoleEdgeDedicatedGroup).
				Label(model.LangJson{"en-US": "Dedicated group"}).
				ManyToOne(GroupSchemaName, dmodel.DynamicFields{
					RoleFieldDedicatedGroupId: GroupFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeDedicatedUser).
				Label(model.LangJson{"en-US": "Dedicated user"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleFieldDedicatedUserId: UserFieldId,
				}),
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
				ManyToMany(GroupSchemaName, RoleAssignmentSchemaName, "role"),
		).
		EdgeTo(
			dmodel.Edge(RoleEdgeAssignedUsers).
				Label(model.LangJson{"en-US": "Assigned users"}).
				ManyToMany(UserSchemaName, RoleAssignmentSchemaName, "role"),
		)
}

type Role struct {
	fields dmodel.DynamicFields
}

func NewRole() *Role {
	return &Role{fields: make(dmodel.DynamicFields)}
}

func NewRoleFrom(src dmodel.DynamicFields) *Role {
	return &Role{fields: src}
}

func (this Role) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Role) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Role) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *Role) SetId(id *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, id)
}

func (this Role) GetOrgId() *model.Id {
	return this.fields.GetModelId(RoleFieldOrgId)
}

func (this *Role) SetOrgId(id *model.Id) {
	this.fields.SetModelId(RoleFieldOrgId, id)
}
