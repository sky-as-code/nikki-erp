package user

import (
	"context"
)

type UserService interface {
	CreateUser(ctx context.Context, cmd CreateUserCommand) (*CreateUserResult, error)
	DeleteUser(ctx context.Context, cmd DeleteUserCommand) (*DeleteUserResult, error)
	GetUserById(ctx context.Context, query GetUserByIdQuery) (*GetUserByIdResult, error)
	SearchUsers(ctx context.Context, query SearchUsersCommand) (*SearchUsersResult, error)
	UpdateUser(ctx context.Context, cmd UpdateUserCommand) (*UpdateUserResult, error)
}
