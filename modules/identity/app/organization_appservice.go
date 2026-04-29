package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewOrganizationApplicationServiceImpl(
	orgSvc it.OrganizationDomainService,
	orgRepo it.OrganizationRepository,
	userPrefSvc itExt.UserPreferenceUiDomainService,
) it.OrganizationAppService {
	return &OrganizationApplicationServiceImpl{
		orgSvc:      orgSvc,
		orgRepo:     orgRepo,
		userPrefSvc: userPrefSvc,
	}
}

type OrganizationApplicationServiceImpl struct {
	orgSvc      it.OrganizationDomainService
	orgRepo     it.OrganizationRepository
	userPrefSvc itExt.UserPreferenceUiDomainService
}

func (this *OrganizationApplicationServiceImpl) CreateOrg(ctx corectx.Context, cmd it.CreateOrgCommand) (*it.CreateOrgResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceIdentityOrganization, c.ResourceScopeDomain); cErr != nil {
		return &it.CreateOrgResult{ClientErrors: *cErr}, nil
	}
	return this.orgSvc.CreateOrg(ctx, cmd)
}

func (this *OrganizationApplicationServiceImpl) DeleteOrg(ctx corectx.Context, cmd it.DeleteOrgCommand) (*it.DeleteOrgResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceIdentityOrganization, c.ResourceScopeDomain); cErr != nil {
		return &it.DeleteOrgResult{ClientErrors: *cErr}, nil
	}
	return this.orgSvc.DeleteOrg(ctx, cmd)
}

func (this *OrganizationApplicationServiceImpl) GetOrg(ctx corectx.Context, query it.GetOrgQuery) (*it.GetOrgResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityOrganization, c.ResourceScopeDomain); cErr != nil {
		return &it.GetOrgResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.Organization, *domain.Organization]{
		Action: "get organization",
		Schema: this.orgRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.Organization], error) {
			return this.orgSvc.GetOrg(ctx, query)
		},
	})
}

func (this *OrganizationApplicationServiceImpl) OrgExists(ctx corectx.Context, query it.OrgExistsQuery) (*it.OrgExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityOrganization, c.ResourceScopeDomain); cErr != nil {
		return &it.OrgExistsResult{ClientErrors: *cErr}, nil
	}
	return this.orgSvc.OrgExists(ctx, query)
}

func (this *OrganizationApplicationServiceImpl) ManageOrgUsers(ctx corectx.Context, cmd it.ManageOrgUsersCommand) (*it.ManageOrgUsersResult, error) {
	if cErr := assertPermission(ctx, "manage_users", c.ResourceIdentityOrganization, c.ResourceScopeDomain); cErr != nil {
		return &it.ManageOrgUsersResult{ClientErrors: *cErr}, nil
	}
	return this.orgSvc.ManageOrgUsers(ctx, cmd)
}

func (this *OrganizationApplicationServiceImpl) SearchOrgs(ctx corectx.Context, query it.SearchOrgsQuery) (*it.SearchOrgsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityOrganization, c.ResourceScopeDomain); cErr != nil {
		return &it.SearchOrgsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.Organization, *domain.Organization]{
		Action:            "search organizations",
		FieldResolver:     this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:            this.orgRepo.GetBaseRepo().Schema(),
		DefaultSearchName: "organization_list",
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.Organization]], error) {
			return this.orgSvc.SearchOrgs(ctx, query)
		},
	})
}

func (this *OrganizationApplicationServiceImpl) SetOrgIsArchived(ctx corectx.Context, cmd it.SetOrgIsArchivedCommand) (*it.SetOrgIsArchivedResult, error) {
	if cErr := assertPermission(ctx, "set_archived", c.ResourceIdentityOrganization, c.ResourceScopeDomain); cErr != nil {
		return &it.SetOrgIsArchivedResult{ClientErrors: *cErr}, nil
	}
	return this.orgSvc.SetOrgIsArchived(ctx, cmd)
}

func (this *OrganizationApplicationServiceImpl) UpdateOrg(ctx corectx.Context, cmd it.UpdateOrgCommand) (*it.UpdateOrgResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceIdentityOrganization, c.ResourceScopeDomain); cErr != nil {
		return &it.UpdateOrgResult{ClientErrors: *cErr}, nil
	}
	return this.orgSvc.UpdateOrg(ctx, cmd)
}
