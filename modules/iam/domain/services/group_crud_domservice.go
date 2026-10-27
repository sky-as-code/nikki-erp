package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itGroup "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// NewGroupCrudDomainService is handed the composable default by the iam_group onion.
func NewGroupCrudDomainService(
	base composable.CrudDomainService,
	permRepo itPerm.PermissionRepository,
	historyRepo itPerm.PermissionHistoryRepository,
	userRepo itUser.UserCrudRepository,
) itGroup.GroupCrudDomainService {
	return &GroupCrudDomainServiceImpl{
		CrudDomainService: base,
		permRepo:          permRepo,
		auditor:           permissionAuditor{historyRepo: historyRepo},
		userRepo:          userRepo,
	}
}

// GroupCrudDomainServiceImpl keeps groups human: a service is neither a member nor the owner.
type GroupCrudDomainServiceImpl struct {
	composable.CrudDomainService
	permRepo itPerm.PermissionRepository
	auditor  permissionAuditor
	userRepo itUser.UserCrudRepository
}

func (this *GroupCrudDomainServiceImpl) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	opts := composable.CreateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(
		refuseServiceIn(this.userRepo, domain.GroupFieldOwnerId), opts.ValidateExtra)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *GroupCrudDomainServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand, options ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	opts := composable.UpdateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(
		refuseServiceIn(this.userRepo, domain.GroupFieldOwnerId), opts.ValidateExtra)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// ManageUsers changes the group's membership, then refreshes each added and removed user's
// permissions, since membership decides which of the group's roles apply to whom.
func (this *GroupCrudDomainServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itGroup.ManageGroupUsersCommand,
) (*itGroup.ManageGroupUsersResult, error) {
	return composable.ManageM2m(ctx, this.Repository(), cmd, composable.ManageM2mParam{
		Action:          "manage group users",
		DestSchemaName:  domain.UserSchemaName,
		JunctionIdField: domain.GrpUsrRelFieldId,
		Validate: func(ctx corectx.Context, change composable.M2mChange, vErrs *ft.ClientErrors) error {
			return serviceTypeScope.refuseIdsOfType(ctx, this.userRepo, composable.FieldM2mAdd, change.Added, vErrs)
		},
		AfterChange: func(tranxCtx corectx.Context, change composable.M2mChange) error {
			for _, userId := range change.AllPeerIds() {
				if err := this.permRepo.RebuildUserPermission(tranxCtx, userId); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

// ManageRoles assigns roles to and removes them from the group, then refreshes every member's
// permissions. The audit is recorded against the group, not its members: membership decides who
// this reaches, and that list changes independently of this event.
func (this *GroupCrudDomainServiceImpl) ManageRoles(
	ctx corectx.Context, cmd itGroup.ManageGroupRolesCommand,
) (*itGroup.ManageGroupRolesResult, error) {
	return composable.ManageM2m(ctx, this.Repository(), cmd, composable.ManageM2mParam{
		Action:          "manage group role assignments",
		DestSchemaName:  domain.RoleSchemaName,
		JunctionIdField: domain.RoleGroupAssignFieldId,
		AfterChange: func(tranxCtx corectx.Context, change composable.M2mChange) error {
			if err := this.permRepo.RebuildUserPermissionsForGroup(tranxCtx, change.SrcId); err != nil {
				return err
			}
			return this.auditor.recordRoleChange(tranxCtx, change.Added, change.Removed,
				[]model.Id{change.SrcId},
				domain.PermissionHistoryReasonRoleAddedGroup, domain.PermissionHistoryReasonRoleRemovedGroup)
		},
	})
}
