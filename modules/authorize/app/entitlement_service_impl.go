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
	this.setEntitlementDefaults(ctx, entitlement)
	entitlement.SetCreatedAt(time.Now())

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = entitlement.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeEntitlement(entitlement)
			return this.assertEntitlementUnique(ctx, entitlement, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateEntitlementResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlement, err = this.entitlementRepo.Create(ctx, *entitlement)
	ft.PanicOnErr(err)

	return &it.CreateEntitlementResult{
		Data:    entitlement,
		HasData: entitlement != nil,
	}, err
}

func (this *EntitlementServiceImpl) EntitlementExists(ctx context.Context, cmd it.EntitlementExistsCommand) (result *it.EntitlementExistsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to check entitlement exists"); e != nil {
			err = e
		}
	}()

	var existsEntitlement bool

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			existsEntitlement, err = this.entitlementRepo.Exists(ctx, it.FindByIdParam{Id: cmd.Id})
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.EntitlementExistsResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.EntitlementExistsResult{
		Data:    existsEntitlement,
		HasData: true,
	}, nil
}

func (this *EntitlementServiceImpl) UpdateEntitlement(ctx context.Context, cmd it.UpdateEntitlementCommand) (result *it.UpdateEntitlementResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update resource"); e != nil {
			err = e
		}
	}()

	entitlement := cmd.ToEntitlement()
	var dbEntitlement *domain.Entitlement

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = entitlement.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbEntitlement, err = this.assertEntitlementExistsById(ctx, *entitlement.Id, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertCorrectEtag(*entitlement.Etag, *dbEntitlement.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeEntitlement(entitlement)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateEntitlementResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := entitlement.Etag
	entitlement.Etag = model.NewEtag()
	entitlement, err = this.entitlementRepo.Update(ctx, *entitlement, *prevEtag)
	ft.PanicOnErr(err)

	return &it.UpdateEntitlementResult{
		Data:    entitlement,
		HasData: entitlement != nil,
	}, err
}

func (this *EntitlementServiceImpl) GetEntitlementById(ctx context.Context, query it.GetEntitlementByIdQuery) (result *it.GetEntitlementByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get entitlement by id"); e != nil {
			err = e
		}
	}()

	var dbEntitlement *domain.Entitlement
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbEntitlement, err = this.assertEntitlementExistsById(ctx, query.Id, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetEntitlementByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetEntitlementByIdResult{
		Data:    dbEntitlement,
		HasData: dbEntitlement != nil,
	}, nil
}

func (this *EntitlementServiceImpl) GetAllEntitlementByIds(ctx context.Context, query it.GetAllEntitlementByIdsQuery) (result *it.GetAllEntitlementByIdsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get all entitlement by ids"); e != nil {
			err = e
		}
	}()

	var dbEntitlements []*domain.Entitlement
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbEntitlements, err = this.entitlementRepo.FindAllByIds(ctx, query)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetAllEntitlementByIdsResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	if len(dbEntitlements) == 0 {
		vErrs.AppendIdNotFound("entitlement")

		return &it.GetAllEntitlementByIdsResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetAllEntitlementByIdsResult{
		Data:    dbEntitlements,
		HasData: dbEntitlements != nil,
	}, nil
}

func (this *EntitlementServiceImpl) SearchEntitlements(ctx context.Context, query it.SearchEntitlementsQuery) (result *it.SearchEntitlementsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list resources"); e != nil {
			err = e
		}
	}()

	query.SetDefaults()
	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.entitlementRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchEntitlementsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	resources, err := this.entitlementRepo.Search(ctx, it.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
	})
	ft.PanicOnErr(err)

	return &it.SearchEntitlementsResult{
		Data:    resources,
		HasData: resources.Items != nil,
	}, nil
}

func (this *EntitlementServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *ft.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *EntitlementServiceImpl) assertEntitlementExistsById(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (dbEntitlement *domain.Entitlement, err error) {
	dbEntitlement, err = this.entitlementRepo.FindById(ctx, it.FindByIdParam{Id: id})
	if dbEntitlement == nil {
		vErrs.AppendIdNotFound("entitlement")
	}
	return
}

func (this *EntitlementServiceImpl) assertEntitlementExistsByName(ctx context.Context, name string, vErrs *ft.ValidationErrors) (dbEntitlement *domain.Entitlement, err error) {
	dbEntitlement, err = this.entitlementRepo.FindByName(ctx, it.FindByNameParam{Name: name})
	if dbEntitlement == nil {
		vErrs.AppendIdNotFound("entitlement")
	}
	return
}

func (this *EntitlementServiceImpl) sanitizeEntitlement(entitlement *domain.Entitlement) {
	if entitlement.Description != nil {
		cleanedName := strings.TrimSpace(*entitlement.Description)
		cleanedName = defense.SanitizePlainText(cleanedName)
		entitlement.Description = &cleanedName
	}
}

func (this *EntitlementServiceImpl) setEntitlementDefaults(ctx context.Context, entitlement *domain.Entitlement) {
	entitlement.SetDefaults()
}

func (this *EntitlementServiceImpl) assertEntitlementUnique(ctx context.Context, entitlement *domain.Entitlement, errors *ft.ValidationErrors) error {
	if errors.Has("name") {
		return nil
	}
	dbEntitlement, err := this.entitlementRepo.FindByName(ctx, it.FindByNameParam{Name: *entitlement.Name})
	ft.PanicOnErr(err)

	if dbEntitlement != nil {
		errors.Append("name", "name already exists")
	}
	return nil
}
