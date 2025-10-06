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

func (this *PermissionHistoryEntRepository) permissionHistoryClient(ctx crud.Context) *ent.PermissionHistoryClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.PermissionHistory
	}
	return this.client.PermissionHistory
}

func (this *PermissionHistoryEntRepository) Create(ctx crud.Context, permissionHistory *domain.PermissionHistory) (*domain.PermissionHistory, error) {
	creation := this.permissionHistoryClient(ctx).Create().
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

func (this *PermissionHistoryEntRepository) EnableField(ctx crud.Context, params it.EnableFieldParam) error {
	enableField := this.permissionHistoryClient(ctx).Update()

	// Enable Entitlement
	if params.EntitlementId != nil {
		enableField.Where(
			entPermissionHistory.EntitlementIDEQ(*params.EntitlementId),
		).
			ClearEntitlementID().
			SetEntitlementExpr(params.EntitlementExpr)
	}

	// Enable Assignment
	if params.AssignmentId != nil {
		enableField.Where(
			entPermissionHistory.EntitlementAssignmentIDEQ(*params.AssignmentId),
		).
			ClearEntitlementAssignmentID().
			SetResolvedExpr(params.ResolvedExpr)
	}

	return enableField.Exec(ctx)
}

func (this *PermissionHistoryEntRepository) FindAllByEntitlementId(ctx crud.Context, param it.FindAllByEntitlementIdParam) ([]domain.PermissionHistory, error) {
	query := this.permissionHistoryClient(ctx).Query().
		Where(entPermissionHistory.EntitlementIDEQ(param.EntitlementId))

	return database.List(ctx, query, entToPermissionHistories)
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
