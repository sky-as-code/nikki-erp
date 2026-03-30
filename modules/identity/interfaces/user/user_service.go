package user

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type UserService interface {
	GetUserContext(ctx crud.Context, query GetUserContextQuery) (any, error)

	CreateUser(ctx corectx.Context, cmd CreateUserCommand) (*CreateUserResult, error)
	DeleteUser(ctx corectx.Context, cmd DeleteUserCommand) (*DeleteUserResult, error)
	GetActiveUser(ctx corectx.Context, query GetUserQuery) (*GetUserResult, error)
	GetUser(ctx corectx.Context, query GetUserQuery) (*GetUserResult, error)
	SearchUsers(ctx corectx.Context, query SearchUsersQuery) (*SearchUsersResult, error)
	SetUserIsArchived(ctx corectx.Context, cmd SetUserIsArchivedCommand) (*SetUserIsArchivedResult, error)
	UserExists(ctx corectx.Context, query UserExistsQuery) (*UserExistsResult, error)
	UpdateUser(ctx corectx.Context, cmd UpdateUserCommand) (*UpdateUserResult, error)
}
