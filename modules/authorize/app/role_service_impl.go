package app

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
)

func NewRoleServiceImpl(roleRepo it.RoleRepository, eventBus event.EventBus) it.RoleService {
	return &RoleServiceImpl{
		roleRepo: roleRepo,
		eventBus: eventBus,
	}
}

type RoleServiceImpl struct {
	roleRepo it.RoleRepository
	eventBus event.EventBus
}

func (this *RoleServiceImpl) CreateRole(ctx context.Context, cmd it.CreateRoleCommand) (result *it.CreateRoleResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create role"); e != nil {
			err = e
		}
	}()

	role := cmd.ToRole()
	err = role.SetDefaults()
	ft.PanicOnErr(err)

	vErrs := role.Validate(false)
	this.assertRoleUnique(ctx, role, &vErrs)
	if vErrs.Count() > 0 {
		return &it.CreateRoleResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	role, err = this.roleRepo.Create(ctx, *role)
	ft.PanicOnErr(err)

	return &it.CreateRoleResult{Data: role}, err
}

func (this *RoleServiceImpl) assertRoleUnique(ctx context.Context, role *domain.Role, errors *ft.ValidationErrors) {
	if errors.Has("name") {
		return
	}
	dbRole, err := this.roleRepo.FindByName(ctx, it.FindByNameParam{Name: *role.Name})
	ft.PanicOnErr(err)

	if dbRole != nil {
		errors.Append("name", "name already exists")
	}
}

// func (this *ResourceServiceImpl) UpdateResource(ctx context.Context, cmd it.UpdateResourceCommand) (result *it.UpdateResourceResult, err error) {
// 	defer func() {
// 		if e := ft.RecoverPanic(recover(), "failed to update resource"); e != nil {
// 			err = e
// 		}
// 	}()

// 	resource := cmd.ToResource()

// 	vErrs := resource.Validate(true)
// 	if resource.Name != nil {
// 		this.assertResourceUnique(ctx, resource, &vErrs)
// 	}
// 	if vErrs.Count() > 0 {
// 		return &it.UpdateResourceResult{
// 			ClientError: vErrs.ToClientError(),
// 		}, nil
// 	}

// 	dbResource, err := this.resourceRepo.FindById(ctx, it.FindByIdParam{Id: *resource.Id})
// 	ft.PanicOnErr(err)

// 	if dbResource == nil {
// 		vErrs = ft.NewValidationErrors()
// 		vErrs.Append("id", "resource not found")

// 		return &it.UpdateResourceResult{
// 			ClientError: vErrs.ToClientError(),
// 		}, nil

// 	} else if *dbResource.Etag != *resource.Etag {
// 		vErrs = ft.NewValidationErrors()
// 		vErrs.Append("etag", "resource has been modified by another process")

// 		return &it.UpdateResourceResult{
// 			ClientError: vErrs.ToClientError(),
// 		}, nil
// 	}

// 	resource.Etag = model.NewEtag()
// 	resource, err = this.resourceRepo.Update(ctx, *resource)
// 	ft.PanicOnErr(err)

// 	return &it.UpdateResourceResult{Data: resource}, err
// }

// func (this *ResourceServiceImpl) GetResourceByName(ctx context.Context, cmd it.GetResourceByNameCommand) (result *it.GetResourceByNameResult, err error) {
// 	defer func() {
// 		if e := ft.RecoverPanic(recover(), "failed to get resource by name"); e != nil {
// 			err = e
// 		}
// 	}()

// 	resource, err := this.resourceRepo.FindByName(ctx, it.FindByNameParam{Name: cmd.Name})
// 	ft.PanicOnErr(err)

// 	return &it.GetResourceByNameResult{Data: resource}, err
// }

// func (this *ResourceServiceImpl) SearchResources(ctx context.Context, query it.SearchResourcesCommand) (result *it.SearchResourcesResult, err error) {
// 	defer func() {
// 		if e := ft.RecoverPanic(recover(), "failed to list resources"); e != nil {
// 			err = e
// 		}
// 	}()

// 	vErrsModel := query.Validate()
// 	predicate, order, vErrsGraph := this.resourceRepo.ParseSearchGraph(query.Graph)

// 	vErrsModel.Merge(vErrsGraph)

// 	if vErrsModel.Count() > 0 {
// 		return &it.SearchResourcesResult{
// 			ClientError: vErrsModel.ToClientError(),
// 		}, nil
// 	}
// 	query.SetDefaults()

// 	resources, err := this.resourceRepo.Search(ctx, it.SearchParam{
// 		Predicate:   predicate,
// 		Order:       order,
// 		Page:        *query.Page,
// 		Size:        *query.Size,
// 		WithActions: query.WithActions,
// 	})
// 	ft.PanicOnErr(err)

// 	return &it.SearchResourcesResult{
// 		Data: resources,
// 	}, nil
// }
