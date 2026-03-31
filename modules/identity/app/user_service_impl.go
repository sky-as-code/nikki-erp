package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewUserServiceImpl(
	enumSvc enum.EnumService,
	userRepo it.UserRepository,
	userRepo2 it.UserRepository,
	cqrsBus cqrs.CqrsBus,
	eventBus event.EventBus,
) it.UserService {
	return &UserServiceImpl{
		enumSvc:   enumSvc,
		userRepo:  userRepo,
		userRepo2: userRepo2,
		cqrs:      cqrsBus,
		eventBus:  eventBus,
	}
}

type UserServiceImpl struct {
	enumSvc   enum.EnumService
	userRepo  it.UserRepository
	userRepo2 it.UserRepository
	eventBus  event.EventBus
	cqrs      cqrs.CqrsBus
}

func (this *UserServiceImpl) GetUserContext(ctx crud.Context, query it.GetUserContextQuery) (result any, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "add or remove users"); e != nil {
			err = e
		}
	}()

	return nil, nil

	// var dbUser *domain.User

	// flow := val.StartValidationFlow()
	// vErrs, err := flow.
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		*vErrs = query.Validate()
	// 		return nil
	// 	}).
	// 	Step(func(vErrs *ft.ValidationErrors) error {
	// 		// dbUser, err = this.userRepo.GetOne(ctx, dyn.RepoGetOneParam{
	// 		// 	Filter: dmodel.DynamicFields{
	// 		// 		basemodel.FieldId: query.UserId,
	// 		// 	},
	// 		// })
	// 		// ft.PanicOnErr(err)
	// 		return nil
	// 	}).
	// 	End()
	// ft.PanicOnErr(err)

	// // permission, err := this.getPermissionsForUser(ctx, &vErrs, query.UserId)
	// // ft.PanicOnErr(err)

	// if vErrs.Count() > 0 {
	// 	return &it.GetUserContextResultData{
	// 		ClientError: vErrs.ToClientError(),
	// 	}, nil
	// }

	// return &it.GetUserContextResultData{
	// 	Data: &it.GetUserContextResult{
	// 		User: dbUser,
	// 		// Permissions: &permission.Permissions,
	// 		Permissions: nil,
	// 	},
	// 	HasData: true,
	// }, nil
}

func (this *UserServiceImpl) CreateUser(ctx corectx.Context, cmd it.CreateUserCommand) (*it.CreateUserResult, error) {
	return corecrud.Create(ctx, dyn.CreateParam[domain.User, *domain.User]{
		Action:         "create user",
		BaseRepoGetter: this.userRepo2,
		Data:           cmd,
		BeforeValidation: func(ctx corectx.Context, model *domain.User) (*domain.User, error) {
			// Normal users must not have this field set.
			model.SetIsOwner(nil)
			return model, nil
		},
	})
}

func (this *UserServiceImpl) DeleteUser(ctx corectx.Context, cmd it.DeleteUserCommand) (*it.DeleteUserResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete user",
		DbRepoGetter: this.userRepo2,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *UserServiceImpl) GetUser(ctx corectx.Context, query it.GetUserQuery) (*it.GetUserResult, error) {
	return this.getUserWithArchived(ctx, query, nil)
}

func (this *UserServiceImpl) GetActiveUser(ctx corectx.Context, query it.GetUserQuery) (*it.GetUserResult, error) {
	return this.getUserWithArchived(ctx, query, util.ToPtr(true))
}

func (this *UserServiceImpl) getUserWithArchived(ctx corectx.Context, query it.GetUserQuery, isArchived *bool) (*it.GetUserResult, error) {
	querySchema := getOneSchema()
	sanitizedFields, cErrs := querySchema.ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &it.GetUserResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitizedFields.(*it.GetUserQuery))

	statusNode := dmodel.NewSearchNode()
	if isArchived != nil {
		statusNode.NewCondition(basemodel.FieldIsArchived, dmodel.Equals, *isArchived)
	}

	keyNode := dmodel.NewSearchNode()
	if query.Id != nil {
		keyNode.NewCondition(domain.UserFieldId, dmodel.Equals, *query.Id)
	}
	if query.Email != nil {
		keyNode.NewCondition(domain.UserFieldEmail, dmodel.Equals, *query.Email)
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*statusNode, *keyNode)

	searchquery := it.SearchUsersQuery{
		Columns: query.Columns,
		Graph:   graph,
		Page:    0,
		Size:    1,
	}

	searchRes, err := this.SearchUsers(ctx, searchquery)
	if err != nil {
		return nil, err
	}
	result := &it.GetUserResult{
		ClientErrors: searchRes.ClientErrors,
		HasData:      searchRes.HasData,
	}

	if searchRes.HasData {
		result.Data = searchRes.Data.Items[0]
	}

	return result, nil
}

func getOneSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"identity.get_user_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				ExclusiveFields(domain.UserFieldId, domain.UserFieldEmail).
				Field(dmodel.DefineField().
					Name(basemodel.FieldColumns).
					DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType())).
				Field(dmodel.DefineField().
					Name(basemodel.FieldId).
					DataType(dmodel.FieldDataTypeUlid())).
				Field(dmodel.DefineField().
					Name(domain.UserFieldEmail).
					DataType(dmodel.FieldDataTypeEmail()))
		},
	)
}

func (this *UserServiceImpl) SearchUsers(
	ctx corectx.Context, query it.SearchUsersQuery,
) (*it.SearchUsersResult, error) {
	return corecrud.Search[domain.User](ctx, corecrud.SearchParam{
		Action:       "search users",
		DbRepoGetter: this.userRepo2,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *UserServiceImpl) SetUserIsArchived(ctx corectx.Context, cmd it.SetUserIsArchivedCommand) (*it.SetUserIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.userRepo2, dyn.SetIsArchivedCommand(cmd))
}

func (this *UserServiceImpl) UserExists(ctx corectx.Context, query it.UserExistsQuery) (*it.UserExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if users exist",
		DbRepoGetter: this.userRepo2,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *UserServiceImpl) UpdateUser(ctx corectx.Context, cmd it.UpdateUserCommand) (*it.UpdateUserResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.User, *domain.User]{
		Action:       "update user",
		DbRepoGetter: this.userRepo2,
		Data:         cmd,
	})
}

// func (this *UserServiceImpl) getUserByIdFull(ctx crud.Context, query it.GetUserQuery, vErrs *ft.ValidationErrors) (dbUser *domain.User, err error) {
// 	dbUser, err = this.userRepo.FindById(ctx, query)
// 	if dbUser == nil {
// 		vErrs.AppendNotFound("id", "user id")
// 	}
// 	return
// }

// func (this *UserServiceImpl) getPermissionsForUser(ctx crud.Context, vErrs *ft.ValidationErrors, userId model.Id) (permissions *itAuthorize.PermissionSnapshotResult, err error) {
// 	result := itAuthorize.PermissionSnapshotResult{}
// 	err = this.cqrs.Request(ctx, &itAuthorize.PermissionSnapshotQuery{UserId: userId}, &result)
// 	fault.PanicOnErr(err)

// 	if result.ClientError != nil {
// 		if !vErrs.MergeClientError(result.ClientError) {
// 			vErrs.AppendNotFound("permissions", "permissions")
// 		}
// 		return nil, result.ClientError
// 	}

// 	permissions = &result
// 	return permissions, err
// }
