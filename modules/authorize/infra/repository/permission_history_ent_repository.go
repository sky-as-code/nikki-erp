package repository

import (
	"github.com/sky-as-code/nikki-erp/common/orm"

	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/permissionhistory"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/permission_history"
)

func NewPermissionHistoryEntRepository(client *ent.Client) it.PermissionHistoryRepository {
	return &PermissionHistoryEntRepository{
		client: client,
	}
}

type PermissionHistoryEntRepository struct {
	client *ent.Client
}

func BuildPermissionHistoryDescriptor() *orm.EntityDescriptor {
	entity := ent.PermissionHistory{}
	builder := orm.DescribeEntity(entPermissionHistory.Label).
		Aliases("permission_histories").
		Field(entPermissionHistory.FieldID, entity.ID).
		Field(entPermissionHistory.FieldApproverID, entity.ApproverID).
		Field(entPermissionHistory.FieldApproverEmail, entity.ApproverEmail).
		Field(entPermissionHistory.FieldEffect, entity.Effect).
		Field(entPermissionHistory.FieldReason, entity.Reason).
		Field(entPermissionHistory.FieldEntitlementID, entity.EntitlementID).
		Field(entPermissionHistory.FieldEntitlementExpr, entity.EntitlementExpr).
		Field(entPermissionHistory.FieldEntitlementAssignmentID, entity.EntitlementAssignmentID).
		Field(entPermissionHistory.FieldResolvedExpr, entity.ResolvedExpr).
		Field(entPermissionHistory.FieldReceiverID, entity.ReceiverID).
		Field(entPermissionHistory.FieldReceiverEmail, entity.ReceiverEmail).
		Field(entPermissionHistory.FieldGrantRequestID, entity.GrantRequestID).
		Field(entPermissionHistory.FieldRevokeRequestID, entity.RevokeRequestID).
		Field(entPermissionHistory.FieldRoleID, entity.RoleID).
		Field(entPermissionHistory.FieldRoleName, entity.RoleName).
		Field(entPermissionHistory.FieldRoleSuiteID, entity.RoleSuiteID).
		Field(entPermissionHistory.FieldRoleSuiteName, entity.RoleSuiteName).
		Field(entPermissionHistory.FieldCreatedAt, entity.CreatedAt)

	return builder.Descriptor()
}
