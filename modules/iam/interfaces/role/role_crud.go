package role

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The composable layers of the iam_role resource, served at /v1/iam/iam_role.

type RoleCrudRepository interface {
	composable.CrudRepository
}

// RoleCrudDomainService is the CRUD of the resource plus its rules: an update may carry an
// entitlement delta applied in the same transaction, archiving revokes the role from its holders,
// a private role cannot be deleted, and a system role cannot be assigned.
type RoleCrudDomainService interface {
	composable.CrudDomainService

	ManageUsers(ctx corectx.Context, cmd ManageRoleUsersCommand) (*ManageRoleUsersResult, error)
	ManageGroups(ctx corectx.Context, cmd ManageRoleGroupsCommand) (*ManageRoleGroupsResult, error)
	DescribeRoles(ctx corectx.Context, query DescribeRolesQuery) (*DescribeRolesResult, error)

	// SearchPrincipalRoles lists the roles assigned to one user or group: principalEdge is the
	// role's edge to that principal (assigned_users or assigned_groups).
	SearchPrincipalRoles(
		ctx corectx.Context, query composable.SearchQuery, principalEdge string, principalId model.Id,
	) (*composable.SearchResult, error)
}

type RoleCrudApplicationService interface {
	composable.CrudApplicationService

	ManageUsers(ctx corectx.Context, cmd ManageRoleUsersCommand) (*ManageRoleUsersResult, error)
	ManageGroups(ctx corectx.Context, cmd ManageRoleGroupsCommand) (*ManageRoleGroupsResult, error)
	Describe(ctx corectx.Context, query DescribeRolesParams) (*DescribeRolesResult, error)
}

// PrincipalRolesApplicationService serves GET iam_user/:id/roles and iam_group/:id/roles: the
// roles assigned to one principal, as a scoped read of the role schema.
type PrincipalRolesApplicationService interface {
	UserRoles(ctx corectx.Context, query composable.SearchQuery) (*composable.SearchResult, error)
	GroupRoles(ctx corectx.Context, query composable.SearchQuery) (*composable.SearchResult, error)
}

type (
	ManageRoleUsersCommand  = composable.ManageM2mCommand
	ManageRoleUsersResult   = composable.MutateResult
	ManageRoleGroupsCommand = composable.ManageM2mCommand
	ManageRoleGroupsResult  = composable.MutateResult

	// DescribeRolesParams carries `role_id`, one value or several.
	DescribeRolesParams = composable.SearchQuery
)
