package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itAct "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

func NewResourceDomainServiceImpl(
	actionRepo itAct.ActionRepository,
	resourceRepo itRes.ResourceRepository,
	cqrsBus cqrs.CqrsBus,
) itRes.ResourceDomainService {
	return &ResourceDomainServiceImpl{
		actionRepo:   actionRepo,
		resourceRepo: resourceRepo,
	}
}

type ResourceDomainServiceImpl struct {
	actionRepo   itAct.ActionRepository
	resourceRepo itRes.ResourceRepository
}

func (this *ResourceDomainServiceImpl) CreateResource(
	ctx corectx.Context, cmd itRes.CreateResourceCommand, options ...corecrud.ServiceCreateOptions[*domain.Resource],
) (*itRes.CreateResourceResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceCreateOptions[*domain.Resource]{})
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Resource, *domain.Resource]{
		Action:         "create resource",
		BaseRepoGetter: this.resourceRepo,
		Data:           cmd,
		ValidateExtra:  validateNewResourceScope,
		AfterValidationSuccess: opts.AfterValidationSuccess,
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
func (this *ResourceDomainServiceImpl) DeleteResource(
	ctx corectx.Context, cmd itRes.DeleteResourceCommand, options ...corecrud.ServiceDeleteOptions,
) (*itRes.DeleteResourceResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceDeleteOptions{})
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:                 "delete resource",
		DbRepoGetter:           this.resourceRepo,
		Cmd:                    dyn.DeleteOneCommand(cmd),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *ResourceDomainServiceImpl) ResourceExists(
	ctx corectx.Context, query itRes.ResourceExistsQuery,
) (*itRes.ResourceExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if resource exists",
		DbRepoGetter: this.resourceRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *ResourceDomainServiceImpl) GetResource(
	ctx corectx.Context, query itRes.GetResourceQuery,
) (*dyn.OpResult[domain.Resource], error) {
	return corecrud.GetOne[domain.Resource](ctx, corecrud.GetOneParam{
		Action:       "get resource",
		DbRepoGetter: this.resourceRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ResourceDomainServiceImpl) SearchResources(
	ctx corectx.Context, query itRes.SearchResourcesQuery, options ...corecrud.ServiceSearchOptions,
) (*itRes.SearchResourcesResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	return corecrud.Search[domain.Resource](ctx, corecrud.SearchParam{
		Action:                 "search resources",
		DbRepoGetter:           this.resourceRepo,
		Query:                  dyn.SearchQuery(query),
		AfterValidationSuccess: opts.AfterValidationSuccess,
	})
}

func (this *ResourceDomainServiceImpl) UpdateResource(
	ctx corectx.Context, cmd itRes.UpdateResourceCommand, options ...corecrud.ServiceUpdateOptions[*domain.Resource],
) (*itRes.UpdateResourceResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceUpdateOptions[*domain.Resource]{})
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Resource, *domain.Resource]{
		Action:                 "update resource",
		DbRepoGetter:           this.resourceRepo,
		Data:                   cmd,
		ValidateExtra:          validateUpdateResourceScope,
		AfterValidationSuccess: opts.AfterValidationSuccess,
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
