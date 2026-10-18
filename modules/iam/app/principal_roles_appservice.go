package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itGroup "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

func NewPrincipalRolesApplicationService(
	userSvc itUser.UserCrudApplicationService,
	groupSvc itGroup.GroupCrudApplicationService,
	roleDomSvc itRole.RoleCrudDomainService,
) itRole.PrincipalRolesApplicationService {
	return &PrincipalRolesApplicationServiceImpl{userSvc: userSvc, groupSvc: groupSvc, roleDomSvc: roleDomSvc}
}

// PrincipalRolesApplicationServiceImpl stands apart from the user and group onions because it
// reads the role resource, whose onion depends - through iam_orgunit - on the user onion, so the
// user onion cannot depend on it in return.
type PrincipalRolesApplicationServiceImpl struct {
	userSvc    itUser.UserCrudApplicationService
	groupSvc   itGroup.GroupCrudApplicationService
	roleDomSvc itRole.RoleCrudDomainService
}

// UserRoles is a scoped read of the role schema, so it asserts read on iam_role like a role
// search does, and confines the principal to the caller's org like a read of the principal.
func (this *PrincipalRolesApplicationServiceImpl) UserRoles(
	ctx corectx.Context, query composable.SearchQuery,
) (*composable.SearchResult, error) {
	return this.principalRoles(ctx, this.userSvc, query, models.RoleEdgeAssignedUsers)
}

func (this *PrincipalRolesApplicationServiceImpl) GroupRoles(
	ctx corectx.Context, query composable.SearchQuery,
) (*composable.SearchResult, error) {
	return this.principalRoles(ctx, this.groupSvc, query, models.RoleEdgeAssignedGroups)
}

func (this *PrincipalRolesApplicationServiceImpl) principalRoles(
	ctx corectx.Context, principalSvc composable.CrudApplicationService, query composable.SearchQuery, edge string,
) (*composable.SearchResult, error) {
	tenant := c.ResourceScopeTenant
	cErrs, err := composable.AssertRecordAction(ctx, principalSvc, query, composable.ActionPermission{
		ActionCode: composable.PermissionRead, ResourceCode: c.ResourceIamRole, Scope: &tenant,
	})
	if err != nil || cErrs != nil {
		return searchRefused(cErrs, err)
	}
	return this.roleDomSvc.SearchPrincipalRoles(ctx, query, edge, composable.ParamId(query))
}
