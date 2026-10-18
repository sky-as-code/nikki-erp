package group

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The composable layers of the iam_group resource, served at /v1/iam/iam_group.

type GroupCrudRepository interface {
	composable.CrudRepository
}

// GroupCrudDomainService is the CRUD of the resource plus its membership and role assignment.
// Both change who holds which permission, so each rebuilds the affected permission caches inside
// its own transaction.
type GroupCrudDomainService interface {
	composable.CrudDomainService

	ManageUsers(ctx corectx.Context, cmd ManageGroupUsersCommand) (*ManageGroupUsersResult, error)
	ManageRoles(ctx corectx.Context, cmd ManageGroupRolesCommand) (*ManageGroupRolesResult, error)
}

type GroupCrudApplicationService interface {
	composable.CrudApplicationService

	ManageUsers(ctx corectx.Context, cmd ManageGroupUsersCommand) (*ManageGroupUsersResult, error)
	ManageRoles(ctx corectx.Context, cmd ManageGroupRolesCommand) (*ManageGroupRolesResult, error)
}

type (
	ManageGroupUsersCommand = composable.ManageM2mCommand
	ManageGroupUsersResult  = composable.MutateResult
	ManageGroupRolesCommand = composable.ManageM2mCommand
	ManageGroupRolesResult  = composable.MutateResult
)
