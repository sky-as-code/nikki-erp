package models

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	PermissionHistorySchemaName = "iam_permission_history"

	PermHistoryFieldApproverId      = "approver_id"
	PermHistoryFieldApproverEmail   = "approver_email"
	PermHistoryFieldEffect          = "effect"
	PermHistoryFieldReason          = "reason"
	PermHistoryFieldEntitlementId   = "entitlement_id"
	PermHistoryFieldEntitlementExpr = "entitlement_expr"
	PermHistoryFieldAssignmentId    = "entitlement_assignment_id"
	PermHistoryFieldResolvedExpr    = "resolved_expr"
	PermHistoryFieldReceiverId      = "receiver_id"
	PermHistoryFieldReceiverEmail   = "receiver_email"
	PermHistoryFieldRoleRequestId   = "grant_request_id"
	PermHistoryFieldRevokeRequestId = "revoke_request_id"
	PermHistoryFieldRoleId          = "role_id"
	PermHistoryFieldRoleName        = "role_name"
)

var permissionHistoryReasonValues = []string{
	"ent_added", "ent_removed", "ent_deleted",
	"ent_added_group", "ent_removed_group", "ent_deleted_group",
	"ent_added_role", "ent_removed_role", "ent_deleted_role",
	"ent_added_role_group", "ent_removed_role_group", "ent_deleted_role_group",
	"role_added", "role_removed", "role_deleted",
	"role_added_group", "role_removed_group", "role_deleted_group",
}

func PermissionHistorySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(PermissionHistorySchemaName).
		Label(model.NewLangJsonRefSf("%s.label", PermissionHistorySchemaName)).
		TableName("iam_permission_histories").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldApproverId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldApproverEmail).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldEffect).
				DataType(dmodel.FieldDataTypeEnumString([]string{"grant", "revoke"})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldReason).
				DataType(dmodel.FieldDataTypeEnumString(permissionHistoryReasonValues)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldEntitlementId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldEntitlementExpr).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldAssignmentId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldResolvedExpr).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldReceiverId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldReceiverEmail).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldRoleRequestId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldRevokeRequestId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldRoleId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(PermHistoryFieldRoleName).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_SHORT_NAME_LENGTH)),
		).
		Extend(basemodel.AuditableReadonlyModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge("role").
				Label(model.LangJson{"en-US": "Role"}).
				ManyToOne(RoleSchemaName, dmodel.DynamicFields{
					PermHistoryFieldRoleId: RoleFieldId,
				}).
				OnDelete(dmodel.RelationCascadeSetNull),
		).
		EdgeTo(
			dmodel.Edge("grant_request").
				Label(model.LangJson{"en-US": "Grant Request"}).
				ManyToOne(RoleRequestSchemaName, dmodel.DynamicFields{
					PermHistoryFieldRoleRequestId: RoleReqFieldId,
				}).
				OnDelete(dmodel.RelationCascadeSetNull),
		)
}

type PermissionHistory struct {
	modelData basemodel.DynamicModelBase `json:"-"`

	ApproverId              *model.Id                `json:"approverId,omitempty"`
	ApproverEmail           *string                  `json:"approverEmail,omitempty"`
	Effect                  *PermissionHistoryEffect `json:"effect,omitempty"`
	Reason                  *PermissionHistoryReason `json:"reason,omitempty"`
	EntitlementId           *model.Id                `json:"entitlementId,omitempty"`
	EntitlementExpr         *string                  `json:"entitlementExpr,omitempty"`
	EntitlementAssignmentId *model.Id                `json:"assignmentId,omitempty"`
	ResolvedExpr            *string                  `json:"resolvedExpr,omitempty"`
	ReceiverId              *model.Id                `json:"receiverId,omitempty"`
	ReceiverEmail           *string                  `json:"receiverEmail,omitempty"`
	GrantRequestId          *model.Id                `json:"grantRequestId,omitempty"`
	RevokeRequestId         *model.Id                `json:"revokeRequestId,omitempty"`
	ResourceId              *model.Id                `json:"resourceId,omitempty"`
	RoleId                  *model.Id                `json:"roleId,omitempty"`
	RoleName                *string                  `json:"roleName,omitempty"`
	ScopeRef                *string                  `json:"scopeRef,omitempty"`
	SubjectRef              *string                  `json:"subjectRef,omitempty"`
	// SubjectType     *EntitlementSubjectType  `json:"subjectType,omitempty"`
}

func NewPermissionHistory() *PermissionHistory {
	return &PermissionHistory{modelData: basemodel.NewDynamicModel()}
}

func NewPermissionHistoryFrom(src dmodel.DynamicFields) *PermissionHistory {
	return &PermissionHistory{modelData: basemodel.NewDynamicModel(src)}
}

func (this PermissionHistory) GetFieldData() dmodel.DynamicFields {
	return this.modelData.GetFieldData()
}

func (this *PermissionHistory) SetFieldData(data dmodel.DynamicFields) {
	this.modelData.SetFieldData(data)
}

type PermissionHistoryEffect string

type PermissionHistoryReason string

const (
	PermissionHistoryReasonEntAdded   = PermissionHistoryReason("ent_added")
	PermissionHistoryReasonEntRemoved = PermissionHistoryReason("ent_removed")
	PermissionHistoryReasonEntDeleted = PermissionHistoryReason("ent_deleted")

	PermissionHistoryReasonEntAddedGroup   = PermissionHistoryReason("ent_added_group")
	PermissionHistoryReasonEntRemovedGroup = PermissionHistoryReason("ent_removed_group")
	PermissionHistoryReasonEntDeletedGroup = PermissionHistoryReason("ent_deleted_group")

	PermissionHistoryReasonEntAddedRole   = PermissionHistoryReason("ent_added_role")
	PermissionHistoryReasonEntRemovedRole = PermissionHistoryReason("ent_removed_role")
	PermissionHistoryReasonEntDeletedRole = PermissionHistoryReason("ent_deleted_role")

	PermissionHistoryReasonEntAddedRoleGroup   = PermissionHistoryReason("ent_added_role_group")
	PermissionHistoryReasonEntRemovedRoleGroup = PermissionHistoryReason("ent_removed_role_group")
	PermissionHistoryReasonEntDeletedRoleGroup = PermissionHistoryReason("ent_deleted_role_group")

	PermissionHistoryReasonRoleAdded   = PermissionHistoryReason("role_added")
	PermissionHistoryReasonRoleRemoved = PermissionHistoryReason("role_removed")
	PermissionHistoryReasonRoleDeleted = PermissionHistoryReason("role_deleted")

	PermissionHistoryReasonRoleAddedGroup   = PermissionHistoryReason("role_added_group")
	PermissionHistoryReasonRoleRemovedGroup = PermissionHistoryReason("role_removed_group")
	PermissionHistoryReasonRoleDeletedGroup = PermissionHistoryReason("role_deleted_group")
)
