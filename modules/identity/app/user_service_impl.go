package app

import (
	"strings"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	util "github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itHierarchy "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewUserServiceImpl(
	enumSvc enum.EnumService,
	userRepo it.UserRepository,
	hierarchyRepo itHierarchy.HierarchyRepository,
	eventBus event.EventBus,
) it.UserService {
	return &UserServiceImpl{
		enumSvc:       enumSvc,
		userRepo:      userRepo,
		hierarchyRepo: hierarchyRepo,
		eventBus:      eventBus,
	}
}

type UserServiceImpl struct {
	enumSvc       enum.EnumService
	userRepo      it.UserRepository
	hierarchyRepo itHierarchy.HierarchyRepository
	eventBus      event.EventBus
}

func (this *UserServiceImpl) CreateUser(ctx crud.Context, cmd it.CreateUserCommand) (*it.CreateUserResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.User, it.CreateUserCommand, it.CreateUserResult]{
		Action:              "create user",
		Command:             cmd,
		AssertBusinessRules: this.assertUserUnique,
		RepoCreate:          this.userRepo.Create,
		SetDefault:          this.setUserDefaults,
		Sanitize:            this.sanitizeUser,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateUserResult {
			return &it.CreateUserResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.User) *it.CreateUserResult {
			return &it.CreateUserResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
	// defer func() {
	// 	if e := ft.RecoverPanicFailedTo(recover(), "create user"); e != nil {
	// 		err = e
	// 	}
	// }()

	// user := cmd.ToUser()
	// this.setUserDefaults(ctx, user)

	// flow := val.StartValidationFlow()
	// vErrs, err := flow.
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		*vErrs = user.Validate(false)
	// 		return nil
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		this.sanitizeUser(user)
	// 		return this.assertUserUnique(ctx, user.Email, vErrs)
	// 	}).
	// 	End()
	// ft.PanicOnErr(err)

	// if vErrs.Count() > 0 {
	// 	return &it.CreateUserResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// user, err = this.userRepo.Create(ctx, *user)
	// ft.PanicOnErr(err)

	// // TODO: If OrgIds is specified, add this user to the orgs

	// return &it.CreateUserResult{
	// 	Data:    user,
	// 	HasData: user != nil, // In rare case, new user was deleted after created. Maybe due to a concurrent cleanup process.
	// }, err
}

func (this *UserServiceImpl) UpdateUser(ctx crud.Context, cmd it.UpdateUserCommand) (*it.UpdateUserResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.User, it.UpdateUserCommand, it.UpdateUserResult]{
		Action:              "update user",
		Command:             cmd,
		AssertBusinessRules: this.assertUpdateRules,
		AssertExists:        this.assertUserIdExists,
		RepoUpdate:          this.userRepo.Update,
		Sanitize:            this.sanitizeUser,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateUserResult {
			return &it.UpdateUserResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.User) *it.UpdateUserResult {
			return &it.UpdateUserResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err

	// defer func() {
	// 	if e := ft.RecoverPanicFailedTo(recover(), "update user"); e != nil {
	// 		err = e
	// 	}
	// }()

	// user := cmd.ToUser()

	// var dbUser *domain.User
	// flow := val.StartValidationFlow()
	// vErrs, err := flow.
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		*vErrs = user.Validate(true)
	// 		return nil
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		dbUser, err = this.assertUserIdExists(ctx, *user.Id, nil, vErrs)
	// 		return err
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		this.assertCorrectEtag(*user.Etag, *dbUser.Etag, vErrs)
	// 		return nil
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		// Sanitize after we've made sure this is the correct user
	// 		this.sanitizeUser(user)
	// 		return this.assertUserUnique(ctx, user.Email, vErrs)
	// 	}).
	// 	End()
	// ft.PanicOnErr(err)

	// if vErrs.Count() > 0 {
	// 	return &it.UpdateUserResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// prevEtag := user.Etag
	// user.Etag = model.NewEtag()
	// user, err = this.userRepo.Update(ctx, *user, *prevEtag)
	// ft.PanicOnErr(err)

	// return &it.UpdateUserResult{
	// 	Data:    user,
	// 	HasData: user != nil,
	// }, err
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

func (this *UserServiceImpl) setUserDefaults(user *domain.User) {
	user.SetDefaults()
	user.Status = util.ToPtr(domain.UserStatusActive)
}

func (this *UserServiceImpl) assertUpdateRules(ctx crud.Context, updatedUser *domain.User, _ *domain.User, vErrs *ft.ValidationErrors) error {
	return this.assertUserUnique(ctx, updatedUser, vErrs)
}

func (this *UserServiceImpl) assertUserUnique(ctx crud.Context, user *domain.User, vErrs *ft.ValidationErrors) error {
	if user.Email == nil {
		return nil
	}
	dbUser, err := this.userRepo.FindByEmail(ctx, it.FindByEmailParam{Email: *user.Email})
	if err != nil {
		return err
	}

	if dbUser != nil {
		vErrs.AppendAlreadyExists("email", "email")
	}
	return nil
}

// func (this *UserServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *ft.ValidationErrors) {
// 	if updatedEtag != dbEtag {
// 		vErrs.AppendEtagMismatched()
// 	}
// }

func (this *UserServiceImpl) assertUserIdExists(ctx crud.Context, user *domain.User, vErrs *ft.ValidationErrors) (dbUser *domain.User, err error) {
	dbUser, err = this.userRepo.FindById(ctx, it.FindByIdParam{Id: *user.Id})
	if dbUser == nil {
		vErrs.AppendNotFound("id", "user id")
	}
	return
}

func (this *UserServiceImpl) assertUserEmailExists(ctx crud.Context, email string, vErrs *ft.ValidationErrors) (dbUser *domain.User, err error) {
	dbUser, err = this.userRepo.FindByEmail(ctx, it.FindByEmailParam{Email: email})
	if dbUser == nil {
		vErrs.AppendNotFound("email", "user email")
	}
	return
}

func (this *UserServiceImpl) DeleteUser(ctx crud.Context, cmd it.DeleteUserCommand) (*it.DeleteUserResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.User, it.DeleteUserCommand, it.DeleteUserResult]{
		Action:       "delete user",
		Command:      cmd,
		AssertExists: this.assertUserIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.User) (int, error) {
			return this.userRepo.DeleteHard(ctx, it.DeleteParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteUserResult {
			return &it.DeleteUserResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.User, deletedCount int) *it.DeleteUserResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})

	return result, err

	// defer func() {
	// 	if e := ft.RecoverPanicFailedTo(recover(), "delete user"); e != nil {
	// 		err = e
	// 	}
	// }()

	// vErrs := cmd.Validate()

	// if vErrs.Count() > 0 {
	// 	return &it.DeleteUserResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// deletedCount, err := this.userRepo.DeleteHard(ctx, it.DeleteParam{Id: cmd.Id})
	// ft.PanicOnErr(err)
	// if deletedCount == 0 {
	// 	vErrs.AppendNotFound("id", "user id")
	// 	return &it.DeleteUserResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount), nil
}

func (this *UserServiceImpl) Exists(ctx crud.Context, cmd it.UserExistsCommand) (result *it.UserExistsResult, err error) {
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

func (this *UserServiceImpl) ExistsMulti(ctx crud.Context, cmd it.UserExistsMultiCommand) (result *it.UserExistsMultiResult, err error) {
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

func (this *UserServiceImpl) GetUserById(ctx crud.Context, query it.GetUserByIdQuery) (*it.GetUserByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.User, it.GetUserByIdQuery, it.GetUserByIdResult]{
		Action:      "get user by Id",
		Query:       query,
		RepoFindOne: this.getUserByIdFull,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetUserByIdResult {
			return &it.GetUserByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.User) *it.GetUserByIdResult {
			return &it.GetUserByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err

	// defer func() {
	// 	if e := ft.RecoverPanicFailedTo(recover(), "get user by Id"); e != nil {
	// 		err = e
	// 	}
	// }()

	// var dbUser *domain.User
	// flow := val.StartValidationFlow()
	// vErrs, err := flow.
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		*vErrs = query.Validate()
	// 		return nil
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		dbUser, err = this.assertUserIdExists(ctx, query.Id, query.Status, vErrs)
	// 		return err
	// 	}).
	// 	End()
	// ft.PanicOnErr(err)

	// if vErrs.Count() > 0 {
	// 	return &it.GetUserByIdResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// return &it.GetUserByIdResult{
	// 	Data:    dbUser,
	// 	HasData: dbUser != nil,
	// }, nil
}

func (this *UserServiceImpl) getUserByIdFull(ctx crud.Context, query it.GetUserByIdQuery, vErrs *ft.ValidationErrors) (dbUser *domain.User, err error) {
	dbUser, err = this.userRepo.FindById(ctx, query)
	if dbUser == nil {
		vErrs.AppendNotFound("id", "user id")
	}
	return
}

func (this *UserServiceImpl) getUserByEmailFull(ctx crud.Context, query it.GetUserByEmailQuery, vErrs *ft.ValidationErrors) (dbUser *domain.User, err error) {
	dbUser, err = this.userRepo.FindByEmail(ctx, it.FindByEmailParam{Email: query.Email, Status: query.Status})
	if dbUser == nil {
		vErrs.AppendNotFound("email", "email")
	}
	return
}

func (this *UserServiceImpl) MustGetActiveUser(ctx crud.Context, query it.MustGetActiveUserQuery) (result *it.MustGetActiveUserResult, err error) {
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
			user := &domain.User{}
			user.Id = query.Id
			dbUser, err = this.assertUserIdExists(ctx, user, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if query.Email == nil {
				return nil
			}
			field = "email"
			*query.Email = strings.ToLower(*query.Email)
			dbUser, err = this.assertUserEmailExists(ctx, *query.Email, vErrs)
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

func (this *UserServiceImpl) GetUserByEmail(ctx crud.Context, query it.GetUserByEmailQuery) (*it.GetUserByEmailResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.User, it.GetUserByEmailQuery, it.GetUserByEmailResult]{
		Action: "get user by email",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, query it.GetUserByEmailQuery, vErrs *ft.ValidationErrors) (dbUser *domain.User, err error) {
			return this.getUserByEmailFull(ctx, query, vErrs)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetUserByIdResult {
			return &it.GetUserByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.User) *it.GetUserByIdResult {
			return &it.GetUserByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
	// defer func() {
	// 	if e := ft.RecoverPanicFailedTo(recover(), "get user by email"); e != nil {
	// 		err = e
	// 	}
	// }()

	// var dbUser *domain.User
	// flow := val.StartValidationFlow()
	// vErrs, err := flow.
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		*vErrs = query.Validate()
	// 		return nil
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		query.Email = strings.ToLower(query.Email)
	// 		dbUser, err = this.userRepo.FindByEmail(ctx, it.FindByEmailParam{Email: query.Email, Status: query.Status})
	// 		if err != nil {
	// 			return err
	// 		}
	// 		if dbUser == nil {
	// 			vErrs.AppendNotFound("email", "user email")
	// 		}
	// 		return nil
	// 	}).
	// 	End()
	// ft.PanicOnErr(err)

	// if vErrs.Count() > 0 {
	// 	return &it.GetUserByIdResult{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// return &it.GetUserByIdResult{
	// 	Data:    dbUser,
	// 	HasData: dbUser != nil,
	// }, nil
}

func (this *UserServiceImpl) SearchUsers(ctx crud.Context, query it.SearchUsersQuery) (*it.SearchUsersResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.User, it.SearchUsersQuery, it.SearchUsersResult]{
		Action: "search users",
		Query:  query,
		SetQueryDefaults: func(query *it.SearchUsersQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.userRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query it.SearchUsersQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.User], error) {
			return this.userRepo.Search(ctx, it.SearchParam{
				Predicate:  predicate,
				Order:      order,
				Page:       *query.Page,
				Size:       *query.Size,
				WithGroups: query.WithGroups,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchUsersResult {
			return &it.SearchUsersResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.User]) *it.SearchUsersResult {
			return &it.SearchUsersResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
	// defer func() {
	// 	if e := ft.RecoverPanicFailedTo(recover(), "list users"); e != nil {
	// 		err = e
	// 	}
	// }()

	// query.SetDefaults()
	// vErrsModel := query.Validate()
	// predicate, order, vErrsGraph := this.userRepo.ParseSearchGraph(query.Graph)

	// vErrsModel.Merge(vErrsGraph)

	// if vErrsModel.Count() > 0 {
	// 	return &it.SearchUsersResult{
	// 		ClientError: vErrsModel.ToClientError(),
	// 	}, nil
	// }

	// users, err := this.userRepo.Search(ctx, it.SearchParam{
	// 	Predicate:  predicate,
	// 	Order:      order,
	// 	Page:       *query.Page,
	// 	Size:       *query.Size,
	// 	WithGroups: query.WithGroups,
	// })
	// ft.PanicOnErr(err)

	// return &it.SearchUsersResult{
	// 	Data:    users,
	// 	HasData: users.Items != nil,
	// }, nil
}

func (this *UserServiceImpl) FindDirectApprover(ctx crud.Context, query it.FindDirectApproverQuery) (result *it.FindDirectApproverResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "find direct approver"); e != nil {
			err = e
		}
	}()

	var dbUser *domain.User
	var approver []domain.User

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
			approver, err = this.findDirectApproverInHierarchy(ctx, dbUser, vErrs)
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

func (this *UserServiceImpl) assertUserExists(ctx crud.Context, id model.Id, vErrs *ft.ValidationErrors) (user *domain.User, err error) {
	user, err = this.userRepo.FindById(ctx, it.FindByIdParam{Id: id})
	ft.PanicOnErr(err)

	if user == nil {
		vErrs.AppendNotFound("user_id", "user")
	}
	return user, err
}

func (this *UserServiceImpl) findDirectApproverInHierarchy(ctx crud.Context, dbUser *domain.User, vErrs *ft.ValidationErrors) ([]domain.User, error) {
	if dbUser.HierarchyId == nil {
		return nil, nil
	}

	hierarchy, err := this.hierarchyRepo.FindById(ctx, itHierarchy.FindByIdParam{Id: *dbUser.HierarchyId})
	ft.PanicOnErr(err)

	if hierarchy == nil {
		vErrs.AppendNotFound("hierarchy_id", "hierarchy")
		return nil, nil
	}

	if hierarchy.ParentId == nil {
		return nil, nil
	}

	directManager, err := this.userRepo.FindByHierarchyId(ctx, it.FindByHierarchyIdParam{HierarchyId: *hierarchy.ParentId, Status: domain.WrapUserStatus(domain.UserStatusActive.String())})
	ft.PanicOnErr(err)

	return directManager, nil
}
