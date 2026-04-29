package user

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type UserDomainService interface {
	GetUserContext(ctx crud.Context, query GetUserContextQuery) (any, error)

	CreateUser(ctx corectx.Context, cmd CreateUserCommand, opts ...corecrud.ServiceCreateOptions[*domain.User]) (*CreateUserResult, error)
	DeleteUser(ctx corectx.Context, cmd DeleteUserCommand, opts ...corecrud.ServiceDeleteOptions) (*DeleteUserResult, error)
	GetEnabledUser(ctx corectx.Context, query GetUserQuery) (*dyn.OpResult[domain.User], error)
	GetUser(ctx corectx.Context, query GetUserQuery) (*dyn.OpResult[domain.User], error)
	SearchUsers(ctx corectx.Context, query SearchUsersQuery, opts ...corecrud.ServiceSearchOptions) (*dyn.OpResult[dyn.PagedResultData[domain.User]], error)
	SetUserIsArchived(ctx corectx.Context, cmd SetUserIsArchivedCommand) (*SetUserIsArchivedResult, error)
	UserExists(ctx corectx.Context, query UserExistsQuery) (*UserExistsResult, error)
	UpdateUser(ctx corectx.Context, cmd UpdateUserCommand, opts ...corecrud.ServiceUpdateOptions[*domain.User]) (*UpdateUserResult, error)
}

type UserAppService interface {
	CreateUser(ctx corectx.Context, cmd CreateUserCommand) (*CreateUserResult, error)
	DeleteUser(ctx corectx.Context, cmd DeleteUserCommand) (*DeleteUserResult, error)
	GetEnabledUser(ctx corectx.Context, query GetUserQuery) (*GetUserResult, error)
	GetUser(ctx corectx.Context, query GetUserQuery) (*GetUserResult, error)
	SearchUsers(ctx corectx.Context, query SearchUsersQuery) (*SearchUsersResult, error)
	SetUserIsArchived(ctx corectx.Context, cmd SetUserIsArchivedCommand) (*SetUserIsArchivedResult, error)
	UserExists(ctx corectx.Context, query UserExistsQuery) (*UserExistsResult, error)
	UpdateUser(ctx corectx.Context, cmd UpdateUserCommand) (*dyn.OpResult[dyn.MutateResultData], error)
}
