package user

import (
	"context"

	"go.bryk.io/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	util "github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/ent"
	entUser "github.com/sky-as-code/nikki-erp/modules/core/infra/ent/user"
	it "github.com/sky-as-code/nikki-erp/modules/core/interfaces/user"
)

func NewUserServiceImpl(client *ent.Client) it.UserService {
	return &UserServiceImpl{
		client: client,
	}
}

type UserServiceImpl struct {
	client *ent.Client
}

func (this *UserServiceImpl) CreateUser(ctx context.Context, cmd *it.CreateUserCommand) (*it.CreateUserResult, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		errors.Wrap(err, "failed to hash password")
	}

	creation := this.client.User.Create().
		SetCreatedBy(cmd.CreatedBy).
		SetDisplayName(cmd.DisplayName).
		SetEmail(cmd.Email).
		SetMustChangePassword(cmd.MustChangePassword).
		SetPasswordHash(string(hashedPassword)).
		SetUsername(cmd.Username)

	if cmd.IsEnabled {
		creation.SetStatus(entUser.StatusActive)
	}

	user, err := creation.Save(ctx)
	if err != nil {
		err = errors.Wrap(err, "failed to persist user")
		return nil, err
	}

	return &it.CreateUserResult{
		Id:        &user.ID,
		CreatedAt: &user.CreatedAt,
		Etag:      &user.Etag,
		Status:    util.ToPtr(string(user.Status)),
	}, err
}

func (thisSvc *UserServiceImpl) UpdateUser(ctx context.Context, cmd *it.UpdateUserCommand) error {
	return nil
}

func (thisSvc *UserServiceImpl) DeleteUser(ctx context.Context, id string, deletedBy string) error {
	return nil
}

func (thisSvc *UserServiceImpl) GetUserByID(ctx context.Context, id string) (*it.User, error) {
	return nil, nil
}

func (thisSvc *UserServiceImpl) GetUserByUsername(ctx context.Context, username string) (*it.User, error) {
	return nil, nil
}

func (thisSvc *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*it.User, error) {
	return nil, nil
}
