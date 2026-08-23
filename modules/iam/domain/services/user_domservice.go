package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	enum "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

func NewUserDomainServiceImpl(
	enumSvc enum.EnumService,
	userRepo it.UserRepository,
	permRepo itPerm.PermissionRepository,
	historyRepo itPerm.PermissionHistoryRepository,
	cqrsBus cqrs.CqrsBus,
	eventBus event.EventBus,
	settingsSvc itExt.UserSettingsExtService,
) it.UserDomainService {
	return &UserDomainServiceImpl{
		enumSvc:     enumSvc,
		userRepo:    userRepo,
		permRepo:    permRepo,
		auditor:     permissionAuditor{historyRepo: historyRepo},
		cqrs:        cqrsBus,
		eventBus:    eventBus,
		settingsSvc: settingsSvc,
	}
}

type UserDomainServiceImpl struct {
	enumSvc     enum.EnumService
	userRepo    it.UserRepository
	permRepo    itPerm.PermissionRepository
	auditor     permissionAuditor
	eventBus    event.EventBus
	cqrs        cqrs.CqrsBus
	settingsSvc itExt.UserSettingsExtService
}

// ManageUserRoleAssignments assigns/removes roles to/from a user, then refreshes that user's
// denormalized permissions so authorization reflects the change immediately.
func (this *UserDomainServiceImpl) ManageUserRoleAssignments(
	ctx corectx.Context, cmd it.ManageUserRoleAssignmentsCommand,
) (*it.ManageUserRoleAssignmentsResult, error) {
	result, err := corecrud.ManageM2m(ctx, corecrud.ManageM2mParam{
		Action:             "manage user role assignments",
		DbRepoGetter:       this.userRepo,
		DestSchemaName:     domain.RoleSchemaName,
		SrcId:              cmd.UserId,
		SrcIdFieldForError: "user_id",
		AssociatedIds:      cmd.Add,
		DisassociatedIds:   cmd.Remove,
		BeforeInsert: func(ctx corectx.Context, dbRecords []dmodel.DynamicFields) error {
			for _, record := range dbRecords {
				assignmentId, err := model.NewId()
				if err != nil {
					return err
				}
				record[domain.RoleUserAssignFieldId] = *assignmentId
			}
			return nil
		},
	})
	if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
		return result, err
	}
	if err := this.permRepo.RebuildUserPermission(ctx, cmd.UserId); err != nil {
		return nil, err
	}
	// Same transaction as the assignment: the trail and the grant stand or fall together.
	if err := this.auditRoleAssignments(ctx, cmd); err != nil {
		return nil, err
	}
	return result, nil
}

func (this *UserDomainServiceImpl) auditRoleAssignments(
	ctx corectx.Context, cmd it.ManageUserRoleAssignmentsCommand,
) error {
	receivers := []model.Id{cmd.UserId}
	for _, roleId := range cmd.Add.ToSlice() {
		if err := this.auditor.recordRoleTransition(
			ctx, domain.PermissionHistoryEffectGrant, domain.PermissionHistoryReasonRoleAdded,
			roleId, receivers,
		); err != nil {
			return err
		}
	}
	for _, roleId := range cmd.Remove.ToSlice() {
		if err := this.auditor.recordRoleTransition(
			ctx, domain.PermissionHistoryEffectRevoke, domain.PermissionHistoryReasonRoleRemoved,
			roleId, receivers,
		); err != nil {
			return err
		}
	}
	return nil
}

