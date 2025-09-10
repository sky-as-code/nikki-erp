package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
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

func (this *PermissionHistoryEntRepository) Create(ctx crud.Context, permissionHistory domain.PermissionHistory) (*domain.PermissionHistory, error) {
	var creation *ent.PermissionHistoryCreate
	tx := ctx.GetDbTranx().(*ent.Tx)

	if tx != nil {
		creation = tx.PermissionHistory.Create()
	} else {
		creation = this.client.PermissionHistory.Create()
	}

	creation = creation.
		SetID(*permissionHistory.Id).
		SetNillableApproverID(permissionHistory.ApproverId).
		SetNillableApproverEmail(permissionHistory.ApproverEmail).
		SetEffect(entPermissionHistory.Effect(*permissionHistory.Effect)).
		SetReason(entPermissionHistory.Reason(*permissionHistory.Reason)).
		SetNillableEntitlementID(permissionHistory.EntitlementId).
		SetNillableEntitlementExpr(permissionHistory.EntitlementExpr).
		SetNillableEntitlementAssignmentID(permissionHistory.EntitlementAssignmentId).
		SetNillableResolvedExpr(permissionHistory.ResolvedExpr).
		SetNillableReceiverID(permissionHistory.ReceiverId).
		SetNillableReceiverEmail(permissionHistory.ReceiverEmail).
		SetNillableGrantRequestID(permissionHistory.GrantRequestId).
		SetNillableRevokeRequestID(permissionHistory.RevokeRequestId).
		SetNillableRoleID(permissionHistory.RoleId).
		SetNillableRoleName(permissionHistory.RoleName).
		SetNillableRoleSuiteID(permissionHistory.RoleSuiteId).
		SetNillableRoleSuiteName(permissionHistory.RoleSuiteName).
		SetCreatedAt(time.Now())

	return database.Mutate(ctx, creation, ent.IsNotFound, entToPermissionHistory)
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
