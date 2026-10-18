package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itGroup "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
)

// NewGroupCrudApplicationService is handed the composable default by the iam_group onion.
func NewGroupCrudApplicationService(
	base composable.CrudApplicationService, domSvc itGroup.GroupCrudDomainService,
) itGroup.GroupCrudApplicationService {
	return &GroupCrudApplicationServiceImpl{CrudApplicationService: base, domSvc: domSvc}
}

type GroupCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	domSvc itGroup.GroupCrudDomainService
}

func (this *GroupCrudApplicationServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itGroup.ManageGroupUsersCommand,
) (*itGroup.ManageGroupUsersResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd, composable.ActionPermission{ActionCode: actionManageUsers})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageUsers(ctx, cmd)
}

// ManageRoles asserts manage_role_assignments, the same permission the role side asserts for
// iam_role/:id/manage_groups, so neither side opens a path the other keeps shut.
func (this *GroupCrudApplicationServiceImpl) ManageRoles(
	ctx corectx.Context, cmd itGroup.ManageGroupRolesCommand,
) (*itGroup.ManageGroupRolesResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd,
		composable.ActionPermission{ActionCode: actionManageRoleAssignments})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageRoles(ctx, cmd)
}
