package app

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	constants "github.com/sky-as-code/nikki-erp/modules/authorize/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itAuthorize "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/resource"
)

func NewResourceServiceImpl(resourceRepo it.ResourceRepository, authorizeService itAuthorize.AuthorizeService) it.ResourceService {
	return &ResourceServiceImpl{
		resourceRepo:     resourceRepo,
		authorizeService: authorizeService,
	}
}

type ResourceServiceImpl struct {
	resourceRepo     it.ResourceRepository
	authorizeService itAuthorize.AuthorizeService
}

func (this *ResourceServiceImpl) assertAuthorized(ctx crud.Context, actionName string, resourceName string, scopeRef string, subjectType itAuthorize.SubjectTypeAuthorize, subjectRef string) error {
	isAuthorized, err := this.authorizeService.IsAuthorized(ctx, itAuthorize.IsAuthorizedQuery{
		ActionName:   actionName,
		ResourceName: resourceName,
		ScopeRef:     scopeRef,
		SubjectType:  subjectType,
		SubjectRef:   subjectRef,
	})
	fault.PanicOnErr(err)

	if isAuthorized.ClientError != nil {
		return isAuthorized.ClientError
	}

	if isAuthorized.Decision == nil || *isAuthorized.Decision != itAuthorize.DecisionAllow {
		return &fault.ClientError{
			Code:    "403",
			Details: itAuthorize.DecisionDeny,
		}
	}

	return nil
}

func (this *ResourceServiceImpl) CreateResource(ctx crud.Context, cmd it.CreateResourceCommand) (*it.CreateResourceResult, error) {
	err := this.assertAuthorized(ctx, constants.ActionCreate, constants.ResourceResource, "", itAuthorize.SubjectTypeUser, "01JZQFY6EXRG0959Z95Y2EM3AM")
	fault.PanicOnErr(err)

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
	err := this.assertAuthorized(ctx, constants.ActionDelete, constants.ResourceResource, "", itAuthorize.SubjectTypeUser, "01JZQFY6EXRG0959Z95Y2EM3AM")
	fault.PanicOnErr(err)

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
	err := this.assertAuthorized(ctx, constants.ActionView, constants.ResourceResource, "", itAuthorize.SubjectTypeUser, "01JZQFY6EXRG0959Z95Y2EM3AM")
	fault.PanicOnErr(err)

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
	err := this.assertAuthorized(ctx, constants.ActionView, constants.ResourceResource, "", itAuthorize.SubjectTypeUser, "01JZQFY6EXRG0959Z95Y2EM3AM")
	fault.PanicOnErr(err)

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
	err := this.assertAuthorized(ctx, constants.ActionView, constants.ResourceResource, "", itAuthorize.SubjectTypeUser, "01JZQFY6EXRG0959Z95Y2EM3AM")
	fault.PanicOnErr(err)

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

func (this *ResourceServiceImpl) Exists(ctx crud.Context, cmd it.ExistsResourceQuery) (result *it.ExistsResourceResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "check resource exists"); e != nil {
			err = e
		}
	}()

	var existsResource bool

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			existsResource, err = this.resourceRepo.Exists(ctx, it.ExistsResourceQuery{Id: cmd.Id})
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.ExistsResourceResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.ExistsResourceResult{
		Data:    existsResource,
		HasData: true,
	}, nil
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
