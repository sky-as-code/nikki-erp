package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

const (
	UserPermissionSchemaName = "authorize.user_permission"

	UserPermFieldUserId            = "user_id"
	UserPermFieldEntExpression     = "ent_expression"
	UserPermFieldEntId             = "ent_id"
	UserPermFieldActionId          = "action_id"
	UserPermFieldActionCode        = "action_code"
	UserPermFieldResourceId        = "resource_id"
	UserPermFieldResourceCode      = "resource_code"
	UserPermFieldRoleAssignmentId  = "role_assignment_id"
	UserPermFieldScope             = "scope"
	UserPermFieldOrgId             = "org_id"
	UserPermFieldOrgUnitId         = "org_unit_id"
	UserPermFieldOrgMembershipId   = "org_membership_id"
	UserPermFieldGroupMembershipId = "group_membership_id"

	UserPermEdgeUser            = "user"
	UserPermEdgeAction          = "action"
	UserPermEdgeResource        = "resource"
	UserPermEdgeEntitlement     = "entitlement"
	UserPermEdgeRoleAssignment  = "role_assignment"
	UserPermEdgeOrg             = "org"
	UserPermEdgeOrgUnit         = "org_unit"
	UserPermEdgeOrgMembership   = "org_membership"
	UserPermEdgeGroupMembership = "group_membership"
)

func UserPermissionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(UserPermissionSchemaName).
		Label(model.LangJson{"en-US": "User Permission"}).
		TableName("authz_user_permissions").
		CompositeUnique(UserPermFieldUserId, UserPermFieldEntExpression).
		ShouldBuildDb().
		Field(
			dmodel.DefineField().Name(UserPermFieldUserId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldEntId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldEntExpression).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_SHORT_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldActionId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldActionCode).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldResourceId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldResourceCode).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldRoleAssignmentId).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldScope).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldOrgId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldOrgMembershipId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(UserPermFieldGroupMembershipId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		//
		// When one of the following record is deleted, the user permission record will be deleted too.
		//
		Field(
			dmodel.DefineField().Name(UserPermFieldOrgUnitId).
				DataType(dmodel.FieldDataTypeUlid()),
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
				ManyToOne(ActionSchemaName, dmodel.DynamicFields{
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
			dmodel.Edge(UserPermEdgeRoleAssignment).
				Label(model.LangJson{"en-US": "Role Assignment"}).
				ManyToOne(RoleAssignmentSchemaName, dmodel.DynamicFields{
					UserPermFieldRoleAssignmentId: RoleAssignFieldId,
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
