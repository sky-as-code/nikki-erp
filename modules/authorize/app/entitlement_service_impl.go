package app

import (
	"context"
	"fmt"

	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/event"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itAction "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	itAssignment "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
)

func NewEntitlementServiceImpl(entitlementRepo it.EntitlementRepository, assignmentRepo itAssignment.EntitlementAssignmentRepository, resourceRepo itResource.ResourceRepository, actionRepo itAction.ActionRepository, eventBus event.EventBus) it.EntitlementService {
	return &EntitlementServiceImpl{
		entitlementRepo: entitlementRepo,
		assignmentRepo:  assignmentRepo,
		resourceRepo:    resourceRepo,
		actionRepo:      actionRepo,
		eventBus:        eventBus,
	}
}

type EntitlementServiceImpl struct {
	entitlementRepo it.EntitlementRepository
	assignmentRepo  itAssignment.EntitlementAssignmentRepository
	resourceRepo    itResource.ResourceRepository
	actionRepo      itAction.ActionRepository
	eventBus        event.EventBus
}

func (this *EntitlementServiceImpl) CreateEntitlement(ctx context.Context, cmd it.CreateEntitlementCommand) (result *it.CreateEntitlementResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "create entitlement"); e != nil {
			err = e
		}
	}()

	entitlement := cmd.ToEntitlement()
	this.setEntitlementDefaults(entitlement)

	var resource *domain.Resource
	var action *domain.Action

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = entitlement.Validate(false)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			resource, err = this.assertResourceExists(ctx, *entitlement.ResourceId, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			action, err = this.assertActionExists(ctx, *entitlement.ActionId, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			return this.assertActionExprValid(resource, action, entitlement.ScopeRef, entitlement, vErrs)
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.sanitizeEntitlement(entitlement)
			return this.assertEntitlementUnique(ctx, entitlement, vErrs)
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateEntitlementResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	entitlement, err = this.entitlementRepo.Create(ctx, *entitlement)
	fault.PanicOnErr(err)

	return &it.CreateEntitlementResult{
		Data:    entitlement,
		HasData: entitlement != nil,
	}, err
}

func (this *EntitlementServiceImpl) EntitlementExists(ctx context.Context, cmd it.EntitlementExistsCommand) (result *it.EntitlementExistsResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "check entitlement exists"); e != nil {
			err = e
		}
	}()

	var existsEntitlement bool

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			existsEntitlement, err = this.entitlementRepo.Exists(ctx, it.FindByIdParam{Id: cmd.Id})
			return err
		}).
		End()
	fault.PanicOnErr(err)

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
		if e := fault.RecoverPanicFailedTo(recover(), "update resource"); e != nil {
			err = e
		}
	}()

	entitlement := cmd.ToEntitlement()
	var dbEntitlement *domain.Entitlement

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = entitlement.Validate(true)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbEntitlement, err = this.assertEntitlementExistsById(ctx, *entitlement.Id, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.assertCorrectEtag(*entitlement.Etag, *dbEntitlement.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			this.sanitizeEntitlement(entitlement)
			return nil
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateEntitlementResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := entitlement.Etag
	entitlement.Etag = model.NewEtag()
	entitlement, err = this.entitlementRepo.Update(ctx, *entitlement, *prevEtag)
	fault.PanicOnErr(err)

	return &it.UpdateEntitlementResult{
		Data:    entitlement,
		HasData: entitlement != nil,
	}, err
}

func (this *EntitlementServiceImpl) GetEntitlementById(ctx context.Context, query it.GetEntitlementByIdQuery) (result *it.GetEntitlementByIdResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get entitlement by id"); e != nil {
			err = e
		}
	}()

	var dbEntitlement *domain.Entitlement
	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbEntitlement, err = this.assertEntitlementExistsById(ctx, query.Id, vErrs)
			return err
		}).
		End()
	fault.PanicOnErr(err)

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
		if e := fault.RecoverPanicFailedTo(recover(), "get all entitlement by ids"); e != nil {
			err = e
		}
	}()

	var dbEntitlements []domain.Entitlement
	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			dbEntitlements, err = this.entitlementRepo.FindAllByIds(ctx, query)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetAllEntitlementByIdsResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	if len(dbEntitlements) == 0 {
		vErrs.AppendNotFound("id", "entitlement")

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
		if e := fault.RecoverPanicFailedTo(recover(), "list entitlements"); e != nil {
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

	entitlements, err := this.entitlementRepo.Search(ctx, it.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
	})
	fault.PanicOnErr(err)

	return &it.SearchEntitlementsResult{
		Data:    entitlements,
		HasData: entitlements.Items != nil,
	}, nil
}