// CreateUser creates the user and seeds their preferences in one transaction.
//
// The seeding shares the transaction so that a user never exists without the preference rows the
// application expects them to have: their first sign-in reads a theme and a language, and there is
// no later point that would repair a user created without them.
//
// It is done around the create rather than in a hook: corecrud has no after-insert hook, both
// BeforeValidation and AfterValidationSuccess run before the row is written, and BeforeValidation
// is already taken here.
func (this *UserDomainServiceImpl) CreateUser(
	ctx corectx.Context, cmd it.CreateUserCommand, options ...corecrud.ServiceCreateOptions[*domain.User],
) (*it.CreateUserResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceCreateOptions[*domain.User]{})

	tranx, err := this.userRepo.GetBaseRepo().BeginTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "create user")
	}
	defer tranx.Rollback()

	// The transaction goes on a clone: setting it on the caller's context would leave a committed
	// transaction visible to whatever runs next.
	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	result, err := corecrud.Create(tranxCtx, corecrud.CreateParam[domain.User, *domain.User]{
		Action:         "create user",
		BaseRepoGetter: this.userRepo,
		Data:           cmd,
		BeforeValidation: func(_ corectx.Context, model *domain.User, cErrs *ft.ClientErrors) (*domain.User, error) {
			// Normal users must not have this field set.
			model.SetIsOwner(nil)
			return model, nil
		},
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
	if err != nil {
		return nil, err
	}
	// A rejected create has written nothing, so there is nothing to seed and nothing to commit.
	if result.ClientErrors.Count() > 0 || !result.HasData {
		return result, nil
	}

	if err := this.initUserPreferences(tranxCtx, &result.Data); err != nil {
		return nil, err
	}
	return result, errors.Wrap(tranx.Commit(), "create user")
}

// initUserPreferences copies the tenant's settings onto the newly created user.
func (this *UserDomainServiceImpl) initUserPreferences(
	ctx corectx.Context, user *domain.User,
) error {
	userId := user.GetId()
	if userId == nil {
		return errors.New("create user: the created user has no id")
	}

	initResult, err := this.settingsSvc.InitUserPreferences(ctx, itExt.InitOwnerSettingsCommand{
		OwnerId: *userId,
	})
	if err != nil {
		return errors.Wrap(err, "create user")
	}
	// A rejected seeding is a defect in this call rather than something the caller creating a user
	// can correct, so it fails the whole create.
	if initResult.ClientErrors.Count() > 0 {
		return errors.Wrap(initResult.ClientErrors.ToError(), "create user")
	}
	return nil
}

func (this *UserDomainServiceImpl) DeleteUser(
	ctx corectx.Context, cmd it.DeleteUserCommand, options ...corecrud.ServiceDeleteOptions,
) (*it.DeleteUserResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceDeleteOptions{})
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:                 "delete user",
		DbRepoGetter:           this.userRepo,
		Cmd:                    dyn.DeleteOneCommand(cmd),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *UserDomainServiceImpl) GetUser(ctx corectx.Context, query it.GetUserQuery) (*dyn.OpResult[domain.User], error) {
	return this.getUserWithArchived(ctx, query, nil)
}

func (this *UserDomainServiceImpl) GetEnabledUser(ctx corectx.Context, query it.GetUserQuery) (*dyn.OpResult[domain.User], error) {
	return this.getUserWithArchived(ctx, query, util.ToPtr(false))
}

func (this *UserDomainServiceImpl) getUserWithArchived(ctx corectx.Context, query it.GetUserQuery, isArchived *bool) (*dyn.OpResult[domain.User], error) {
	sanitizedFields, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[domain.User]{ClientErrors: cErrs}, nil
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
		Fields: query.Fields,
		Graph:  graph,
		Page:   0,
		Size:   1,
	}

	searchRes, err := this.SearchUsers(ctx, searchquery)
	if err != nil {
		return nil, err
	}
	result := &dyn.OpResult[domain.User]{
		ClientErrors: searchRes.ClientErrors,
		HasData:      searchRes.HasData,
	}

	if searchRes.HasData {
		result.Data = searchRes.Data.Items[0]
	}

	return result, nil
}

func (this *UserDomainServiceImpl) SearchUsers(
	ctx corectx.Context, query it.SearchUsersQuery, options ...corecrud.ServiceSearchOptions,
) (*it.SearchUsersResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	return corecrud.Search[domain.User](ctx, corecrud.SearchParam{
		Action:                 "search users",
		DbRepoGetter:           this.userRepo,
		Query:                  dyn.SearchQuery(query),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *UserDomainServiceImpl) SetUserIsArchived(ctx corectx.Context, cmd it.SetUserIsArchivedCommand) (*it.SetUserIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.userRepo, dyn.SetIsArchivedCommand(cmd))
}

func (this *UserDomainServiceImpl) UserExists(ctx corectx.Context, query it.UserExistsQuery) (*it.UserExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if users exist",
		DbRepoGetter: this.userRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *UserDomainServiceImpl) UpdateUser(
	ctx corectx.Context, cmd it.UpdateUserCommand, options ...corecrud.ServiceUpdateOptions[*domain.User],
) (*it.UpdateUserResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceUpdateOptions[*domain.User]{})
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.User, *domain.User]{
		Action:                 "update user",
		DbRepoGetter:           this.userRepo,
		Data:                   cmd,
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

// func (this *UserDomainServiceImpl) getUserByIdFull(ctx crud.Context, query it.GetUserQuery, vErrs *ft.ValidationErrors) (dbUser *domain.User, err error) {
// 	dbUser, err = this.userRepo.FindById(ctx, query)
// 	if dbUser == nil {
// 		vErrs.AppendNotFound("id", "user id")
// 	}
// 	return
// }

// func (this *UserDomainServiceImpl) getPermissionsForUser(ctx crud.Context, vErrs *ft.ValidationErrors, userId model.Id) (permissions *itAuthorize.PermissionSnapshotResult, err error) {
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
