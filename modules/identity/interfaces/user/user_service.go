package user

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type UserService interface {
	CreateUser(ctx crud.Context, cmd CreateUserCommand) (*CreateUserResult, error)
	DeleteUser(ctx crud.Context, cmd DeleteUserCommand) (*DeleteUserResult, error)
	Exists(ctx crud.Context, cmd UserExistsQuery) (*UserExistsResult, error)
	ExistsMulti(ctx crud.Context, cmd UserExistsMultiQuery) (*UserExistsMultiResult, error)
	GetUserById(ctx crud.Context, query GetUserByIdQuery) (*GetUserByIdResult, error)
	GetUserByEmail(ctx crud.Context, query GetUserByEmailQuery) (*GetUserByEmailResult, error)
	MustGetActiveUser(ctx crud.Context, query MustGetActiveUserQuery) (*MustGetActiveUserResult, error)
	SearchUsers(ctx crud.Context, query SearchUsersQuery) (*SearchUsersResult, error)
	UpdateUser(ctx crud.Context, cmd UpdateUserCommand) (*UpdateUserResult, error)
	FindDirectApprover(ctx crud.Context, query FindDirectApproverQuery) (*FindDirectApproverResult, error)
}
