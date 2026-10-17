package app

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

func NewRoleApplicationServiceImpl(
	roleSvc it.RoleDomainService,
	roleRepo it.RoleRepository,
) it.RoleAppService {
	return &RoleApplicationServiceImpl{
		roleSvc:  roleSvc,
		roleRepo: roleRepo,
	}
}

type RoleApplicationServiceImpl struct {
	roleSvc  it.RoleDomainService
	roleRepo it.RoleRepository
}

func (this *RoleApplicationServiceImpl) CreateRole(ctx corectx.Context, cmd it.CreateRoleCommand) (*it.CreateRoleResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.CreateRoleResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.CreateRole(ctx, cmd)
}

func (this *RoleApplicationServiceImpl) DeleteRole(ctx corectx.Context, cmd it.DeleteRoleCommand) (*it.DeleteRoleResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.DeleteRoleResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.DeleteRole(ctx, cmd)
}

func (this *RoleApplicationServiceImpl) GetRole(ctx corectx.Context, query it.GetRoleQuery) (*it.GetRoleResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.GetRoleResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.Role, *domain.Role]{
		Action: "get role",
		Schema: this.roleRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.Role], error) {
			return this.roleSvc.GetRole(ctx, query)
		},
	})
}

func (this *RoleApplicationServiceImpl) ManageRoleEntitlements(ctx corectx.Context, cmd it.ManageRoleEntitlementsCommand) (*it.ManageRoleEntitlementsResult, error) {
	if cErr := assertPermission(ctx, "manage_entitlements", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.ManageRoleEntitlementsResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.ManageRoleEntitlements(ctx, cmd)
}

func (this *RoleApplicationServiceImpl) RoleExists(ctx corectx.Context, query it.RoleExistsQuery) (*it.RoleExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.RoleExistsResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.RoleExists(ctx, query)
}

func (this *RoleApplicationServiceImpl) SearchRoles(ctx corectx.Context, query it.SearchRolesQuery) (*it.SearchRolesResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.SearchRolesResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.Role, *domain.Role]{
		Action: "search roles",
		Schema: this.roleRepo.GetBaseRepo().Schema(),
		DefaultFields: []string{
			models.RoleFieldName,
			models.RoleFieldDescription,
			models.RoleFieldIsPrivate,
			fmt.Sprintf("%s.%s", models.RoleEdgeOwnerGroup, models.GroupFieldName),
			fmt.Sprintf("%s.%s", models.RoleEdgeOwnerUser, models.UserFieldDisplayName),
		},
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.Role]], error) {
			return this.roleSvc.SearchRoles(ctx, query, corecrud.ServiceSearchOptions{
				AfterValidationSuccess: fn,
			})
		},
	})
}

// SearchUserRoles and SearchGroupRoles list the roles assigned to one principal. They are
// scoped reads of the role schema, so they gate on the same permission as SearchRoles.
func (this *RoleApplicationServiceImpl) SearchUserRoles(ctx corectx.Context, query it.SearchUserRolesQuery) (*it.SearchUserRolesResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.SearchUserRolesResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.SearchUserRoles(ctx, query)
}

func (this *RoleApplicationServiceImpl) SearchGroupRoles(ctx corectx.Context, query it.SearchGroupRolesQuery) (*it.SearchGroupRolesResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.SearchGroupRolesResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.SearchGroupRoles(ctx, query)
}

func (this *RoleApplicationServiceImpl) DescribeRoles(ctx corectx.Context, query it.DescribeRolesQuery) (*it.DescribeRolesResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.DescribeRolesResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.DescribeRoles(ctx, query)
}

func (this *RoleApplicationServiceImpl) SetRoleIsArchived(ctx corectx.Context, cmd it.SetRoleIsArchivedCommand) (*it.SetRoleIsArchivedResult, error) {
	if cErr := assertPermission(ctx, "set_archived", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.SetRoleIsArchivedResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.SetRoleIsArchived(ctx, cmd)
}

func (this *RoleApplicationServiceImpl) UpdateRole(ctx corectx.Context, cmd it.UpdateRoleCommand) (*it.UpdateRoleResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.UpdateRoleResult{ClientErrors: *cErr}, nil
	}
	if !cmd.HasEntitlementsDelta() {
		return this.roleSvc.UpdateRole(ctx, cmd)
	}
	// Changing what a role grants is the manage_entitlements permission, whichever route
	// carries it, so the role's update permission alone must not be enough.
	if cErr := assertPermission(ctx, "manage_entitlements", c.ResourceIamRole, c.ResourceScopeTenant); cErr != nil {
		return &it.UpdateRoleResult{ClientErrors: *cErr}, nil
	}
	return execRefusableInTranx(ctx, this.roleRepo, func(tranxCtx corectx.Context) (*it.UpdateRoleResult, error) {
		return this.roleSvc.UpdateRole(tranxCtx, cmd)
	})
}

// errRollbackRefusedUpdate unwinds the transaction when a later step refuses the request.
// ExecInTranx commits whenever fn returns no Go error, and a refusal is a result, not an error,
// so without it the role row could commit while the entitlement half was refused.
var errRollbackRefusedUpdate = errors.New("role update refused; rolling back")

func execRefusableInTranx(
	ctx corectx.Context, repo dyn.DynamicModelRepository,
	fn func(tranxCtx corectx.Context) (*it.UpdateRoleResult, error),
) (*it.UpdateRoleResult, error) {
	var refused *it.UpdateRoleResult
	result, err := corecrud.ExecInTranx(ctx, repo, func(tranxCtx corectx.Context) (*it.UpdateRoleResult, error) {
		res, err := fn(tranxCtx)
		if err != nil {
			return nil, err
		}
		if res.ClientErrors.Count() > 0 || !res.HasData {
			refused = res
			return nil, errRollbackRefusedUpdate
		}
		return res, nil
	})
	if errors.Is(err, errRollbackRefusedUpdate) {
		return refused, nil
	}
	return result, err
}

// ManageRoleUsers asserts exactly what POST /users/:user_id/roles asserts, so the role side
// opens no path the principal side keeps shut.
func (this *RoleApplicationServiceImpl) ManageRoleUsers(
	ctx corectx.Context, cmd it.ManageRoleUsersCommand,
) (*it.ManageRoleUsersResult, error) {
	if cErr := assertPermission(ctx, "manage_role_assignments", c.ResourceIamUser, c.ResourceScopeOrg); cErr != nil {
		return &it.ManageRoleUsersResult{ClientErrors: *cErr}, nil
	}
	return corecrud.ExecInTranx(ctx, this.roleRepo, func(tranxCtx corectx.Context) (*it.ManageRoleUsersResult, error) {
		return this.roleSvc.ManageRoleUsers(tranxCtx, cmd)
	})
}

// ManageRoleGroups mirrors POST /groups/:group_id/roles the same way.
func (this *RoleApplicationServiceImpl) ManageRoleGroups(
	ctx corectx.Context, cmd it.ManageRoleGroupsCommand,
) (*it.ManageRoleGroupsResult, error) {
	if cErr := assertPermission(ctx, "manage_role_assignments", c.ResourceIamGroup, c.ResourceScopeOrg); cErr != nil {
		return &it.ManageRoleGroupsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.ExecInTranx(ctx, this.roleRepo, func(tranxCtx corectx.Context) (*it.ManageRoleGroupsResult, error) {
		return this.roleSvc.ManageRoleGroups(tranxCtx, cmd)
	})
}
