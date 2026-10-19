package app

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

// NewRoleCrudApplicationService is handed the composable default by the iam_role onion.
func NewRoleCrudApplicationService(
	base composable.CrudApplicationService, domSvc itRole.RoleCrudDomainService,
) itRole.RoleCrudApplicationService {
	return &RoleCrudApplicationServiceImpl{CrudApplicationService: base, domSvc: domSvc}
}

type RoleCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	domSvc itRole.RoleCrudDomainService
}

// Update needs manage_entitlements on top of update when the body carries an entitlement delta:
// changing what a role grants is that permission, whichever route carries it.
func (this *RoleCrudApplicationServiceImpl) Update(
	ctx corectx.Context, cmd composable.UpdateCommand,
) (*composable.MutateResult, error) {
	if itRole.HasEntitlementsDelta(cmd) {
		cErrs, err := composable.AssertRecordAction(ctx, this, cmd,
			composable.ActionPermission{ActionCode: actionManageEntitlements})
		if err != nil || cErrs != nil {
			return mutateRefused(cErrs, err)
		}
	}
	return this.CrudApplicationService.Update(ctx, cmd)
}

// ManageUsers asserts exactly what iam_user/:id/manage_roles asserts, so the role side opens no
// path the principal side keeps shut.
func (this *RoleCrudApplicationServiceImpl) ManageUsers(
	ctx corectx.Context, cmd itRole.ManageRoleUsersCommand,
) (*itRole.ManageRoleUsersResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd, composable.ActionPermission{
		ActionCode: actionManageRoleAssignments, ResourceCode: c.ResourceIamUser,
	})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageUsers(ctx, cmd)
}

// ManageGroups mirrors iam_group/:id/manage_roles the same way.
func (this *RoleCrudApplicationServiceImpl) ManageGroups(
	ctx corectx.Context, cmd itRole.ManageRoleGroupsCommand,
) (*itRole.ManageRoleGroupsResult, error) {
	cErrs, err := composable.AssertRecordAction(ctx, this, cmd, composable.ActionPermission{
		ActionCode: actionManageRoleAssignments, ResourceCode: c.ResourceIamGroup,
	})
	if err != nil || cErrs != nil {
		return mutateRefused(cErrs, err)
	}
	return this.domSvc.ManageGroups(ctx, cmd)
}

// Describe resolves roles into display names; `role_id` may repeat.
func (this *RoleCrudApplicationServiceImpl) Describe(
	ctx corectx.Context, query itRole.DescribeRolesParams,
) (*itRole.DescribeRolesResult, error) {
	_, cErrs := composable.AssertActionPermission(ctx, this, query,
		composable.ActionPermission{ActionCode: composable.PermissionRead})
	if cErrs != nil {
		return &itRole.DescribeRolesResult{ClientErrors: *cErrs}, nil
	}
	return this.domSvc.DescribeRoles(ctx, itRole.DescribeRolesQuery{RoleIds: describedRoleIds(query)})
}

func describedRoleIds(query dmodel.DynamicFields) []model.Id {
	raw, ok := query["role_id"]
	if !ok {
		raw = query["role_ids"]
	}
	switch typed := raw.(type) {
	case string:
		return []model.Id{model.Id(typed)}
	case []string:
		ids := make([]model.Id, 0, len(typed))
		for _, id := range typed {
			ids = append(ids, model.Id(id))
		}
		return ids
	case []any:
		ids := make([]model.Id, 0, len(typed))
		for _, id := range typed {
			ids = append(ids, model.Id(fmt.Sprint(id)))
		}
		return ids
	}
	return nil
}
