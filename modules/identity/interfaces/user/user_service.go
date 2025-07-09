package user

import (
	"context"
)

type UserService interface {
	CreateUser(ctx context.Context, cmd CreateUserCommand) (*CreateUserResult, error)
	DeleteUser(ctx context.Context, cmd DeleteUserCommand) (*DeleteUserResult, error)
	Exists(ctx context.Context, cmd UserExistsCommand) (*UserExistsResult, error)
	ExistsMulti(ctx context.Context, cmd UserExistsMultiCommand) (*UserExistsMultiResult, error)
	GetUserById(ctx context.Context, query GetUserByIdQuery) (*GetUserByIdResult, error)
	ListUserStatuses(ctx context.Context, query ListUserStatusesQuery) (*ListUserStatusesResult, error)
	SearchUsers(ctx context.Context, query SearchUsersQuery) (*SearchUsersResult, error)
	UpdateUser(ctx context.Context, cmd UpdateUserCommand) (*UpdateUserResult, error)
}
