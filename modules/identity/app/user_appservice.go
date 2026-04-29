package app

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/external"
	itRole "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/role"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewUserApplicationServiceImpl(
	roleSvc itRole.RoleDomainService,
	userDomSvc itUser.UserDomainService,
	userRepo itUser.UserRepository,
	userPrefSvc itExt.UserPreferenceUiDomainService,
) itUser.UserAppService {
	return &UserApplicationServiceImpl{
		roleSvc:     roleSvc,
		userDomSvc:  userDomSvc,
		userRepo:    userRepo,
		userPrefSvc: userPrefSvc,
	}
}

type UserApplicationServiceImpl struct {
	roleSvc     itRole.RoleDomainService
	userDomSvc  itUser.UserDomainService
	userRepo    itUser.UserRepository
	userPrefSvc itExt.UserPreferenceUiDomainService
}

func (this *UserApplicationServiceImpl) GetUserContext(ctx crud.Context, query itUser.GetUserContextQuery) (result any, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "add or remove users"); e != nil {
			err = e
		}
	}()

	return nil, nil
}

func (this *UserApplicationServiceImpl) CreateUser(ctx corectx.Context, cmd itUser.CreateUserCommand) (*itUser.CreateUserResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceIdentityUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.CreateUserResult{
			ClientErrors: *cErr,
		}, nil
	}
	return corecrud.ExecInTranx(ctx, this.userRepo, func(tranxCtx corectx.Context) (*itUser.CreateUserResult, error) {
		result, err := this.userDomSvc.CreateUser(tranxCtx, cmd)
		if err != nil {
			return nil, err
		}
		if result.ClientErrors.Count() > 0 {
			return result, nil
		}
		return this.createPrivateRole(tranxCtx, result)
	})
}

func (this *UserApplicationServiceImpl) createPrivateRole(tranxCtx corectx.Context, usrResult *itUser.CreateUserResult) (*itUser.CreateUserResult, error) {
	oid := string(*usrResult.Data.GetId())
	roleRes, rErr := this.roleSvc.CreatePrivateRole(tranxCtx, itRole.CreatePrivateRoleCommand{OwnerId: oid, OwnerType: "user"})
	if rErr != nil {
		return nil, rErr
	}
	if roleRes.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(roleRes.ClientErrors.ToError(), "createPrivateRole")
	}
	return usrResult, nil
}

func (this *UserApplicationServiceImpl) DeleteUser(ctx corectx.Context, cmd itUser.DeleteUserCommand) (*itUser.DeleteUserResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceIdentityUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.DeleteUserResult{ClientErrors: *cErr}, nil
	}
	return this.userDomSvc.DeleteUser(ctx, cmd, corecrud.ServiceDeleteOptions{
		AfterValidationSuccess: func(tranxCtx corectx.Context, cmd dyn.DeleteOneCommand) (dyn.DeleteOneCommand, error) {
			privRes, pErr := this.roleSvc.DeletePrivateRole(tranxCtx, itRole.DeletePrivateRoleCommand{OwnerId: cmd.Id})
			if pErr != nil {
				return cmd, pErr
			}
			if privRes.ClientErrors.Count() > 0 {
				return cmd, errors.Wrap(privRes.ClientErrors.ToError(), "deletePrivateRole")
			}
			return cmd, nil
		},
	})
}

func (this *UserApplicationServiceImpl) GetUser(ctx corectx.Context, query itUser.GetUserQuery) (*itUser.GetUserResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityUser, c.ResourceScopeOrg); cErr != nil {
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

func (this *UserApplicationServiceImpl) SearchUsers(
	ctx corectx.Context, query itUser.SearchUsersQuery,
) (*itUser.SearchUsersResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.SearchUsersResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.User, *domain.User]{
		Action:            "search users",
		FieldResolver:     this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:            this.userRepo.GetBaseRepo().Schema(),
		DefaultSearchName: "user_list",
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.User]], error) {
			return this.userDomSvc.SearchUsers(ctx, query)
		},
	})
}

func (this *UserApplicationServiceImpl) SetUserIsArchived(ctx corectx.Context, cmd itUser.SetUserIsArchivedCommand) (*itUser.SetUserIsArchivedResult, error) {
	if cErr := assertPermission(ctx, "set_archived", c.ResourceIdentityUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.SetUserIsArchivedResult{ClientErrors: *cErr}, nil
	}
	return this.userDomSvc.SetUserIsArchived(ctx, cmd)
}

func (this *UserApplicationServiceImpl) UserExists(ctx corectx.Context, query itUser.UserExistsQuery) (*itUser.UserExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.UserExistsResult{ClientErrors: *cErr}, nil
	}
	return this.userDomSvc.UserExists(ctx, query)
}

func (this *UserApplicationServiceImpl) UpdateUser(ctx corectx.Context, cmd itUser.UpdateUserCommand) (*itUser.UpdateUserResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceIdentityUser, c.ResourceScopeOrg); cErr != nil {
		return &itUser.UpdateUserResult{ClientErrors: *cErr}, nil
	}
	return this.userDomSvc.UpdateUser(ctx, cmd)
}
