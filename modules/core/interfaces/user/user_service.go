package user

import (
	"context"
)

type UserService interface {
	CreateUser(ctx context.Context, cmd *CreateUserCommand) (*CreateUserResult, error)
	DeleteUser(ctx context.Context, id string, deletedBy string) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, cmd *UpdateUserCommand) error
}
