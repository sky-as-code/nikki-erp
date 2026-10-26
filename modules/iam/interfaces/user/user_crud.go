package user

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The composable layers of the iam_user resource, served at /v1/iam/iam_user. Sign-in, permission
// evaluation, /me/context and the command bus read users through the same layers.

// UserCrudRepository reads and writes user rows for the composable resource.
type UserCrudRepository interface {
	composable.CrudRepository
}

// UserCrudDomainService is the CRUD of the resource plus its own rules: a create seeds the
// user's preferences and permission cache, and a role assignment rebuilds that cache.
type UserCrudDomainService interface {
	composable.CrudDomainService

	ManageRoles(ctx corectx.Context, cmd ManageUserRolesCommand) (*ManageUserRolesResult, error)

	// GetUser and GetEnabledUser are the sign-in and request-guard reads. They assert nothing, and
	// GetEnabledUser answers the getEnabledUser command bus query, whose shape core copies.
	GetUser(ctx corectx.Context, query GetUserQuery) (*GetUserDomainResult, error)
	GetEnabledUser(ctx corectx.Context, query GetUserQuery) (*GetUserDomainResult, error)
}

// UserCrudApplicationService is the authorized surface the REST handler serves.
type UserCrudApplicationService interface {
	composable.CrudApplicationService

	SendInvitation(ctx corectx.Context, cmd SendUserInvitationCommand) (*SendUserInvitationResult, error)
	ManageRoles(ctx corectx.Context, cmd ManageUserRolesCommand) (*ManageUserRolesResult, error)
}

type (
	CreateUserCrudCommand = composable.CreateCommand
	CreateUserCrudResult  = composable.CreateResult

	SendUserInvitationCommand = dmodel.DynamicFields
	SendUserInvitationResult  = dyn.OpResult[any]

	ManageUserRolesCommand = composable.ManageM2mCommand
	ManageUserRolesResult  = composable.MutateResult
)
