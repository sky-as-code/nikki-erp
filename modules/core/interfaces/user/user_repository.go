package user

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, cmd *CreateUserCommand) error
	Update(ctx context.Context, cmd *UpdateUserCommand) error
	Delete(ctx context.Context, cmd *DeleteUserCommand) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}
