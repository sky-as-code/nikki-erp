package app

import (
	"context"
	"strings"
	"time"

	"go.bryk.io/pkg/ulid"
	"golang.org/x/crypto/bcrypt"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	util "github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/enum"
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
		if e := ft.RecoverPanic(recover(), "failed to create user"); e != nil {
			err = e
		}
	}()

	user := cmd.ToUser()
	this.sanitizeUser(user)
	this.setUserDefaults(ctx, user)

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

	// TODO: If OrgIds is specified, add this user to the orgs

	return &it.CreateUserResult{
		Data:    user,
		HasData: true,
	}, err
}

func (this *UserServiceImpl) sanitizeUser(user *domain.User) {
	if user.Email != nil {
		user.Email = util.ToPtr(strings.ToLower(*user.Email))
	}
	if user.DisplayName != nil {
		user.DisplayName = util.ToPtr(strings.TrimSpace(*user.DisplayName))
	}
	user.Etag = model.NewEtag()
}

func (this *UserServiceImpl) setUserDefaults(ctx context.Context, user *domain.User) {
	err := user.SetDefaults()
	ft.PanicOnErr(err)
	activeEnum, err := this.enumSvc.GetEnum(ctx, enum.GetEnumQuery{
		Value:    util.ToPtr(domain.UserStatusActive),
		EnumType: util.ToPtr(domain.UserStatusEnumType),
	})
	ft.PanicOnErr(err)
	ft.PanicOnErr(activeEnum.ClientError)

	user.Status = domain.WrapUserStatus(activeEnum.Data)
	user.StatusId = activeEnum.Data.Id
	user.Email = util.ToPtr(strings.ToLower(*user.Email))
}

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

func (this *UserServiceImpl) assertStatusExists(ctx context.Context, user *domain.User, errors *ft.ValidationErrors) {
	if errors.Has("status") || (user.StatusValue == nil) {
		return
	}
	dbStatus, err := this.enumSvc.GetEnum(ctx, enum.GetEnumQuery{
		Value:    user.StatusValue,
		EnumType: util.ToPtr(domain.UserStatusEnumType),
	})
	ft.PanicOnErr(err)
	ft.PanicOnErr(dbStatus.ClientError)

	if !dbStatus.HasData {
		errors.Append("status", "invalid user status")
		return
	}
	user.StatusId = dbStatus.Data.Id
}

func (this *UserServiceImpl) UpdateUser(ctx context.Context, cmd it.UpdateUserCommand) (result *it.UpdateUserResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update user"); e != nil {
			err = e
		}
	}()

	user := cmd.ToUser()
	this.sanitizeUser(user)

	vErrs := user.Validate(true)
	if user.Email != nil {
		this.assertUserUnique(ctx, user, &vErrs)
	}
	if user.StatusId != nil {
		this.assertStatusExists(ctx, user, &vErrs)
		// user.Status = this.enumRepo.FindById(ctx, *user.StatusId)
	}
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

	} else if *dbUser.Etag != *user.Etag {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("etag", "user has been modified by another process")

		return &it.UpdateUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	if user.PasswordRaw != nil {
		user.PasswordHash = this.encrypt(user.PasswordRaw)
	}

	user, err = this.userRepo.Update(ctx, *user)
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

func (this *UserServiceImpl) DeleteUser(ctx context.Context, cmd it.DeleteUserCommand) (result *it.DeleteUserResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to delete user"); e != nil {
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
		vErrs.Append("id", "user not found")
		return &it.DeleteUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	err = this.userRepo.Delete(ctx, it.DeleteParam{Id: cmd.Id})
	ft.PanicOnErr(err)
	// Publish user deleted event
	err = this.publishUserDeletedEvent(ctx, user)
	ft.PanicOnErr(err)

	return &it.DeleteUserResult{
		Data: &it.DeleteUserResultData{
			DeletedAt: time.Now(),
		},
		HasData: true,
	}, nil
}

// publishUserDeletedEvent publishes a "user.deleted.done" event
func (this *UserServiceImpl) publishUserDeletedEvent(ctx context.Context, user *domain.User) error {
	// Generate event ID
	eventId, err := ulid.New()
	if err != nil {
		return err
	}

	// Create the event payload
	userDeletedEvent := &it.UserDeletedEvent{
		ID:        *user.Id,
		DeletedBy: "", // You might want to get this from context or pass it as parameter
		EventID:   eventId.String(),
	}

	// Use the interface method to publish the event
	return this.eventBus.PublishEvent(ctx, "user.deleted.done", userDeletedEvent)
}

func (this *UserServiceImpl) Exists(ctx context.Context, cmd it.UserExistsCommand) (result *it.UserExistsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to check if user exists"); e != nil {
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
		if e := ft.RecoverPanic(recover(), "failed to get user"); e != nil {
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
		vErrs.Append("id", "user not found")
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
		if e := ft.RecoverPanic(recover(), "failed to list users"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.userRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchUsersResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

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

func (this *UserServiceImpl) ListUserStatuses(ctx context.Context, query it.ListUserStatusesQuery) (result *it.ListUserStatusesResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list user statuses"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	if vErrsModel.Count() > 0 {
		return &it.ListUserStatusesResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()
	userStatuses, err := this.enumSvc.ListEnums(ctx, enum.ListEnumsQuery{
		EnumType:     util.ToPtr(domain.UserStatusEnumType),
		Page:         query.Page,
		Size:         query.Size,
		SortedByLang: query.SortedByLang,
	})
	ft.PanicOnErr(err)

	result = &it.ListUserStatusesResult{
		ClientError: userStatuses.ClientError,
	}

	if result.ClientError == nil {
		result.HasData = userStatuses.HasData
		result.Data = &it.ListUserStatusesResultData{
			Items: domain.WrapUserStatuses(userStatuses.Data.Items),
			Total: userStatuses.Data.Total,
		}
	}

	return result, nil
}