func (this *EntitlementServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *fault.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}

func (this *EntitlementServiceImpl) assertEntitlementExistsById(ctx context.Context, id model.Id, vErrs *fault.ValidationErrors) (dbEntitlement *domain.Entitlement, err error) {
	dbEntitlement, err = this.entitlementRepo.FindById(ctx, it.FindByIdParam{Id: id})
	if dbEntitlement == nil {
		vErrs.AppendNotFound("entitlement_id", "entitlement")
	}
	return
}

func (this *EntitlementServiceImpl) sanitizeEntitlement(entitlement *domain.Entitlement) {
	if entitlement.Description != nil {
		entitlement.Description = util.ToPtr(defense.SanitizePlainText(*entitlement.Description, true))
	}
}

func (this *EntitlementServiceImpl) setEntitlementDefaults(entitlement *domain.Entitlement) {
	entitlement.SetDefaults()
}

func (this *EntitlementServiceImpl) assertEntitlementUnique(ctx context.Context, entitlement *domain.Entitlement, vErrs *fault.ValidationErrors) error {
	dbEntitlement, err := this.entitlementRepo.FindByName(ctx, it.FindByNameParam{Name: *entitlement.Name})
	fault.PanicOnErr(err)

	if dbEntitlement != nil {
		vErrs.AppendAlreadyExists("entitlement_name", "entitlement name")
	}
	return nil
}

func (this *EntitlementServiceImpl) assertResourceExists(ctx context.Context, id model.Id, vErrs *fault.ValidationErrors) (resource *domain.Resource, err error) {
	resource, err = this.resourceRepo.FindById(ctx, itResource.FindByIdParam{Id: id})
	fault.PanicOnErr(err)

	if resource == nil {
		vErrs.AppendNotFound("resource_id", "resource")
	}
	return resource, err
}

func (this *EntitlementServiceImpl) assertActionExists(ctx context.Context, id model.Id, vErrs *fault.ValidationErrors) (action *domain.Action, err error) {
	action, err = this.actionRepo.FindById(ctx, itAction.FindByIdParam{Id: id})
	fault.PanicOnErr(err)

	if action == nil {
		vErrs.AppendNotFound("action_id", "action")
	}
	return action, err
}

func (this *EntitlementServiceImpl) assertActionExprValid(resource *domain.Resource, action *domain.Action, scopeRef *string, entitlement *domain.Entitlement, vErrs *fault.ValidationErrors) error {
	var actionName *string
	if action != nil {
		actionName = action.Name
	}

	var resourceName *string
	if resource != nil {
		resourceName = resource.Name
	}

	scopeExpr := "*"
	if scopeRef != nil {
		scopeExpr = *scopeRef
	}

	actionDefault := "*"
	if actionName != nil {
		actionDefault = *actionName
	}

	resourceDefault := "*"
	if resourceName != nil {
		resourceDefault = *resourceName
	}

	actionExpr := fmt.Sprintf("%s:%s.%s", actionDefault, scopeExpr, resourceDefault)

	if entitlement.ActionExpr != nil {
		if *entitlement.ActionExpr != actionExpr {
			vErrs.Append("action_expr", "action_expr is not valid")
		}
	}

	return nil
}
