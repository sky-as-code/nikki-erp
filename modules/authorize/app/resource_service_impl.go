package app

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/event"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
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
		if e := fault.RecoverPanicFailedTo(recover(), "failed to create resource"); e != nil {
			err = e
		}
	}()

	resource := cmd.ToResource()
	this.setResourceDefaults(ctx, resource)

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = resource.Validate(false)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.sanitizeResource(resource)
			return this.assertResourceUnique(ctx, resource, vErrs)
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateResourceResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	resource, err = this.resourceRepo.Create(ctx, *resource)
	fault.PanicOnErr(err)

	return &it.CreateResourceResult{
		Data:    resource,
		HasData: resource != nil,
	}, err
}

func (this *ResourceServiceImpl) UpdateResource(ctx context.Context, cmd it.UpdateResourceCommand) (result *it.UpdateResourceResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "failed to update resource"); e != nil {
			err = e
		}
	}()

	resource := cmd.ToResource()
	var dbResource *domain.Resource

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = resource.Validate(true)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbResource, err = this.assertResourceExistsById(ctx, *resource.Id, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.assertCorrectEtag(*resource.Etag, *dbResource.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.sanitizeResource(resource)
			return nil
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateResourceResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := resource.Etag
	resource.Etag = model.NewEtag()
	resource, err = this.resourceRepo.Update(ctx, *resource, *prevEtag)
	fault.PanicOnErr(err)

	return &it.UpdateResourceResult{
		Data:    resource,
		HasData: resource != nil,
	}, err
}

func (this *ResourceServiceImpl) GetResourceByName(ctx context.Context, query it.GetResourceByNameQuery) (result *it.GetResourceByNameResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "failed to get resource by name"); e != nil {
			err = e
		}
	}()

	var dbResource *domain.Resource
	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbResource, err = this.assertResourceExistsByName(ctx, query.Name, vErrs)
			return err
		}).
		End()
	fault.PanicOnErr(err)

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
		if e := fault.RecoverPanicFailedTo(recover(), "failed to list resources"); e != nil {
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
	fault.PanicOnErr(err)

	return &it.SearchResourcesResult{
		Data:    resources,
		HasData: resources.Items != nil,
	}, nil
}

func (this *ResourceServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *fault.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *ResourceServiceImpl) assertResourceExistsByName(ctx context.Context, name string, vErrs *fault.ValidationErrors) (dbResource *domain.Resource, err error) {
	dbResource, err = this.resourceRepo.FindByName(ctx, it.FindByNameParam{Name: name})
	if dbResource == nil {
		vErrs.AppendNotFound("id", "resource")
	}
	return
}

func (this *ResourceServiceImpl) assertResourceExistsById(ctx context.Context, id model.Id, vErrs *fault.ValidationErrors) (dbResource *domain.Resource, err error) {
	dbResource, err = this.resourceRepo.FindById(ctx, it.FindByIdParam{Id: id})
	if dbResource == nil {
		vErrs.AppendNotFound("id", "resource")
	}
	return
}

func (this *ResourceServiceImpl) sanitizeResource(resource *domain.Resource) {
	if resource.Description != nil {
		resource.Description = util.ToPtr(defense.SanitizePlainText(*resource.Description, true))
	}
}

func (this *ResourceServiceImpl) setResourceDefaults(ctx context.Context, resource *domain.Resource) {
	resource.SetDefaults()
}

func (this *ResourceServiceImpl) assertResourceUnique(ctx context.Context, resource *domain.Resource, vErrs *fault.ValidationErrors) error {
	dbResource, err := this.resourceRepo.FindByName(ctx, it.FindByNameParam{Name: *resource.Name})
	fault.PanicOnErr(err)

	if dbResource != nil {
		vErrs.AppendAlreadyExists("name", "resource name")
	}

	return nil
}
