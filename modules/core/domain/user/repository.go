package user

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, user *CreateUserCommand) error
	Update(ctx context.Context, user *UpdateUserCommand) error
	Delete(ctx context.Context, id string, deletedBy string) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}
