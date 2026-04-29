package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

func NewGroupApplicationServiceImpl(
	groupSvc it.GroupDomainService,
	groupRepo it.GroupRepository,
	userPrefSvc itExt.UserPreferenceUiDomainService,
) it.GroupAppService {
	return &GroupApplicationServiceImpl{
		groupSvc:    groupSvc,
		groupRepo:   groupRepo,
		userPrefSvc: userPrefSvc,
	}
}

type GroupApplicationServiceImpl struct {
	groupSvc    it.GroupDomainService
	groupRepo   it.GroupRepository
	userPrefSvc itExt.UserPreferenceUiDomainService
}

func (this *GroupApplicationServiceImpl) CreateGroup(ctx corectx.Context, cmd it.CreateGroupCommand) (*it.CreateGroupResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceIdentityGroup, c.ResourceScopeOrg); cErr != nil {
		return &it.CreateGroupResult{ClientErrors: *cErr}, nil
	}
	return this.groupSvc.CreateGroup(ctx, cmd)
}

func (this *GroupApplicationServiceImpl) DeleteGroup(ctx corectx.Context, cmd it.DeleteGroupCommand) (*it.DeleteGroupResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceIdentityGroup, c.ResourceScopeOrg); cErr != nil {
		return &it.DeleteGroupResult{ClientErrors: *cErr}, nil
	}
	return this.groupSvc.DeleteGroup(ctx, cmd)
}

func (this *GroupApplicationServiceImpl) GetGroup(ctx corectx.Context, query it.GetGroupQuery) (*it.GetGroupResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityGroup, c.ResourceScopeOrg); cErr != nil {
		return &it.GetGroupResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.Group, *domain.Group]{
		Action: "get group",
		Schema: this.groupRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.Group], error) {
			return this.groupSvc.GetGroup(ctx, query)
		},
	})
}

func (this *GroupApplicationServiceImpl) GroupExists(ctx corectx.Context, query it.GroupExistsQuery) (*it.GroupExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityGroup, c.ResourceScopeOrg); cErr != nil {
		return &it.GroupExistsResult{ClientErrors: *cErr}, nil
	}
	return this.groupSvc.GroupExists(ctx, query)
}

func (this *GroupApplicationServiceImpl) ManageGroupUsers(ctx corectx.Context, cmd it.ManageGroupUsersCommand) (*it.ManageGroupUsersResult, error) {
	if cErr := assertPermission(ctx, "manage_users", c.ResourceIdentityGroup, c.ResourceScopeOrg); cErr != nil {
		return &it.ManageGroupUsersResult{ClientErrors: *cErr}, nil
	}
	return this.groupSvc.ManageGroupUsers(ctx, cmd)
}

func (this *GroupApplicationServiceImpl) SearchGroups(ctx corectx.Context, query it.SearchGroupsQuery) (*it.SearchGroupsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceIdentityGroup, c.ResourceScopeOrg); cErr != nil {
		return &it.SearchGroupsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.Group, *domain.Group]{
		Action:            "search groups",
		FieldResolver:     this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:            this.groupRepo.GetBaseRepo().Schema(),
		DefaultSearchName: "group_list",
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.Group]], error) {
			return this.groupSvc.SearchGroups(ctx, query)
		},
	})
}

func (this *GroupApplicationServiceImpl) UpdateGroup(ctx corectx.Context, cmd it.UpdateGroupCommand) (*it.UpdateGroupResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceIdentityGroup, c.ResourceScopeOrg); cErr != nil {
		return &it.UpdateGroupResult{ClientErrors: *cErr}, nil
	}
	return this.groupSvc.UpdateGroup(ctx, cmd)
}
