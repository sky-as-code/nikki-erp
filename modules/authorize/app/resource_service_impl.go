package app

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
)

func NewResourceServiceImpl(resourceRepo it.ResourceRepository) it.ResourceService {
	return &ResourceServiceImpl{
		resourceRepo: resourceRepo,
	}
}

type ResourceServiceImpl struct {
	resourceRepo it.ResourceRepository
}

func (this *ResourceServiceImpl) CreateResource(ctx crud.Context, cmd it.CreateResourceCommand) (*it.CreateResourceResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.Resource, it.CreateResourceCommand, it.CreateResourceResult]{
		Action:              "create resource",
		Command:             cmd,
		AssertBusinessRules: this.assertResourceUnique,
		RepoCreate:          this.resourceRepo.Create,
		SetDefault:          this.setResourceDefaults,
		Sanitize:            this.sanitizeResource,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.CreateResourceResult {
			return &it.CreateResourceResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Resource) *it.CreateResourceResult {
			return &it.CreateResourceResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ResourceServiceImpl) UpdateResource(ctx crud.Context, cmd it.UpdateResourceCommand) (*it.UpdateResourceResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.Resource, it.UpdateResourceCommand, it.UpdateResourceResult]{
		Action:       "update resource",
		Command:      cmd,
		AssertExists: this.assertResourceExistsById,
		RepoUpdate:   this.resourceRepo.Update,
		Sanitize:     this.sanitizeResource,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.UpdateResourceResult {
			return &it.UpdateResourceResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Resource) *it.UpdateResourceResult {
			return &it.UpdateResourceResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ResourceServiceImpl) DeleteResourceHard(ctx crud.Context, cmd it.DeleteResourceHardByNameQuery) (*it.DeleteResourceHardByNameResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Resource, it.DeleteResourceHardByNameQuery, it.DeleteResourceHardByNameResult]{
		Action:              "delete resource",
		Command:             cmd,
		AssertExists:        this.assertResourceExistsByName,
		AssertBusinessRules: this.assertDeleteRules,
		RepoDelete: func(ctx crud.Context, model *domain.Resource) (int, error) {
			return this.resourceRepo.DeleteHard(ctx, it.DeleteParam{Name: *model.Name})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.DeleteResourceHardByNameResult {
			return &it.DeleteResourceHardByNameResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Resource, deletedCount int) *it.DeleteResourceHardByNameResult {
			return crud.NewSuccessDeletionResult(*model.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *ResourceServiceImpl) GetResourceById(ctx crud.Context, query it.GetResourceByIdQuery) (*it.GetResourceByIdResult, error) {
	return crud.GetOne(ctx, crud.GetOneParam[*domain.Resource, it.GetResourceByIdQuery, it.GetResourceByIdResult]{
		Action:      "get resource by id",
		Query:       query,
		RepoFindOne: this.getResourceByIdFull,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.GetResourceByIdResult {
			return &it.GetResourceByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Resource) *it.GetResourceByIdResult {
			return &it.GetResourceByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})
}

func (this *ResourceServiceImpl) GetResourceByName(ctx crud.Context, query it.GetResourceByNameQuery) (*it.GetResourceByNameResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Resource, it.GetResourceByNameQuery, it.GetResourceByNameResult]{
		Action:      "get resource by name",
		Query:       query,
		RepoFindOne: this.getResourceByNameFull,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.GetResourceByNameResult {
			return &it.GetResourceByNameResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Resource) *it.GetResourceByNameResult {
			return &it.GetResourceByNameResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ResourceServiceImpl) SearchResources(ctx crud.Context, query it.SearchResourcesQuery) (*it.SearchResourcesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.Resource, it.SearchResourcesQuery, it.SearchResourcesResult]{
		Action: "search resources",
		Query:  query,
		SetQueryDefaults: func(query *it.SearchResourcesQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.resourceRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query it.SearchResourcesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Resource], error) {
			return this.resourceRepo.Search(ctx, it.SearchParam{
				Predicate:   predicate,
				Order:       order,
				Page:        *query.Page,
				Size:        *query.Size,
				WithActions: query.WithActions,
			})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.SearchResourcesResult {
			return &it.SearchResourcesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.Resource]) *it.SearchResourcesResult {
			return &it.SearchResourcesResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *ResourceServiceImpl) getResourceByIdFull(ctx crud.Context, query it.GetResourceByIdQuery, vErrs *fault.ValidationErrors) (dbResource *domain.Resource, err error) {
	dbResource, err = this.resourceRepo.FindById(ctx, query)
	fault.PanicOnErr(err)

	if dbResource == nil {
		vErrs.AppendNotFound("resource_id", "resource")
	}
	return
}

func (this *ResourceServiceImpl) getResourceByNameFull(ctx crud.Context, query it.GetResourceByNameQuery, vErrs *fault.ValidationErrors) (dbResource *domain.Resource, err error) {
	dbResource, err = this.resourceRepo.FindByName(ctx, query)
	fault.PanicOnErr(err)

	if dbResource == nil {
		vErrs.AppendNotFound("resource_name", "resource")
	}
	return
}

func (this *ResourceServiceImpl) assertResourceExistsByName(ctx crud.Context, resource *domain.Resource, vErrs *fault.ValidationErrors) (dbResource *domain.Resource, err error) {
	dbResource, err = this.resourceRepo.FindByName(ctx, it.FindByNameParam{Name: *resource.Name})
	fault.PanicOnErr(err)

	if dbResource == nil {
		vErrs.AppendNotFound("resource_name", "resource")
	}
	return
}

func (this *ResourceServiceImpl) assertResourceExistsById(ctx crud.Context, resource *domain.Resource, vErrs *fault.ValidationErrors) (dbResource *domain.Resource, err error) {
	dbResource, err = this.resourceRepo.FindById(ctx, it.FindByIdParam{Id: *resource.Id})
	fault.PanicOnErr(err)

	if dbResource == nil {
		vErrs.AppendNotFound("resource_id", "resource")
	}
	return
}

func (this *ResourceServiceImpl) sanitizeResource(resource *domain.Resource) {
	if resource.Description != nil {
		resource.Description = util.ToPtr(defense.SanitizePlainText(*resource.Description, true))
	}

	if resource.Name != nil {
		resource.Name = util.ToPtr(defense.SanitizePlainText(*resource.Name, true))
	}
}

func (this *ResourceServiceImpl) setResourceDefaults(resource *domain.Resource) {
	resource.SetDefaults()
}

func (this *ResourceServiceImpl) assertResourceUnique(ctx crud.Context, resource *domain.Resource, vErrs *fault.ValidationErrors) error {
	dbResource, err := this.resourceRepo.FindByName(ctx, it.FindByNameParam{Name: *resource.Name})
	fault.PanicOnErr(err)

	if dbResource != nil {
		vErrs.AppendAlreadyExists("resource_name", "resource name")
	}

	return nil
}

func (this *ResourceServiceImpl) assertDeleteRules(ctx crud.Context, _ it.DeleteResourceHardByNameQuery, deletedResource *domain.Resource, vErrs *fault.ValidationErrors) error {
	return this.assertConstraintViolated(deletedResource, vErrs)
}

func (this *ResourceServiceImpl) assertConstraintViolated(resource *domain.Resource, vErrs *fault.ValidationErrors) error {
	if len(resource.Actions) > 0 {
		for _, action := range resource.Actions {
			vErrs.AppendConstraintViolated("actions", *action.Name)
		}
	}

	if len(resource.Entitlements) > 0 {
		for _, entitlement := range resource.Entitlements {
			vErrs.AppendConstraintViolated("entitlements", *entitlement.Name)
		}
	}

	return nil
}
