package app

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	util "github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewUserServiceImpl(
	enumSvc enum.EnumService,
	userRepo it.UserRepository,
	eventBus event.EventBus,
) it.UserService {
	return &UserServiceImpl{
		enumSvc:  enumSvc,
		userRepo: userRepo,
		eventBus: eventBus,
	}
}

type UserServiceImpl struct {
	enumSvc  enum.EnumService
	userRepo it.UserRepository
	eventBus event.EventBus
}

func (this *UserServiceImpl) CreateUser(ctx context.Context, cmd it.CreateUserCommand) (result *it.CreateUserResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "create user"); e != nil {
			err = e
		}
	}()

	user := cmd.ToUser()
	this.setUserDefaults(ctx, user)

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = user.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeUser(user)
			return this.assertUserUnique(ctx, user, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	user.PasswordHash = this.encrypt(user.PasswordRaw)
	user, err = this.userRepo.Create(ctx, *user)
	ft.PanicOnErr(err)

	// TODO: If OrgIds is specified, add this user to the orgs

	return &it.CreateUserResult{
		Data:    user,
		HasData: true,
	}, err
}

func (this *UserServiceImpl) UpdateUser(ctx context.Context, cmd it.UpdateUserCommand) (result *it.UpdateUserResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "update user"); e != nil {
			err = e
		}
	}()

	user := cmd.ToUser()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = user.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertCorrectUser(ctx, user, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			// Sanitize after we've made sure this is the correct user
			this.sanitizeUser(user)

			if user.Email != nil {
				return this.assertUserUnique(ctx, user, vErrs)
			}
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if user.StatusId != nil || user.StatusValue != nil {
				dbStatus := this.assertStatusExists(ctx, user, vErrs)
				user.StatusId = dbStatus.Data.Id
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	if user.PasswordRaw != nil {
		user.PasswordHash = this.encrypt(user.PasswordRaw)
	}

	prevEtag := user.Etag
	user.Etag = model.NewEtag()
	user, err = this.userRepo.Update(ctx, *user, *prevEtag)
	ft.PanicOnErr(err)

	return &it.UpdateUserResult{
		Data:    user,
		HasData: true,
	}, err
}

func (this *UserServiceImpl) encrypt(str *string) *string {
	if str == nil {
		return nil
	}
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(*str), bcrypt.DefaultCost)
	ft.PanicOnErr(err)
	return util.ToPtr(string(hashedBytes))
}

func (this *UserServiceImpl) sanitizeUser(user *domain.User) {
	if user.Email != nil {
		user.Email = util.ToPtr(strings.ToLower(*user.Email))
	}
	if user.DisplayName != nil {
		cleanedName := strings.TrimSpace(*user.DisplayName)
		cleanedName = defense.SanitizePlainText(cleanedName)
		user.DisplayName = &cleanedName
	}
}

func (this *UserServiceImpl) setUserDefaults(ctx context.Context, user *domain.User) {
	user.SetDefaults()

	activeEnum, err := this.enumSvc.GetEnum(ctx, enum.GetEnumQuery{
		Value: util.ToPtr(domain.UserStatusActive),
		Type:  util.ToPtr(domain.UserStatusEnumType),
	})
	ft.PanicOnErr(err)
	ft.PanicOnErr(activeEnum.ClientError)

	user.Status = domain.WrapIdentStatus(activeEnum.Data)
	user.StatusId = activeEnum.Data.Id
}

func (this *UserServiceImpl) assertUserUnique(ctx context.Context, user *domain.User, vErrs *ft.ValidationErrors) error {
	dbUser, err := this.userRepo.FindByEmail(ctx, *user.Email)
	if err != nil {
		return err
	}

	if dbUser != nil {
		vErrs.AppendAlreadyExists("email", "email")
	}
	return nil
}

