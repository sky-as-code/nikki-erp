package user

import (
	"context"

	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type UserService interface {
	CreateUser(ctx context.Context, cmd CreateUserCommand) (*CreateUserResult, error)
	DeleteUser(ctx context.Context, id string, deletedBy string) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateUser(ctx context.Context, cmd UpdateUserCommand) error
}
