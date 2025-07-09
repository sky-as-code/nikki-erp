package app

import (
	"context"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
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
	err = action.SetDefaults()
	ft.PanicOnErr(err)
	action.SetCreatedAt(time.Now())

	vErrs := action.Validate(false)
	this.assertActionUnique(ctx, action, &vErrs)
	if vErrs.Count() > 0 {
		return &it.CreateActionResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	action, err = this.actionRepo.Create(ctx, *action)
	ft.PanicOnErr(err)

	return &it.CreateActionResult{Data: action}, err
}

func (this *ActionServiceImpl) assertActionUnique(ctx context.Context, action *domain.Action, errors *ft.ValidationErrors) {
	if errors.Has("name") {
		return
	}
	dbAction, err := this.actionRepo.FindByName(ctx, it.FindByNameParam{Name: *action.Name})
	ft.PanicOnErr(err)

	if dbAction != nil {
		errors.Append("name", "name already exists")
	}
}

func (this *ActionServiceImpl) UpdateAction(ctx context.Context, cmd it.UpdateActionCommand) (result *it.UpdateActionResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update resource"); e != nil {
			err = e
		}
	}()

	action := cmd.ToAction()

	vErrs := action.Validate(true)
	if action.Name != nil {
		this.assertActionUnique(ctx, action, &vErrs)
	}
	if vErrs.Count() > 0 {
		return &it.UpdateActionResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbAction, err := this.actionRepo.FindById(ctx, it.FindByIdParam{Id: *action.Id})
	ft.PanicOnErr(err)

	if dbAction == nil {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("id", "action not found")

		return &it.UpdateActionResult{
			ClientError: vErrs.ToClientError(),
		}, nil

	} else if *dbAction.Etag != *action.Etag {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("etag", "action has been modified by another process")

		return &it.UpdateActionResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	action.Etag = model.NewEtag()
	action, err = this.actionRepo.Update(ctx, *action)
	ft.PanicOnErr(err)

	return &it.UpdateActionResult{Data: action}, err
}

func (this *ActionServiceImpl) GetActionById(ctx context.Context, query it.GetActionByIdQuery) (result *it.GetActionByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get action by id"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetActionByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	action, err := this.actionRepo.FindById(ctx, query)
	ft.PanicOnErr(err)

	if action == nil {
		vErrs.Append("id", "action not found")
		return &it.GetActionByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.GetActionByIdResult{
		Data: action,
	}, nil
}

func (this *ActionServiceImpl) SearchActions(ctx context.Context, query it.SearchActionsCommand) (result *it.SearchActionsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list resources"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.actionRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchActionsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	actions, err := this.actionRepo.Search(ctx, it.SearchParam{
		Predicate: predicate,
		Order:     order,
		Page:      *query.Page,
		Size:      *query.Size,
	})
	ft.PanicOnErr(err)

	return &it.SearchActionsResult{
		Data: actions,
	}, nil
}
