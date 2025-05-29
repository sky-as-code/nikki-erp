package app

import (
	"context"
	"time"

	"go.bryk.io/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	util "github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"

	// entUser "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/user"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewUserServiceImpl(userRepo it.UserRepository) it.UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

type UserServiceImpl struct {
	userRepo it.UserRepository
}

func (this *UserServiceImpl) CreateUser(ctx context.Context, cmd it.CreateUserCommand) (result *it.CreateUserResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to create user")
		}
	}()

	user := cmd.ToUser()
	this.setUserDefaults(user)

	valErr := user.Validate(false)
	this.assertUserUnique(ctx, user, &valErr)
	if valErr.Count() > 0 {
		return &it.CreateUserResult{
			ClientError: ft.WrapValidationErrors(valErr),
		}, nil
	}

	user.PasswordHash = this.encrypt(user.PasswordRaw)
	user, err = this.userRepo.Create(ctx, *user)
	ft.PanicOnErr(err)

	return &it.CreateUserResult{Data: user}, err
}

func (this *UserServiceImpl) setUserDefaults(user *domain.User) {
	id, err := model.NewId()
	ft.PanicOnErr(err)
	user.Id = id
	user.Etag = model.NewEtag()
	user.PasswordChangedAt = util.ToPtr(time.Now())

	if user.Status == nil {
		user.Status = util.ToPtr(domain.UserStatusInactive)
	}
}

func (this *UserServiceImpl) assertUserUnique(ctx context.Context, user *domain.User, errors *ft.ValidationErrors) {
	if errors.Has("email") {
		return
	}
	existingUser, err := this.userRepo.FindByEmail(ctx, *user.Email)
	ft.PanicOnErr(err)

	if existingUser != nil {
		errors.Append(ft.ValidationErrorItem{
			Field: "email",
			Error: "email already exists",
		})
	}
}

func (this *UserServiceImpl) UpdateUser(ctx context.Context, cmd it.UpdateUserCommand) (result *it.UpdateUserResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(r.(error), "failed to update user")
		}
	}()

	user := cmd.ToUser()
	user.Etag = model.NewEtag()

	valErr := user.Validate(true)

	if user.Email != nil {
		this.assertUserUnique(ctx, user, &valErr)
	}

	if valErr.Count() > 0 {
		return &it.UpdateUserResult{
			ClientError: ft.WrapValidationErrors(valErr),
		}, nil
	}

	user.PasswordHash = this.encrypt(user.PasswordRaw)
	user, err = this.userRepo.Update(ctx, *user)
	ft.PanicOnErr(err)

	return &it.UpdateUserResult{Data: user}, err
}

func (this *UserServiceImpl) encrypt(str *string) *string {
	if str == nil {
		return nil
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(*str), bcrypt.DefaultCost)
	ft.PanicOnErr(err)
	return util.ToPtr(string(hashedBytes))
}

func (thisSvc *UserServiceImpl) DeleteUser(ctx context.Context, id string, deletedBy string) error {
	return nil
}

func (thisSvc *UserServiceImpl) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, nil
}

func (thisSvc *UserServiceImpl) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return nil, nil
}

func (thisSvc *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}