func (this *UserServiceImpl) assertStatusExists(ctx context.Context, user *domain.User, vErrs *ft.ValidationErrors) *enum.GetEnumResult {
	dbStatus, err := this.enumSvc.GetEnum(ctx, enum.GetEnumQuery{
		Id:         user.StatusId,
		Value:      user.StatusValue,
		Type:       util.ToPtr(domain.UserStatusEnumType),
		EntityName: "user status",
	})

	ft.PanicOnErr(err)
	ft.PanicOnErr(dbStatus.ClientError)

	if !dbStatus.HasData {
		vErrs.Append("status", "invalid user status")
		return nil
	}
	return dbStatus
}

func (this *UserServiceImpl) assertCorrectUser(ctx context.Context, user *domain.User, vErrs *ft.ValidationErrors) error {
	dbUser, err := this.userRepo.FindById(ctx, it.FindByIdParam{Id: *user.Id})
	if err != nil {
		return err
	}

	if dbUser == nil {
		vErrs.AppendIdNotFound("user")
		return nil
	} else if *dbUser.Etag != *user.Etag {
		vErrs.AppendEtagMismatched()
		return nil
	}
	return nil
}

func (this *UserServiceImpl) DeleteUser(ctx context.Context, cmd it.DeleteUserCommand) (result *it.DeleteUserResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "delete user"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	if vErrs.Count() > 0 {
		return &it.DeleteUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	user, err := this.userRepo.FindById(ctx, it.FindByIdParam{Id: cmd.Id})
	ft.PanicOnErr(err)

	if user == nil {
		vErrs.AppendIdNotFound("user")
		return &it.DeleteUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	err = this.userRepo.Delete(ctx, it.DeleteParam{Id: cmd.Id})
	ft.PanicOnErr(err)

	return crud.NewSuccessDeletionResult(*user.Id), nil
}

func (this *UserServiceImpl) Exists(ctx context.Context, cmd it.UserExistsCommand) (result *it.UserExistsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "check if user exists"); e != nil {
			err = e
		}
	}()

	exists, err := this.userRepo.Exists(ctx, cmd.Id)
	ft.PanicOnErr(err)

	return &it.UserExistsResult{
		Data:    exists,
		HasData: true,
	}, nil
}

func (this *UserServiceImpl) ExistsMulti(ctx context.Context, cmd it.UserExistsMultiCommand) (result *it.UserExistsMultiResult, err error) {
	exists, notExisting, err := this.userRepo.ExistsMulti(ctx, cmd.Ids)
	ft.PanicOnErr(err)

	return &it.UserExistsMultiResult{
		Data: &it.ExistsMultiResultData{
			Existing:    exists,
			NotExisting: notExisting,
		},
		HasData: true,
	}, nil
}

func (this *UserServiceImpl) GetUserById(ctx context.Context, query it.GetUserByIdQuery) (result *it.GetUserByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "get user"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetUserByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	user, err := this.userRepo.FindById(ctx, query)
	ft.PanicOnErr(err)

	if user == nil {
		vErrs.AppendIdNotFound("user")
		return &it.GetUserByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetUserByIdResult{
		Data:    user,
		HasData: true,
	}, nil
}

func (this *UserServiceImpl) SearchUsers(ctx context.Context, query it.SearchUsersQuery) (result *it.SearchUsersResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "list users"); e != nil {
			err = e
		}
	}()

	query.SetDefaults()
	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.userRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchUsersResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	users, err := this.userRepo.Search(ctx, it.SearchParam{
		Predicate:  predicate,
		Order:      order,
		Page:       *query.Page,
		Size:       *query.Size,
		WithGroups: query.WithGroups,
	})
	ft.PanicOnErr(err)

	return &it.SearchUsersResult{
		Data:    users,
		HasData: len(users.Items) > 0,
	}, nil
}

func (this *UserServiceImpl) ListUserStatuses(ctx context.Context, query it.ListUserStatusesQuery) (*it.ListIdentStatusesResult, error) {
	result, err := this.enumSvc.ListEnums(ctx, enum.ListEnumsQuery{
		EntityName: "user statuses",
		Type:       util.ToPtr(domain.UserStatusEnumType),
		Page:       query.Page,
		Size:       query.Size,
		SortByLang: query.SortByLang,
	})
	return (*it.ListIdentStatusesResult)(result), err
}
