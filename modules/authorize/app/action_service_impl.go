package app

import (
	"github.com/sky-as-code/nikki-erp/common/defense"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/event"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
)

func NewActionServiceImpl(actionRepo it.ActionRepository, resourceRepo itResource.ResourceRepository, eventBus event.EventBus) it.ActionService {
	return &ActionServiceImpl{
		actionRepo:   actionRepo,
		resourceRepo: resourceRepo,
	}
}

type ActionServiceImpl struct {
	actionRepo   it.ActionRepository
	resourceRepo itResource.ResourceRepository
}

func (this *ActionServiceImpl) CreateAction(ctx crud.Context, cmd it.CreateActionCommand) (*it.CreateActionResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.Action, it.CreateActionCommand, it.CreateActionResult]{
		Action:              "create action",
		Command:             cmd,
		AssertBusinessRules: this.assertBusinessRuleCreateAction,
		RepoCreate:          this.actionRepo.Create,
		SetDefault:          this.setActionDefaults,
		Sanitize:            this.sanitizeAction,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.CreateActionResult {
			return &it.CreateActionResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Action) *it.CreateActionResult {
			return &it.CreateActionResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ActionServiceImpl) UpdateAction(ctx crud.Context, cmd it.UpdateActionCommand) (*it.UpdateActionResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.Action, it.UpdateActionCommand, it.UpdateActionResult]{
		Action:       "update action",
		Command:      cmd,
		AssertExists: this.assertActionExistsById,
		RepoUpdate:   this.actionRepo.Update,
		Sanitize:     this.sanitizeAction,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.UpdateActionResult {
			return &it.UpdateActionResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Action) *it.UpdateActionResult {
			return &it.UpdateActionResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ActionServiceImpl) DeleteActionHard(ctx crud.Context, cmd it.DeleteActionHardByIdCommand) (*it.DeleteActionHardByIdResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Action, it.DeleteActionHardByIdCommand, it.DeleteActionHardByIdResult]{
		Action:              "delete action",
		Command:             cmd,
		AssertExists:        this.assertActionExistsById,
		AssertBusinessRules: this.assertBusinessRuleDeleteAction,
		RepoDelete: func(ctx crud.Context, model *domain.Action) (int, error) {
			return this.actionRepo.DeleteHard(ctx, it.DeleteParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.DeleteActionHardByIdResult {
			return &it.DeleteActionHardByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Action, deletedCount int) *it.DeleteActionHardByIdResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *ActionServiceImpl) GetActionById(ctx crud.Context, query it.GetActionByIdQuery) (*it.GetActionByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Action, it.GetActionByIdQuery, it.GetActionByIdResult]{
		Action:      "get action by Id",
		Query:       query,
		RepoFindOne: this.getActionByIdFull,
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.GetActionByIdResult {
			return &it.GetActionByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Action) *it.GetActionByIdResult {
			return &it.GetActionByIdResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ActionServiceImpl) SearchActions(ctx crud.Context, query it.SearchActionsQuery) (*it.SearchActionsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.Action, it.SearchActionsQuery, it.SearchActionsResult]{
		Action: "search actions",
		Query:  query,
		SetQueryDefaults: func(query *it.SearchActionsQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.actionRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query it.SearchActionsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Action], error) {
			return this.actionRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *fault.ValidationErrors) *it.SearchActionsResult {
			return &it.SearchActionsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.Action]) *it.SearchActionsResult {
			return &it.SearchActionsResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *ActionServiceImpl) sanitizeAction(action *domain.Action) {
	if action.Description != nil {
		action.Description = util.ToPtr(defense.SanitizePlainText(*action.Description, true))
	}

	if action.Name != nil {
		action.Name = util.ToPtr(defense.SanitizePlainText(*action.Name, true))
	}
}

func (this *ActionServiceImpl) setActionDefaults(action *domain.Action) {
	action.SetDefaults()
}

func (this *ActionServiceImpl) assertBusinessRuleCreateAction(ctx crud.Context, action *domain.Action, vErrs *fault.ValidationErrors) error {
	err := this.assertActionUnique(ctx, action, vErrs)
	fault.PanicOnErr(err)

	err = this.assertResourceExists(ctx, *action.ResourceId, vErrs)
	fault.PanicOnErr(err)

	return nil
}

func (this *ActionServiceImpl) assertActionUnique(ctx crud.Context, action *domain.Action, vErrs *fault.ValidationErrors) error {
	dbAction, err := this.actionRepo.FindByName(ctx, it.FindByNameParam{Name: *action.Name, ResourceId: *action.ResourceId})
	fault.PanicOnErr(err)

	if dbAction != nil {
		vErrs.AppendAlreadyExists("action_name", "action name")
	}
	return nil
}

func (this *ActionServiceImpl) getActionByIdFull(ctx crud.Context, query it.GetActionByIdQuery, vErrs *fault.ValidationErrors) (dbAction *domain.Action, err error) {
	dbAction, err = this.actionRepo.FindById(ctx, query)
	fault.PanicOnErr(err)

	if dbAction == nil {
		vErrs.AppendNotFound("action_id", "action")
	}
	return
}

func (this *ActionServiceImpl) assertActionExistsById(ctx crud.Context, action *domain.Action, vErrs *fault.ValidationErrors) (dbAction *domain.Action, err error) {
	dbAction, err = this.actionRepo.FindById(ctx, it.FindByIdParam{Id: *action.Id})
	fault.PanicOnErr(err)

	if dbAction == nil {
		vErrs.AppendNotFound("action_id", "action")
	}
	return
}

func (this *ActionServiceImpl) assertResourceExists(ctx crud.Context, id model.Id, vErrs *fault.ValidationErrors) (err error) {
	exist, err := this.resourceRepo.Exist(ctx, itResource.ExistParam{Id: id})
	fault.PanicOnErr(err)

	if !exist {
		vErrs.AppendNotFound("resource_id", "resource")
	}
	return err
}

func (this *ActionServiceImpl) assertConstraintViolated(action *domain.Action, vErrs *fault.ValidationErrors) error {
	if len(action.Entitlements) > 0 {
		for _, entitlement := range action.Entitlements {
			vErrs.AppendConstraintViolated("entitlements", *entitlement.Name)
		}
	}

	return nil
}

func (this *ActionServiceImpl) assertBusinessRuleDeleteAction(ctx crud.Context, command it.DeleteActionHardByIdCommand, action *domain.Action, vErrs *fault.ValidationErrors) error {
	err := this.assertConstraintViolated(action, vErrs)
	return err
}
