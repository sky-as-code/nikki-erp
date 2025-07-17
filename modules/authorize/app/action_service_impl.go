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
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
)

func NewActionServiceImpl(actionRepo it.ActionRepository, eventBus event.EventBus) it.ActionService {
	return &ActionServiceImpl{
		actionRepo: actionRepo,
		eventBus:   eventBus,
	}
}

type ActionServiceImpl struct {
	actionRepo it.ActionRepository
	eventBus   event.EventBus
}

func (this *ActionServiceImpl) CreateAction(ctx context.Context, cmd it.CreateActionCommand) (result *it.CreateActionResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create action"); e != nil {
			err = e
		}
	}()

	action := cmd.ToAction()
	this.setActionDefaults(ctx, action)
	action.SetCreatedAt(time.Now())

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = action.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeAction(action)
			return this.assertActionUnique(ctx, action, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateActionResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	action, err = this.actionRepo.Create(ctx, *action)
	ft.PanicOnErr(err)

	return &it.CreateActionResult{
		Data:    action,
		HasData: action != nil,
	}, err
}

func (this *ActionServiceImpl) UpdateAction(ctx context.Context, cmd it.UpdateActionCommand) (result *it.UpdateActionResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update action"); e != nil {
			err = e
		}
	}()

	action := cmd.ToAction()
	var dbAction *domain.Action

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = action.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbAction, err = this.assertActionExists(ctx, *action.Id, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertCorrectEtag(*action.Etag, *dbAction.Etag, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.sanitizeAction(action)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateActionResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := action.Etag
	action.Etag = model.NewEtag()
	action, err = this.actionRepo.Update(ctx, *action, *prevEtag)
	ft.PanicOnErr(err)

	return &it.UpdateActionResult{
		Data:    action,
		HasData: action != nil,
	}, err
}

func (this *ActionServiceImpl) GetActionById(ctx context.Context, query it.GetActionByIdQuery) (result *it.GetActionByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get action by id"); e != nil {
			err = e
		}
	}()

	var dbAction *domain.Action
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbAction, err = this.assertActionExists(ctx, query.Id, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.GetActionByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetActionByIdResult{
		Data:    dbAction,
		HasData: dbAction != nil,
	}, nil
}

func (this *ActionServiceImpl) SearchActions(ctx context.Context, query it.SearchActionsCommand) (result *it.SearchActionsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list actions"); e != nil {
			err = e
		}
	}()

	query.SetDefaults()
	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.actionRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchActionsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	actions, err := this.actionRepo.Search(ctx, it.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
	})
	ft.PanicOnErr(err)

	return &it.SearchActionsResult{
		Data:    actions,
		HasData: actions.Items != nil,
	}, nil
}

func (this *ActionServiceImpl) sanitizeAction(action *domain.Action) {
	if action.Description != nil {
		cleanedName := strings.TrimSpace(*action.Description)
		cleanedName = defense.SanitizePlainText(cleanedName)
		action.Description = &cleanedName
	}
}

func (this *ActionServiceImpl) setActionDefaults(ctx context.Context, action *domain.Action) {
	action.SetDefaults()
}

func (this *ActionServiceImpl) assertActionUnique(ctx context.Context, action *domain.Action, vErrs *ft.ValidationErrors) error {
	if vErrs.Has("name") {
		return nil
	}
	dbAction, err := this.actionRepo.FindByName(ctx, it.FindByNameParam{Name: *action.Name, ResourceId: *action.ResourceId})
	ft.PanicOnErr(err)

	if dbAction != nil {
		vErrs.AppendAlreadyExists("name", "name")
	}
	return nil
}

func (this *ActionServiceImpl) assertActionExists(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) (dbAction *domain.Action, err error) {
	dbAction, err = this.actionRepo.FindById(ctx, it.FindByIdParam{Id: id})
	if dbAction == nil {
		vErrs.AppendIdNotFound("action")
	}
	return
}

func (this *ActionServiceImpl) assertCorrectEtag(updatedEtag model.Etag, dbEtag model.Etag, vErrs *ft.ValidationErrors) {
	if updatedEtag != dbEtag {
		vErrs.AppendEtagMismatched()
	}
}
