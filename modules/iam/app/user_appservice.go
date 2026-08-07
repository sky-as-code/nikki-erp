package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

func NewUserApplicationServiceImpl(
	userDomSvc itUser.UserDomainService,
	userRepo itUser.UserRepository,
	userPrefSvc itExt.UserPreferenceUiDomainService,
) itUser.UserAppService {
	return &UserApplicationServiceImpl{
		userDomSvc:  userDomSvc,
		userRepo:    userRepo,
		userPrefSvc: userPrefSvc,
	}
}

type UserApplicationServiceImpl struct {
	userDomSvc  itUser.UserDomainService
	userRepo    itUser.UserRepository
	userPrefSvc itExt.UserPreferenceUiDomainService
}

func (this *UserApplicationServiceImpl) CreateUser(ctx corectx.Context, cmd itUser.CreateUserCommand) (*itUser.CreateUserResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceIamUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.CreateUserResult{
			ClientErrors: *cErr,
		}, nil
	}
	return this.userDomSvc.CreateUser(ctx, cmd)
}

func (this *UserApplicationServiceImpl) DeleteUser(ctx corectx.Context, cmd itUser.DeleteUserCommand) (*itUser.DeleteUserResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceIamUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.DeleteUserResult{ClientErrors: *cErr}, nil
	}
	return this.userDomSvc.DeleteUser(ctx, cmd)
}

func (this *UserApplicationServiceImpl) GetUser(ctx corectx.Context, query itUser.GetUserQuery) (*itUser.GetUserResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.GetUserResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.User, *domain.User]{
		Action: "get user",
		Schema: this.userRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.User], error) {
			return this.userDomSvc.GetUser(ctx, query)
		},
	})
}

func (this *UserApplicationServiceImpl) GetEnabledUser(ctx corectx.Context, query itUser.GetUserQuery) (*itUser.GetUserResult, error) {
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.User, *domain.User]{
		Action: "get enabled user",
		Schema: this.userRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.User], error) {
			return this.userDomSvc.GetEnabledUser(ctx, query)
		},
	})
}

func (this *UserApplicationServiceImpl) ManageUserRoleAssignments(
	ctx corectx.Context, cmd itUser.ManageUserRoleAssignmentsCommand,
) (*itUser.ManageUserRoleAssignmentsResult, error) {
	if cErr := assertPermission(ctx, "manage_role_assignments", c.ResourceIamUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.ManageUserRoleAssignmentsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.ExecInTranx(ctx, this.userRepo, func(tranxCtx corectx.Context) (*itUser.ManageUserRoleAssignmentsResult, error) {
		return this.userDomSvc.ManageUserRoleAssignments(tranxCtx, cmd)
	})
}

func (this *UserApplicationServiceImpl) SearchUsers(
	ctx corectx.Context, query itUser.SearchUsersQuery,
) (*itUser.SearchUsersResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.SearchUsersResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.User, *domain.User]{
		Action:        "search users",
		FieldResolver: this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:        this.userRepo.GetBaseRepo().Schema(),
		DefaultFields: []string{models.UserFieldAvatarUrl, models.UserFieldDisplayName, models.UserFieldEmail, models.UserFieldStatus},
		// MaskedFields:  []string{models.UserFieldPassword},
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.User]], error) {
			return this.userDomSvc.SearchUsers(ctx, query, corecrud.ServiceSearchOptions{
				AfterValidationSuccess: fn,
			})
		},
	})
}

func (this *UserApplicationServiceImpl) SetUserIsArchived(ctx corectx.Context, cmd itUser.SetUserIsArchivedCommand) (*itUser.SetUserIsArchivedResult, error) {
	if cErr := assertPermission(ctx, "set_archived", c.ResourceIamUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.SetUserIsArchivedResult{ClientErrors: *cErr}, nil
	}
	return this.userDomSvc.SetUserIsArchived(ctx, cmd)
}

func (this *UserApplicationServiceImpl) UserExists(ctx corectx.Context, query itUser.UserExistsQuery) (*itUser.UserExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIamUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.UserExistsResult{ClientErrors: *cErr}, nil
	}
	return this.userDomSvc.UserExists(ctx, query)
}

func (this *UserApplicationServiceImpl) UpdateUser(ctx corectx.Context, cmd itUser.UpdateUserCommand) (*itUser.UpdateUserResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceIamUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.UpdateUserResult{ClientErrors: *cErr}, nil
	}
	return this.userDomSvc.UpdateUser(ctx, cmd)
}
