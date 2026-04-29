package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/rolerequest"
)

func NewRoleRequestApplicationServiceImpl(
	roleRequestSvc it.RoleRequestDomainService,
	roleRequestRepo it.RoleRequestRepository,
	userPrefSvc itExt.UserPreferenceUiDomainService,
) it.RoleRequestAppService {
	return &RoleRequestApplicationServiceImpl{
		roleRequestSvc:  roleRequestSvc,
		roleRequestRepo: roleRequestRepo,
		userPrefSvc:     userPrefSvc,
	}
}

type RoleRequestApplicationServiceImpl struct {
	roleRequestSvc  it.RoleRequestDomainService
	roleRequestRepo it.RoleRequestRepository
	userPrefSvc     itExt.UserPreferenceUiDomainService
}

func (this *RoleRequestApplicationServiceImpl) CreateRoleRequest(ctx corectx.Context, cmd it.CreateRoleRequestCommand) (*it.CreateRoleRequestResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceAuthorizationGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.CreateRoleRequestResult{ClientErrors: *cErr}, nil
	}
	return this.roleRequestSvc.CreateRoleRequest(ctx, cmd)
}

func (this *RoleRequestApplicationServiceImpl) DeleteRoleRequest(ctx corectx.Context, cmd it.DeleteRoleRequestCommand) (*it.DeleteRoleRequestResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceAuthorizationGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.DeleteRoleRequestResult{ClientErrors: *cErr}, nil
	}
	return this.roleRequestSvc.DeleteRoleRequest(ctx, cmd)
}

func (this *RoleRequestApplicationServiceImpl) GetRoleRequest(ctx corectx.Context, query it.GetRoleRequestQuery) (*it.GetRoleRequestResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.GetRoleRequestResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.RoleRequest, *domain.RoleRequest]{
		Action: "get role request",
		Schema: this.roleRequestRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.RoleRequest], error) {
			return this.roleRequestSvc.GetRoleRequest(ctx, query)
		},
	})
}

func (this *RoleRequestApplicationServiceImpl) RoleRequestExists(ctx corectx.Context, query it.RoleRequestExistsQuery) (*it.RoleRequestExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.RoleRequestExistsResult{ClientErrors: *cErr}, nil
	}
	return this.roleRequestSvc.RoleRequestExists(ctx, query)
}

func (this *RoleRequestApplicationServiceImpl) SearchRoleRequests(ctx corectx.Context, query it.SearchRoleRequestsQuery) (*it.SearchRoleRequestsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.SearchRoleRequestsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.RoleRequest, *domain.RoleRequest]{
		Action:            "search role requests",
		FieldResolver:     this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:            this.roleRequestRepo.GetBaseRepo().Schema(),
		DefaultSearchName: "role_request_list",
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.RoleRequest]], error) {
			return this.roleRequestSvc.SearchRoleRequests(ctx, query)
		},
	})
}

func (this *RoleRequestApplicationServiceImpl) UpdateRoleRequest(ctx corectx.Context, cmd it.UpdateRoleRequestCommand) (*it.UpdateRoleRequestResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceAuthorizationGrantRequest, c.ResourceScopeDomain); cErr != nil {
		return &it.UpdateRoleRequestResult{ClientErrors: *cErr}, nil
	}
	return this.roleRequestSvc.UpdateRoleRequest(ctx, cmd)
}
