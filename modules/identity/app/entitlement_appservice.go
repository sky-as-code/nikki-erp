package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/entitlement"
	itExt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/external"
)

func NewEntitlementApplicationServiceImpl(
	entitlementSvc it.EntitlementDomainService,
	entitlementRepo it.EntitlementRepository,
	userPrefSvc itExt.UserPreferenceUiDomainService,
) it.EntitlementAppService {
	return &EntitlementApplicationServiceImpl{
		entitlementSvc:  entitlementSvc,
		entitlementRepo: entitlementRepo,
		userPrefSvc:     userPrefSvc,
	}
}

type EntitlementApplicationServiceImpl struct {
	entitlementSvc  it.EntitlementDomainService
	entitlementRepo it.EntitlementRepository
	userPrefSvc     itExt.UserPreferenceUiDomainService
}

func (this *EntitlementApplicationServiceImpl) CreateEntitlement(ctx corectx.Context, cmd it.CreateEntitlementCommand) (*it.CreateEntitlementResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceAuthorizationEntitlement, c.ResourceScopeDomain); cErr != nil {
		return &it.CreateEntitlementResult{ClientErrors: *cErr}, nil
	}
	return this.entitlementSvc.CreateEntitlement(ctx, cmd)
}

func (this *EntitlementApplicationServiceImpl) DeleteEntitlement(ctx corectx.Context, cmd it.DeleteEntitlementCommand) (*it.DeleteEntitlementResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceAuthorizationEntitlement, c.ResourceScopeDomain); cErr != nil {
		return &it.DeleteEntitlementResult{ClientErrors: *cErr}, nil
	}
	return this.entitlementSvc.DeleteEntitlement(ctx, cmd)
}

func (this *EntitlementApplicationServiceImpl) EntitlementExists(ctx corectx.Context, query it.EntitlementExistsQuery) (*it.EntitlementExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationEntitlement, c.ResourceScopeDomain); cErr != nil {
		return &it.EntitlementExistsResult{ClientErrors: *cErr}, nil
	}
	return this.entitlementSvc.EntitlementExists(ctx, query)
}

func (this *EntitlementApplicationServiceImpl) GetEntitlement(ctx corectx.Context, query it.GetEntitlementQuery) (*it.GetEntitlementResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationEntitlement, c.ResourceScopeDomain); cErr != nil {
		return &it.GetEntitlementResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.Entitlement, *domain.Entitlement]{
		Action: "get entitlement",
		Schema: this.entitlementRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.Entitlement], error) {
			return this.entitlementSvc.GetEntitlement(ctx, query)
		},
	})
}

func (this *EntitlementApplicationServiceImpl) ManageEntitlementRoles(ctx corectx.Context, cmd it.ManageEntitlementRolesCommand) (*it.ManageEntitlementRolesResult, error) {
	if cErr := assertPermission(ctx, "manage_roles", c.ResourceAuthorizationEntitlement, c.ResourceScopeDomain); cErr != nil {
		return &it.ManageEntitlementRolesResult{ClientErrors: *cErr}, nil
	}
	return this.entitlementSvc.ManageEntitlementRoles(ctx, cmd)
}

func (this *EntitlementApplicationServiceImpl) SearchEntitlements(ctx corectx.Context, query it.SearchEntitlementsQuery) (*it.SearchEntitlementsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationEntitlement, c.ResourceScopeDomain); cErr != nil {
		return &it.SearchEntitlementsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.Entitlement, *domain.Entitlement]{
		Action:            "search entitlements",
		FieldResolver:     this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:            this.entitlementRepo.GetBaseRepo().Schema(),
		DefaultSearchName: "entitlement_list",
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.Entitlement]], error) {
			return this.entitlementSvc.SearchEntitlements(ctx, query)
		},
	})
}

func (this *EntitlementApplicationServiceImpl) SetEntitlementIsArchived(ctx corectx.Context, cmd it.SetEntitlementIsArchivedCommand) (*it.SetEntitlementIsArchivedResult, error) {
	if cErr := assertPermission(ctx, "set_archived", c.ResourceAuthorizationEntitlement, c.ResourceScopeDomain); cErr != nil {
		return &it.SetEntitlementIsArchivedResult{ClientErrors: *cErr}, nil
	}
	return this.entitlementSvc.SetEntitlementIsArchived(ctx, cmd)
}

func (this *EntitlementApplicationServiceImpl) UpdateEntitlement(ctx corectx.Context, cmd it.UpdateEntitlementCommand) (*it.UpdateEntitlementResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceAuthorizationEntitlement, c.ResourceScopeDomain); cErr != nil {
		return &it.UpdateEntitlementResult{ClientErrors: *cErr}, nil
	}
	return this.entitlementSvc.UpdateEntitlement(ctx, cmd)
}
