package group

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The composable layers of the iam_group resource, served at /v1/iam/iam_group.

type GroupCrudRepository interface {
	dynamicresource.CrudRepository
}

// GroupCrudDomainService is the CRUD of the resource plus its membership and role assignment.
// Both change who holds which permission, so each rebuilds the affected permission caches inside
// its own transaction.
type GroupCrudDomainService interface {
	dynamicresource.CrudDomainService

	ManageUsers(ctx corectx.Context, cmd ManageGroupUsersCommand) (*ManageGroupUsersResult, error)
	ManageRoles(ctx corectx.Context, cmd ManageGroupRolesCommand) (*ManageGroupRolesResult, error)
}

type GroupCrudApplicationService interface {
	dynamicresource.CrudApplicationService

	ManageUsers(ctx corectx.Context, cmd ManageGroupUsersCommand) (*ManageGroupUsersResult, error)
	ManageRoles(ctx corectx.Context, cmd ManageGroupRolesCommand) (*ManageGroupRolesResult, error)
}

type (
	ManageGroupUsersCommand = dynamicresource.ManageM2mCommand
	ManageGroupUsersResult  = dynamicresource.MutateResult
	ManageGroupRolesCommand = dynamicresource.ManageM2mCommand
	ManageGroupRolesResult  = dynamicresource.MutateResult
)
