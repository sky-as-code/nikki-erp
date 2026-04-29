package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/orgunit"
)

func NewOrgUnitApplicationServiceImpl(
	orgUnitSvc it.OrgUnitDomainService,
	orgUnitRepo it.OrgUnitRepository,
	userPrefSvc itExt.UserPreferenceUiDomainService,
) it.OrgUnitAppService {
	return &OrgUnitApplicationServiceImpl{
		orgUnitSvc:  orgUnitSvc,
		orgUnitRepo: orgUnitRepo,
		userPrefSvc: userPrefSvc,
	}
}

type OrgUnitApplicationServiceImpl struct {
	orgUnitSvc  it.OrgUnitDomainService
	orgUnitRepo it.OrgUnitRepository
	userPrefSvc itExt.UserPreferenceUiDomainService
}

func (this *OrgUnitApplicationServiceImpl) CreateOrgUnit(ctx corectx.Context, cmd it.CreateOrgUnitCommand) (*it.CreateOrgUnitResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceIdentityOrgUnit, c.ResourceScopeOrg); cErr != nil {
		return &it.CreateOrgUnitResult{ClientErrors: *cErr}, nil
	}
	return this.orgUnitSvc.CreateOrgUnit(ctx, cmd)
}

func (this *OrgUnitApplicationServiceImpl) DeleteOrgUnit(ctx corectx.Context, cmd it.DeleteOrgUnitCommand) (*it.DeleteOrgUnitResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceIdentityOrgUnit, c.ResourceScopeOrg); cErr != nil {
		return &it.DeleteOrgUnitResult{ClientErrors: *cErr}, nil
	}
	return this.orgUnitSvc.DeleteOrgUnit(ctx, cmd)
}

func (this *OrgUnitApplicationServiceImpl) GetOrgUnit(ctx corectx.Context, query it.GetOrgUnitQuery) (*it.GetOrgUnitResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityOrgUnit, c.ResourceScopeOrg); cErr != nil {
		return &it.GetOrgUnitResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.OrganizationalUnit, *domain.OrganizationalUnit]{
		Action: "get org unit",
		Schema: this.orgUnitRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.OrganizationalUnit], error) {
			return this.orgUnitSvc.GetOrgUnit(ctx, query)
		},
	})
}

func (this *OrgUnitApplicationServiceImpl) OrgUnitExists(ctx corectx.Context, cmd it.OrgUnitExistsQuery) (*it.OrgUnitExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityOrgUnit, c.ResourceScopeOrg); cErr != nil {
		return &it.OrgUnitExistsResult{ClientErrors: *cErr}, nil
	}
	return this.orgUnitSvc.OrgUnitExists(ctx, cmd)
}

func (this *OrgUnitApplicationServiceImpl) ManageOrgUnitUsers(ctx corectx.Context, cmd it.ManageOrgUnitUsersCommand) (*it.ManageOrgUnitUsersResult, error) {
	if cErr := assertPermission(ctx, "manage_users", c.ResourceIdentityOrgUnit, c.ResourceScopeOrg); cErr != nil {
		return &it.ManageOrgUnitUsersResult{ClientErrors: *cErr}, nil
	}
	return this.orgUnitSvc.ManageOrgUnitUsers(ctx, cmd)
}

func (this *OrgUnitApplicationServiceImpl) SearchOrgUnits(ctx corectx.Context, query it.SearchOrgUnitsQuery) (*it.SearchOrgUnitsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityOrgUnit, c.ResourceScopeOrg); cErr != nil {
		return &it.SearchOrgUnitsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.OrganizationalUnit, *domain.OrganizationalUnit]{
		Action:            "search org units",
		FieldResolver:     this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:            this.orgUnitRepo.GetBaseRepo().Schema(),
		DefaultSearchName: "orgunit_list",
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.OrganizationalUnit]], error) {
			return this.orgUnitSvc.SearchOrgUnits(ctx, query)
		},
	})
}

func (this *OrgUnitApplicationServiceImpl) UpdateOrgUnit(ctx corectx.Context, cmd it.UpdateOrgUnitCommand) (*it.UpdateOrgUnitResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceIdentityOrgUnit, c.ResourceScopeOrg); cErr != nil {
		return &it.UpdateOrgUnitResult{ClientErrors: *cErr}, nil
	}
	return this.orgUnitSvc.UpdateOrgUnit(ctx, cmd)
}
