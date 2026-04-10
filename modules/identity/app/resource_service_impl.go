package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itRes "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

func NewResourceServiceImpl(
	resourceRepo itRes.ResourceRepository,
	cqrsBus cqrs.CqrsBus,
) itRes.ResourceService {
	return &ResourceServiceImpl{cqrsBus: cqrsBus, resourceRepo: resourceRepo}
}

type ResourceServiceImpl struct {
	cqrsBus      cqrs.CqrsBus
	resourceRepo itRes.ResourceRepository
}

func (this *ResourceServiceImpl) CreateResource(
	ctx corectx.Context, cmd itRes.CreateResourceCommand,
) (*itRes.CreateResourceResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Resource, *domain.Resource]{
		Action:         "create resource",
		BaseRepoGetter: this.resourceRepo,
		Data:           cmd,
		ValidateExtra:  validateNewResourceScope,
	})
}
func validateNewResourceScope(ctx corectx.Context, inputResrc *domain.Resource, vErrs *ft.ClientErrors) error {
	minScope := domain.AuthzScopeWidth(*inputResrc.GetMinScope())
	maxScope := domain.AuthzScopeWidth(*inputResrc.GetMaxScope())
	if minScope > maxScope {
		vErrs.Append(*ft.NewValidationError(domain.ResourceFieldMinScope, "err_min_scope_greater_than_max_scope", "min_scope must be less than or equal to max_scope"))
	}
	return nil
}
func (this *ResourceServiceImpl) DeleteResource(
	ctx corectx.Context, cmd itRes.DeleteResourceCommand,
) (*itRes.DeleteResourceResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete resource",
		DbRepoGetter: this.resourceRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ResourceServiceImpl) ResourceExists(
	ctx corectx.Context, query itRes.ResourceExistsQuery,
) (*itRes.ResourceExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if resource exists",
		DbRepoGetter: this.resourceRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *ResourceServiceImpl) GetResource(
	ctx corectx.Context, query itRes.GetResourceQuery,
) (*itRes.GetResourceResult, error) {
	return corecrud.GetOne[domain.Resource](ctx, corecrud.GetOneParam{
		Action:       "get resource",
		DbRepoGetter: this.resourceRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ResourceServiceImpl) SearchResources(
	ctx corectx.Context, query itRes.SearchResourcesQuery,
) (*itRes.SearchResourcesResult, error) {
	return corecrud.Search[domain.Resource](ctx, corecrud.SearchParam{
		Action:       "search resources",
		DbRepoGetter: this.resourceRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *ResourceServiceImpl) UpdateResource(
	ctx corectx.Context, cmd itRes.UpdateResourceCommand,
) (*itRes.UpdateResourceResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Resource, *domain.Resource]{
		Action:        "update resource",
		DbRepoGetter:  this.resourceRepo,
		Data:          cmd,
		ValidateExtra: validateUpdateResourceScope,
	})
}

func validateUpdateResourceScope(ctx corectx.Context, inputResrc *domain.Resource, foundResrc *domain.Resource, vErrs *ft.ClientErrors) error {
	inputMinScope := inputResrc.GetMinScope()
	foundMaxScope := foundResrc.GetMaxScope()
	if inputMinScope != nil && domain.AuthzScopeWidth(*inputMinScope) > domain.AuthzScopeWidth(*foundMaxScope) {
		vErrs.Append(*ft.NewValidationError(domain.ResourceFieldMinScope, "identity.err_min_scope_greater_than_max_scope", "min_scope must be less than or equal to max_scope"))
	}
	inputMaxScope := inputResrc.GetMaxScope()
	foundMinScope := foundResrc.GetMinScope()
	if inputMaxScope != nil && domain.AuthzScopeWidth(*inputMaxScope) < domain.AuthzScopeWidth(*foundMinScope) {
		vErrs.Append(*ft.NewValidationError(domain.ResourceFieldMaxScope, "identity.err_max_scope_less_than_min_scope", "max_scope must be greater than or equal to min_scope"))
	}
	return nil
}
