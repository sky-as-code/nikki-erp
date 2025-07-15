package app

import (
	"context"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
)

func NewEntitlementServiceImpl(entitlementRepo it.EntitlementRepository, eventBus event.EventBus) it.EntitlementService {
	return &EntitlementServiceImpl{
		entitlementRepo: entitlementRepo,
		eventBus:        eventBus,
	}
}

type EntitlementServiceImpl struct {
	entitlementRepo it.EntitlementRepository
	eventBus        event.EventBus
}

func (this *EntitlementServiceImpl) CreateEntitlement(ctx context.Context, cmd it.CreateEntitlementCommand) (result *it.CreateEntitlementResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create entitlement"); e != nil {
			err = e
		}
	}()

	entitlement := cmd.ToEntitlement()
	err = entitlement.SetDefaults()
	ft.PanicOnErr(err)
	entitlement.SetCreatedAt(time.Now())

	vErrs := entitlement.Validate(false)
	this.assertEntitlementUnique(ctx, entitlement, &vErrs)
	if vErrs.Count() > 0 {
		return &it.CreateEntitlementResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlement, err = this.entitlementRepo.Create(ctx, *entitlement)
	ft.PanicOnErr(err)

	return &it.CreateEntitlementResult{Data: entitlement}, err
}

func (this *EntitlementServiceImpl) assertEntitlementUnique(ctx context.Context, entitlement *domain.Entitlement, errors *ft.ValidationErrors) {
	if errors.Has("name") {
		return
	}
	dbEntitlement, err := this.entitlementRepo.FindByName(ctx, it.FindByNameParam{Name: *entitlement.Name})
	ft.PanicOnErr(err)

	if dbEntitlement != nil {
		errors.Append("name", "name already exists")
	}
}

func (this *EntitlementServiceImpl) EntitlementExists(ctx context.Context, cmd it.EntitlementExistsCommand) (result *it.EntitlementExistsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to check entitlement exists"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	if vErrs.Count() > 0 {
		return &it.EntitlementExistsResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	exists, err := this.entitlementRepo.Exists(ctx, it.FindByIdParam{Id: cmd.Id})
	ft.PanicOnErr(err)

	return &it.EntitlementExistsResult{
		Data:    exists,
		HasData: true,
	}, nil
}

func (this *EntitlementServiceImpl) UpdateEntitlement(ctx context.Context, cmd it.UpdateEntitlementCommand) (result *it.UpdateEntitlementResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update entitlement"); e != nil {
			err = e
		}
	}()

	entitlement := cmd.ToEntitlement()

	vErrs := entitlement.Validate(true)
	if entitlement.Name != nil {
		this.assertEntitlementUnique(ctx, entitlement, &vErrs)
	}
	if vErrs.Count() > 0 {
		return &it.UpdateEntitlementResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbEntitlement, err := this.entitlementRepo.FindById(ctx, it.FindByIdParam{Id: *entitlement.Id})
	ft.PanicOnErr(err)

	if dbEntitlement == nil {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("id", "entitlement not found")

		return &it.UpdateEntitlementResult{
			ClientError: vErrs.ToClientError(),
		}, nil

	} else if *dbEntitlement.Etag != *entitlement.Etag {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("etag", "entitlement has been modified by another process")

		return &it.UpdateEntitlementResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlement.Etag = model.NewEtag()
	entitlement, err = this.entitlementRepo.Update(ctx, *entitlement)
	ft.PanicOnErr(err)

	return &it.UpdateEntitlementResult{Data: entitlement}, err
}

func (this *EntitlementServiceImpl) GetEntitlementById(ctx context.Context, query it.GetEntitlementByIdQuery) (result *it.GetEntitlementByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get entitlement by id"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetEntitlementByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlement, err := this.entitlementRepo.FindById(ctx, query)
	ft.PanicOnErr(err)

	if entitlement == nil {
		vErrs.Append("id", "entitlement not found")
		return &it.GetEntitlementByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetEntitlementByIdResult{
		Data: entitlement,
	}, nil
}

func (this *EntitlementServiceImpl) GetAllEntitlementByIds(ctx context.Context, query it.GetAllEntitlementByIdsQuery) (result *it.GetAllEntitlementByIdsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get all entitlement by ids"); e != nil {
			err = e
		}
	}()

	entitlements, err := this.entitlementRepo.FindAllByIds(ctx, query)
	ft.PanicOnErr(err)

	return &it.GetAllEntitlementByIdsResult{
		Data: entitlements,
	}, nil
}

func (this *EntitlementServiceImpl) SearchEntitlements(ctx context.Context, query it.SearchEntitlementsQuery) (result *it.SearchEntitlementsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list resources"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.entitlementRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchEntitlementsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	entitlements, err := this.entitlementRepo.Search(ctx, it.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
	})
	ft.PanicOnErr(err)

	return &it.SearchEntitlementsResult{
		Data: entitlements,
	}, nil
}
