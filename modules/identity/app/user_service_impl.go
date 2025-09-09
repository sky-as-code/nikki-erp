package app

import (
	"context"
	"strings"

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
			return this.assertUserUnique(ctx, user.Email, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	user, err = this.userRepo.Create(ctx, *user)
	ft.PanicOnErr(err)

	// TODO: If OrgIds is specified, add this user to the orgs

	return &it.CreateUserResult{
		Data:    user,
		HasData: user != nil, // In rare case, new user was deleted after created. Maybe due to a concurrent cleanup process.
	}, err
}

func (this *UserServiceImpl) UpdateUser(ctx context.Context, cmd it.UpdateUserCommand) (result *it.UpdateUserResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "update user"); e != nil {
			err = e
		}
	}()

	user := cmd.ToUser()

	var dbUser *domain.User
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = user.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbUser, err = this.assertUserIdExists(ctx, *user.Id, nil, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertCorrectEtag(*user.Etag, *dbUser.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			// Sanitize after we've made sure this is the correct user
			this.sanitizeUser(user)
			return this.assertUserUnique(ctx, user.Email, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := user.Etag
	user.Etag = model.NewEtag()
	user, err = this.userRepo.Update(ctx, *user, *prevEtag)
	ft.PanicOnErr(err)

	return &it.UpdateUserResult{
		Data:    user,
		HasData: user != nil,
	}, err
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
	user.Status = util.ToPtr(domain.UserStatusActive)
}

func (this *UserServiceImpl) assertUserUnique(ctx context.Context, newEmail *string, vErrs *ft.ValidationErrors) error {
	if newEmail == nil {
		return nil
	}
	dbUser, err := this.userRepo.FindByEmail(ctx, it.FindByEmailParam{Email: *newEmail})
	if err != nil {
		return err
	}

	if dbUser != nil {
		vErrs.AppendAlreadyExists("email", "email")
	}
	return nil
}

func (this *UserServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *ft.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *UserServiceImpl) assertUserIdExists(ctx context.Context, id model.Id, status *domain.UserStatus, vErrs *ft.ValidationErrors) (dbUser *domain.User, err error) {
	dbUser, err = this.userRepo.FindById(ctx, it.FindByIdParam{Id: id, Status: status})
	if dbUser == nil {
		vErrs.AppendNotFound("id", "user id")
	}
	return
}

func (this *UserServiceImpl) assertUserEmailExists(ctx context.Context, email string, status *domain.UserStatus, vErrs *ft.ValidationErrors) (dbUser *domain.User, err error) {
	dbUser, err = this.userRepo.FindByEmail(ctx, it.FindByEmailParam{Email: email, Status: status})
	if dbUser == nil {
		vErrs.AppendNotFound("email", "user email")
	}
	return
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

	deletedCount, err := this.userRepo.DeleteHard(ctx, it.DeleteParam{Id: cmd.Id})
	ft.PanicOnErr(err)
	if deletedCount == 0 {
		vErrs.AppendNotFound("id", "user id")
		return &it.DeleteUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount), nil
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
		if e := ft.RecoverPanicFailedTo(recover(), "get user by Id"); e != nil {
			err = e
		}
	}()

	var dbUser *domain.User
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbUser, err = this.assertUserIdExists(ctx, query.Id, query.Status, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetUserByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetUserByIdResult{
		Data:    dbUser,
		HasData: dbUser != nil,
	}, nil
}

func (this *UserServiceImpl) MustGetActiveUser(ctx context.Context, query it.MustGetActiveUserQuery) (result *it.MustGetActiveUserResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "get active user"); e != nil {
			err = e
		}
	}()

	var dbUser *domain.User
	var field string
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if query.Id == nil {
				return nil
			}
			field = "id"
			dbUser, err = this.assertUserIdExists(ctx, *query.Id, nil, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if query.Email == nil {
				return nil
			}
			field = "email"
			*query.Email = strings.ToLower(*query.Email)
			dbUser, err = this.assertUserEmailExists(ctx, *query.Email, nil, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if dbUser != nil {
		switch *dbUser.Status {
		case domain.UserStatusArchived:
			vErrs.AppendNotFound(field, "user is archived")
		case domain.UserStatusLocked:
			vErrs.AppendNotFound(field, "user is locked")
		}
	}

	if vErrs.Count() > 0 {
		return &it.MustGetActiveUserResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.MustGetActiveUserResult{
		Data:    dbUser,
		HasData: true,
	}, nil
}

func (this *UserServiceImpl) GetUserByEmail(ctx context.Context, query it.GetUserByEmailQuery) (result *it.GetUserByEmailResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "get user by email"); e != nil {
			err = e
		}
	}()

	var dbUser *domain.User
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			query.Email = strings.ToLower(query.Email)
			dbUser, err = this.userRepo.FindByEmail(ctx, it.FindByEmailParam{Email: query.Email, Status: query.Status})
			if err != nil {
				return err
			}
			if dbUser == nil {
				vErrs.AppendNotFound("email", "user email")
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetUserByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetUserByIdResult{
		Data:    dbUser,
		HasData: dbUser != nil,
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
		HasData: users.Items != nil,
	}, nil
}

func (this *UserServiceImpl) FindDirectApprover(ctx context.Context, query it.FindDirectApproverQuery) (result *it.FindDirectApproverResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "find direct approver"); e != nil {
			err = e
		}
	}()

	var dbUser *domain.User
	var approver *domain.User

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbUser, err = this.assertUserExists(ctx, query.Id, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if dbUser.HierarchyId == nil {
				return nil
			}

			approver, err = this.findApproverInParentHierarchy(ctx, dbUser, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.FindDirectApproverResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.FindDirectApproverResult{
		Data:    approver,
		HasData: approver != nil,
	}, nil
}

func (this *UserServiceImpl) assertUserExists(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (user *domain.User, err error) {
	user, err = this.userRepo.FindById(ctx, it.FindByIdParam{Id: id})
	ft.PanicOnErr(err)

	if user == nil {
		vErrs.AppendNotFound("user_id", "user")
	}
	return user, err
}

func (this *UserServiceImpl) findApproverInParentHierarchy(ctx context.Context, dbUser *domain.User, vErrs *ft.ValidationErrors) (*domain.User, error) {
	directManager, err := this.userRepo.FindUsersByHierarchyId(ctx, it.FindByHierarchyIdParam{HierarchyId: *dbUser.HierarchyId})
	ft.PanicOnErr(err)

	if len(directManager) == 0 {
		vErrs.AppendNotFound("hierarchy_id", "hierarchy")
		return nil, nil
	}
	return &directManager[0], nil
}
