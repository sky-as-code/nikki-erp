package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	UserPermissionSchemaName = "authorize.user_permission"

	UserPermFieldUserId                = "user_id"
	UserPermFieldEntExpression         = "ent_expression"
	UserPermFieldEntId                 = "ent_id"
	UserPermFieldActionId              = "action_id"
	UserPermFieldResourceId            = "resource_id"
	UserPermFieldResourceCode          = "resource_code"
	UserPermFieldRoleGroupAssignmentId = "role_group_assignment_id"
	UserPermFieldRoleUserAssignmentId  = "role_user_assignment_id"
	UserPermFieldScope                 = "scope"
	UserPermFieldOrgId                 = "org_id"
	UserPermFieldOrgUnitId             = "org_unit_id"
	UserPermFieldOrgMembershipId       = "org_membership_id"
	UserPermFieldGroupMembershipId     = "group_membership_id"

	UserPermEdgeUser                = "user"
	UserPermEdgeAction              = "action"
	UserPermEdgeResource            = "resource"
	UserPermEdgeEntitlement         = "entitlement"
	UserPermEdgeRoleGroupAssignment = "role_group_assignment"
	UserPermEdgeRoleUserAssignment  = "role_user_assignment"
	UserPermEdgeOrg                 = "org"
	UserPermEdgeOrgUnit             = "org_unit"
	UserPermEdgeOrgMembership       = "org_membership"
	UserPermEdgeGroupMembership     = "group_membership"
)

func UserPermissionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(UserPermissionSchemaName).
		Label(model.LangJson{"en-US": "User Permission"}).
		TableName("authz_user_permissions").
		CompositeUnique(UserPermFieldUserId, UserPermFieldEntExpression).
		ShouldBuildDb().
		Field(
			basemodel.DefineFieldId(UserPermFieldUserId).
				PrimaryKey(),
		).
		Field(
			basemodel.DefineFieldId(UserPermFieldEntId).
				PrimaryKey(),
		).
		Field(
			DefineEntitlementFieldExpression(UserPermFieldEntExpression).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(UserPermFieldActionId),
		).
		Field(
			basemodel.DefineFieldId(UserPermFieldResourceId),
		).
		Field(
			DefineResourceFieldCode(UserPermFieldResourceCode),
		).
		Field(
			basemodel.DefineFieldId(UserPermFieldRoleGroupAssignmentId),
		).
		Field(
			basemodel.DefineFieldId(UserPermFieldRoleUserAssignmentId),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldScope).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(UserPermFieldOrgId),
		).
		Field(
			basemodel.DefineFieldId(UserPermFieldOrgMembershipId),
		).
		Field(
			basemodel.DefineFieldId(UserPermFieldGroupMembershipId),
		).
		//
		// When one of the following record is deleted, the user permission record will be deleted too.
		//
		Field(
			basemodel.DefineFieldId(UserPermFieldOrgUnitId),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeUser).
				Label(model.LangJson{"en-US": "User"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					UserPermFieldUserId: UserFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeAction).
				Label(model.LangJson{"en-US": "Action"}).
				ManyToOne(ActionSchemaName, dmodel.DynamicFields{
					UserPermFieldActionId: ActionFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeResource).
				Label(model.LangJson{"en-US": "Resource"}).
				ManyToOne(ResourceSchemaName, dmodel.DynamicFields{
					UserPermFieldResourceId: ResourceFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeEntitlement).
				Label(model.LangJson{"en-US": "Entitlement"}).
				ManyToOne(EntitlementSchemaName, dmodel.DynamicFields{
					UserPermFieldEntId: EntitlementFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeRoleGroupAssignment).
				Label(model.LangJson{"en-US": "Role-Group Assignment"}).
				ManyToOne(RoleGroupAssignmentSchemaName, dmodel.DynamicFields{
					UserPermFieldRoleGroupAssignmentId: RoleGroupAssignFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeRoleUserAssignment).
				Label(model.LangJson{"en-US": "Role-User Assignment"}).
				ManyToOne(RoleUserAssignmentSchemaName, dmodel.DynamicFields{
					UserPermFieldRoleUserAssignmentId: RoleUserAssignFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeOrg).
				Label(model.LangJson{"en-US": "Organization"}).
				ManyToOne(OrganizationSchemaName, dmodel.DynamicFields{
					UserPermFieldOrgId: OrgFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeOrgUnit).
				Label(model.LangJson{"en-US": "Organizational Unit"}).
				ManyToOne(OrganizationalUnitSchemaName, dmodel.DynamicFields{
					UserPermFieldOrgUnitId: OrgUnitFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeOrgMembership).
				Label(model.LangJson{"en-US": "Organization Membership"}).
				ManyToOne(OrgUsrRelSchemaName, dmodel.DynamicFields{
					UserPermFieldOrgMembershipId: OrgUsrRelFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(UserPermEdgeGroupMembership).
				Label(model.LangJson{"en-US": "Group Membership"}).
				ManyToOne(GrpUsrRelSchemaName, dmodel.DynamicFields{
					UserPermFieldGroupMembershipId: GrpUsrRelFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

// Represents a cached item from user permission calculation process.
// Used for O(log n) lookup of user permissions.
// The records will be deleted when the related data are deleted.
type UserPermission struct {
	fields dmodel.DynamicFields
}

func NewUserPermission() *UserPermission {
	return &UserPermission{fields: make(dmodel.DynamicFields)}
}

func NewUserPermissionFrom(src dmodel.DynamicFields) *UserPermission {
	return &UserPermission{fields: src}
}

func (this UserPermission) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *UserPermission) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}
