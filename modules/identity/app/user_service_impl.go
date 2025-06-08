package app

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	util "github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
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
		err = ft.RecoverPanic(recover(), "failed to create user")
	}()

	user := cmd.ToUser()
	err = user.SetDefaults()
	ft.PanicOnErr(err)
	// this.setUserDefaults(user)

	vErrs := user.Validate(false)
	this.assertUserUnique(ctx, user, &vErrs)
	if vErrs.Count() > 0 {
		return &it.CreateUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	user.PasswordHash = this.encrypt(user.PasswordRaw)
	user, err = this.userRepo.Create(ctx, *user)
	ft.PanicOnErr(err)

	return &it.CreateUserResult{Data: user}, err
}

// func (this *UserServiceImpl) setUserDefaults(user *domain.User) {
// 	id, err := model.NewId()
// 	ft.PanicOnErr(err)
// 	user.Id = id
// 	user.Etag = model.NewEtag()
// 	user.PasswordChangedAt = util.ToPtr(time.Now())

// 	if user.Status == nil {
// 		user.Status = util.ToPtr(domain.UserStatusInactive)
// 	}
// }

func (this *UserServiceImpl) assertUserUnique(ctx context.Context, user *domain.User, errors *ft.ValidationErrors) {
	if errors.Has("email") {
		return
	}
	dbUser, err := this.userRepo.FindByEmail(ctx, *user.Email)
	ft.PanicOnErr(err)

	if dbUser != nil {
		errors.Append("email", "email already exists")
	}
}

func (this *UserServiceImpl) UpdateUser(ctx context.Context, cmd it.UpdateUserCommand) (result *it.UpdateUserResult, err error) {
	defer func() {
		err = ft.RecoverPanic(recover(), "failed to update user")
	}()

	user := cmd.ToUser()

	vErrs := user.Validate(true)

	if vErrs.Count() > 0 {
		return &it.UpdateUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbUser, err := this.userRepo.FindById(ctx, it.FindByIdParam{Id: *user.Id})
	ft.PanicOnErr(err)

	if dbUser == nil {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("id", "user not found")

		return &it.UpdateUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil

	} else if dbUser.Etag.String() != user.Etag.String() {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("etag", "user has been modified by another process")

		return &it.UpdateUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	if user.PasswordRaw != nil {
		user.PasswordHash = this.encrypt(user.PasswordRaw)
	}

	user.Etag = model.NewEtag()
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

func (thisSvc *UserServiceImpl) DeleteUser(ctx context.Context, cmd it.DeleteUserCommand) (result *it.DeleteUserResult, err error) {
	defer func() {
		err = ft.RecoverPanic(recover(), "failed to update user")
	}()

	vErrs := cmd.Validate()
	if vErrs.Count() > 0 {
		return &it.DeleteUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	user, err := thisSvc.userRepo.FindById(ctx, it.FindByIdParam{Id: cmd.Id})
	ft.PanicOnErr(err)

	if user == nil {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("id", "user not found")
		return &it.DeleteUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	err = thisSvc.userRepo.Delete(ctx, it.DeleteUserParam{Id: cmd.Id})
	ft.PanicOnErr(err)

	return &it.DeleteUserResult{
		Data: it.DeleteUserResultData{
			DeletedAt: time.Now(),
		},
	}, nil
}

func (thisSvc *UserServiceImpl) GetUserById(ctx context.Context, query it.GetUserByIdQuery) (result *it.GetUserByIdResult, err error) {
	defer func() {
		err = ft.RecoverPanic(recover(), "failed to get user")
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetUserByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	user, err := thisSvc.userRepo.FindById(ctx, query)
	ft.PanicOnErr(err)

	if user == nil {
		vErrs.Append("id", "user not found")
		return &it.GetUserByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetUserByIdResult{
		Data: user,
	}, nil
}

func (thisSvc *UserServiceImpl) SearchUsers(ctx context.Context, query it.SearchUsersCommand) (result *it.SearchUsersResult, err error) {
	defer func() {
		err = ft.RecoverPanic(recover(), "failed to list users")
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := thisSvc.userRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchUsersResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	users, err := thisSvc.userRepo.Search(ctx, predicate, order, crud.PagingOptions{
		Page: *query.Page,
		Size: *query.Size,
	})
	ft.PanicOnErr(err)

	return &it.SearchUsersResult{
		Data: users,
	}, nil
}
