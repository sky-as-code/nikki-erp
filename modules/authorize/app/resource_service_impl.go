package app

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
)

func NewResourceServiceImpl(resourceRepo it.ResourceRepository, eventBus event.EventBus) it.ResourceService {
	return &ResourceServiceImpl{
		resourceRepo: resourceRepo,
		eventBus:     eventBus,
	}
}

type ResourceServiceImpl struct {
	resourceRepo it.ResourceRepository
	eventBus     event.EventBus
}

func (this *ResourceServiceImpl) CreateResource(ctx context.Context, cmd it.CreateResourceCommand) (result *it.CreateResourceResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create resource"); e != nil {
			err = e
		}
	}()

	resource := cmd.ToResource()
	err = resource.SetDefaults()
	ft.PanicOnErr(err)

	vErrs := resource.Validate(false)
	this.assertResourceUnique(ctx, resource, &vErrs)
	if vErrs.Count() > 0 {
		return &it.CreateResourceResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	resource, err = this.resourceRepo.Create(ctx, *resource)
	ft.PanicOnErr(err)

	return &it.CreateResourceResult{Data: resource}, err
}

func (this *ResourceServiceImpl) assertResourceUnique(ctx context.Context, resource *domain.Resource, errors *ft.ValidationErrors) {
	if errors.Has("name") {
		return
	}
	dbResource, err := this.resourceRepo.FindByName(ctx, it.FindByNameParam{Name: *resource.Name})
	ft.PanicOnErr(err)

	if dbResource != nil {
		errors.Append("name", "name already exists")
	}
}

func (this *ResourceServiceImpl) UpdateResource(ctx context.Context, cmd it.UpdateResourceCommand) (result *it.UpdateResourceResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update resource"); e != nil {
			err = e
		}
	}()

	resource := cmd.ToResource()

	vErrs := resource.Validate(true)
	if vErrs.Count() > 0 {
		return &it.UpdateResourceResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbResource, err := this.resourceRepo.FindById(ctx, it.FindByIdParam{Id: *resource.Id})
	ft.PanicOnErr(err)

	if dbResource == nil {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("id", "resource not found")

		return &it.UpdateResourceResult{
			ClientError: vErrs.ToClientError(),
		}, nil

	} else if *dbResource.Etag != *resource.Etag {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("etag", "resource has been modified by another process")

		return &it.UpdateResourceResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	resource.Etag = model.NewEtag()
	resource, err = this.resourceRepo.Update(ctx, *resource)
	ft.PanicOnErr(err)

	return &it.UpdateResourceResult{Data: resource}, err
}

func (this *ResourceServiceImpl) GetResourceByName(ctx context.Context, query it.GetResourceByNameQuery) (result *it.GetResourceByNameResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get resource by name"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetResourceByNameResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	resource, err := this.resourceRepo.FindByName(ctx, it.FindByNameParam{Name: query.Name})
	ft.PanicOnErr(err)

	if resource == nil {
		vErrs.Append("name", "resource not found")
		return &it.GetResourceByNameResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetResourceByNameResult{Data: resource}, err
}

func (this *ResourceServiceImpl) SearchResources(ctx context.Context, query it.SearchResourcesQuery) (result *it.SearchResourcesResult, err error) {
	defer func() {	
		if e := ft.RecoverPanic(recover(), "failed to list resources"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.resourceRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchResourcesResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	resources, err := this.resourceRepo.Search(ctx, it.SearchParam{
		Predicate:   predicate,
		Order:       order,
		Page:        *query.Page,
		Size:        *query.Size,
		WithActions: query.WithActions,
	})
	ft.PanicOnErr(err)

	return &it.SearchResourcesResult{
		Data: resources,
	}, nil
}
