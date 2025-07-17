package app

import (
	"context"
	"strings"
	"time"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
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
	this.setResourceDefaults(ctx, resource)
	resource.SetCreatedAt(time.Now())

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = resource.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeResource(resource)
			return this.assertResourceUnique(ctx, resource, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateResourceResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	resource, err = this.resourceRepo.Create(ctx, *resource)
	ft.PanicOnErr(err)

	return &it.CreateResourceResult{
		Data:    resource,
		HasData: resource != nil,
	}, err
}

func (this *ResourceServiceImpl) UpdateResource(ctx context.Context, cmd it.UpdateResourceCommand) (result *it.UpdateResourceResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update resource"); e != nil {
			err = e
		}
	}()

	resource := cmd.ToResource()
	var dbResource *domain.Resource

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = resource.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbResource, err = this.assertResourceExistsById(ctx, *resource.Id, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertCorrectEtag(*resource.Etag, *dbResource.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeResource(resource)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateResourceResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := resource.Etag
	resource.Etag = model.NewEtag()
	resource, err = this.resourceRepo.Update(ctx, *resource, *prevEtag)
	ft.PanicOnErr(err)

	return &it.UpdateResourceResult{
		Data:    resource,
		HasData: resource != nil,
	}, err
}

func (this *ResourceServiceImpl) GetResourceByName(ctx context.Context, query it.GetResourceByNameQuery) (result *it.GetResourceByNameResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get resource by name"); e != nil {
			err = e
		}
	}()

	var dbResource *domain.Resource
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbResource, err = this.assertResourceExistsByName(ctx, query.Name, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetResourceByNameResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetResourceByNameResult{
		Data:    dbResource,
		HasData: dbResource != nil,
	}, nil
}

func (this *ResourceServiceImpl) SearchResources(ctx context.Context, query it.SearchResourcesQuery) (result *it.SearchResourcesResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list resources"); e != nil {
			err = e
		}
	}()

	query.SetDefaults()
	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.resourceRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchResourcesResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	resources, err := this.resourceRepo.Search(ctx, it.SearchParam{
		Predicate:   predicate,
		Order:       order,
		Page:        *query.Page,
		Size:        *query.Size,
		WithActions: query.WithActions,
	})
	ft.PanicOnErr(err)

	return &it.SearchResourcesResult{
		Data:    resources,
		HasData: resources.Items != nil,
	}, nil
}

func (this *ResourceServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *ft.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *ResourceServiceImpl) assertResourceExistsByName(ctx context.Context, name string, vErrs *ft.ValidationErrors) (dbResource *domain.Resource, err error) {
	dbResource, err = this.resourceRepo.FindByName(ctx, it.FindByNameParam{Name: name})
	if dbResource == nil {
		vErrs.AppendIdNotFound("resource")
	}
	return
}

func (this *ResourceServiceImpl) assertResourceExistsById(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (dbResource *domain.Resource, err error) {
	dbResource, err = this.resourceRepo.FindById(ctx, it.FindByIdParam{Id: id})
	if dbResource == nil {
		vErrs.AppendIdNotFound("resource")
	}
	return
}

func (this *ResourceServiceImpl) sanitizeResource(resource *domain.Resource) {
	if resource.Description != nil {
		cleanedName := strings.TrimSpace(*resource.Description)
		cleanedName = defense.SanitizePlainText(cleanedName)
		resource.Description = &cleanedName
	}
}

func (this *ResourceServiceImpl) setResourceDefaults(ctx context.Context, resource *domain.Resource) {
	resource.SetDefaults()
}

func (this *ResourceServiceImpl) assertResourceUnique(ctx context.Context, resource *domain.Resource, vErrs *ft.ValidationErrors) error {
	if vErrs.Has("name") {
		return nil
	}

	dbResource, err := this.resourceRepo.FindByName(ctx, it.FindByNameParam{Name: *resource.Name})
	ft.PanicOnErr(err)

	if dbResource != nil {
		vErrs.Append("name", "name already exists")
	}

	return nil
}
