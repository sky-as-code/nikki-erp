package app

import (
	"go.uber.org/dig"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itAct "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
	itExt "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/external"
	itRes "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

type NewResourceApplicationServiceImplParam struct {
	dig.In

	ActionSvc    itAct.ActionDomainService
	ActionRepo   itAct.ActionRepository
	ResourceSvc  itRes.ResourceDomainService
	ResourceRepo itRes.ResourceRepository
	UserPrefSvc  itExt.UserPreferenceUiDomainService
}

func NewResourceApplicationServiceImpl(param NewResourceApplicationServiceImplParam) itRes.ResourceAppService {
	return &ResourceApplicationServiceImpl{
		actionSvc:    param.ActionSvc,
		actionRepo:   param.ActionRepo,
		resourceSvc:  param.ResourceSvc,
		resourceRepo: param.ResourceRepo,
		userPrefSvc:  param.UserPrefSvc,
	}
}

type ResourceApplicationServiceImpl struct {
	actionSvc    itAct.ActionDomainService
	actionRepo   itAct.ActionRepository
	resourceSvc  itRes.ResourceDomainService
	resourceRepo itRes.ResourceRepository
	userPrefSvc  itExt.UserPreferenceUiDomainService
}

func (this *ResourceApplicationServiceImpl) CreateResource(ctx corectx.Context, cmd itRes.CreateResourceCommand) (*itRes.CreateResourceResult, error) {
	if cErr := assertPermission(ctx, "create", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itRes.CreateResourceResult{ClientErrors: *cErr}, nil
	}
	return this.resourceSvc.CreateResource(ctx, cmd)
}

func (this *ResourceApplicationServiceImpl) DeleteResource(ctx corectx.Context, cmd itRes.DeleteResourceCommand) (*itRes.DeleteResourceResult, error) {
	if cErr := assertPermission(ctx, "delete", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itRes.DeleteResourceResult{ClientErrors: *cErr}, nil
	}
	return this.resourceSvc.DeleteResource(ctx, cmd)
}

func (this *ResourceApplicationServiceImpl) ResourceExists(ctx corectx.Context, query itRes.ResourceExistsQuery) (*itRes.ResourceExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itRes.ResourceExistsResult{ClientErrors: *cErr}, nil
	}
	return this.resourceSvc.ResourceExists(ctx, query)
}

func (this *ResourceApplicationServiceImpl) GetResource(ctx corectx.Context, query itRes.GetResourceQuery) (*itRes.GetResourceResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itRes.GetResourceResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.Resource, *domain.Resource]{
		Action: "get resource",
		Schema: this.resourceRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.Resource], error) {
			return this.resourceSvc.GetResource(ctx, query)
		},
	})
}

func (this *ResourceApplicationServiceImpl) SearchResources(ctx corectx.Context, query itRes.SearchResourcesQuery) (*itRes.SearchResourcesResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itRes.SearchResourcesResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.Resource, *domain.Resource]{
		Action:            "search resources",
		FieldResolver:     this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:            this.resourceRepo.GetBaseRepo().Schema(),
		DefaultSearchName: "resource_list",
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.Resource]], error) {
			return this.resourceSvc.SearchResources(ctx, query)
		},
	})
}

func (this *ResourceApplicationServiceImpl) UpdateResource(ctx corectx.Context, cmd itRes.UpdateResourceCommand) (*itRes.UpdateResourceResult, error) {
	if cErr := assertPermission(ctx, "update", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itRes.UpdateResourceResult{ClientErrors: *cErr}, nil
	}
	return this.resourceSvc.UpdateResource(ctx, cmd)
}
