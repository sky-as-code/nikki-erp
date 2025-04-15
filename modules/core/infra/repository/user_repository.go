package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/core/domain/user"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/ent"
	entUser "github.com/sky-as-code/nikki-erp/modules/core/infra/ent/user"
)

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

func (r *UserRepository) Create(ctx context.Context, cmd *user.CreateUserCommand) error {
	_, err := r.client.User.Create().
		SetID(cmd.ID).
		SetUsername(cmd.Username).
		SetEmail(cmd.Email).
		SetDisplayName(cmd.DisplayName).
		SetPasswordHash(cmd.Password). // Note: Hash password before storing
		SetAvatarURL(cmd.AvatarURL).
		SetStatus(entUser.Status(cmd.Status)).
		SetMustChangePassword(cmd.MustChangePassword).
		SetCreatedBy(cmd.CreatedBy).
		Save(ctx)

	return err
}

func (r *UserRepository) Update(ctx context.Context, cmd *user.UpdateUserCommand) error {
	return r.client.User.UpdateOneID(cmd.ID).
		SetDisplayName(cmd.DisplayName).
		SetAvatarURL(cmd.AvatarURL).
		SetStatus(entUser.Status(cmd.Status)).
		SetMustChangePassword(cmd.MustChangePassword).
		Exec(ctx)
}

func (r *UserRepository) Delete(ctx context.Context, id string, deletedBy string) error {
	return r.client.User.UpdateOneID(id).
		Exec(ctx)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	u, err := r.client.User.Query().
		Where(entUser.ID(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return mapEntToUser(u), nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	u, err := r.client.User.Query().
		Where(entUser.Username(username)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return mapEntToUser(u), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	u, err := r.client.User.Query().
		Where(entUser.Email(email)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return mapEntToUser(u), nil
}

func mapEntToUser(u *ent.User) *user.User {
	return &user.User{
		ID:                  u.ID,
		Username:            u.Username,
		Email:               u.Email,
		DisplayName:         u.DisplayName,
		PasswordHash:        u.PasswordHash,
		AvatarURL:           u.AvatarURL,
		Status:              string(u.Status),
		CreatedAt:           u.CreatedAt.String(),
		UpdatedAt:           u.UpdatedAt.String(),
		CreatedBy:           u.CreatedBy,
		MustChangePassword:  u.MustChangePassword,
		FailedLoginAttempts: u.FailedLoginAttempts,
		LockedUntil:         timeToStringPtr(u.LockedUntil),
	}
}

func timeToStringPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.String()
	return &s
}
