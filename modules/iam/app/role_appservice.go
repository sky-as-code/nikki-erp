package app

import (
	"fmt"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

func NewRoleApplicationServiceImpl(
	roleSvc it.RoleDomainService,
	roleRepo it.RoleRepository,
	userPrefSvc itExt.UserPreferenceUiDomainService,
) it.RoleAppService {
	return &RoleApplicationServiceImpl{
		roleSvc:     roleSvc,
		roleRepo:    roleRepo,
		userPrefSvc: userPrefSvc,
	}
}

type RoleApplicationServiceImpl struct {
	roleSvc     it.RoleDomainService
	roleRepo    it.RoleRepository
	userPrefSvc itExt.UserPreferenceUiDomainService
}

func (this *RoleApplicationServiceImpl) CreateRole(ctx corectx.Context, cmd it.CreateRoleCommand) (*it.CreateRoleResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.CreateRoleResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.CreateRole(ctx, cmd)
}

func (this *RoleApplicationServiceImpl) DeleteRole(ctx corectx.Context, cmd it.DeleteRoleCommand) (*it.DeleteRoleResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.DeleteRoleResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.DeleteRole(ctx, cmd)
}

func (this *RoleApplicationServiceImpl) GetRole(ctx corectx.Context, query it.GetRoleQuery) (*it.GetRoleResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
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
	if cErr := assertPermission(ctx, "manage_entitlements", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.ManageRoleEntitlementsResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.ManageRoleEntitlements(ctx, cmd)
}

func (this *RoleApplicationServiceImpl) RoleExists(ctx corectx.Context, query it.RoleExistsQuery) (*it.RoleExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.RoleExistsResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.RoleExists(ctx, query)
}

func (this *RoleApplicationServiceImpl) SearchRoles(ctx corectx.Context, query it.SearchRolesQuery) (*it.SearchRolesResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.SearchRolesResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.Role, *domain.Role]{
		Action:        "search roles",
		FieldResolver: this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:        this.roleRepo.GetBaseRepo().Schema(),
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
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.SearchUserRolesResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.SearchUserRoles(ctx, query)
}

func (this *RoleApplicationServiceImpl) SearchGroupRoles(ctx corectx.Context, query it.SearchGroupRolesQuery) (*it.SearchGroupRolesResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.SearchGroupRolesResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.SearchGroupRoles(ctx, query)
}

func (this *RoleApplicationServiceImpl) DescribeRoles(ctx corectx.Context, query it.DescribeRolesQuery) (*it.DescribeRolesResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.DescribeRolesResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.DescribeRoles(ctx, query)
}

func (this *RoleApplicationServiceImpl) SetRoleIsArchived(ctx corectx.Context, cmd it.SetRoleIsArchivedCommand) (*it.SetRoleIsArchivedResult, error) {
	if cErr := assertPermission(ctx, "set_archived", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.SetRoleIsArchivedResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.SetRoleIsArchived(ctx, cmd)
}

func (this *RoleApplicationServiceImpl) UpdateRole(ctx corectx.Context, cmd it.UpdateRoleCommand) (*it.UpdateRoleResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceIamRole, c.ResourceScopeDomain); cErr != nil {
		return &it.UpdateRoleResult{ClientErrors: *cErr}, nil
	}
	return this.roleSvc.UpdateRole(ctx, cmd)
}
