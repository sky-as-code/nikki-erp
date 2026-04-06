package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itAct "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
)

func NewActionServiceImpl(
	actionRepo itAct.ActionRepository,
	cqrsBus cqrs.CqrsBus,
) itAct.ActionService {
	return &ActionServiceImpl{cqrsBus: cqrsBus, actionRepo: actionRepo}
}

type ActionServiceImpl struct {
	cqrsBus    cqrs.CqrsBus
	actionRepo itAct.ActionRepository
}

func (this *ActionServiceImpl) CreateAction(
	ctx corectx.Context, cmd itAct.CreateActionCommand,
) (*itAct.CreateActionResult, error) {
	return corecrud.Create(ctx, dyn.CreateParam[domain.Action, *domain.Action]{
		Action:         "create action",
		BaseRepoGetter: this.actionRepo,
		Data:           cmd,
	})
}

func (this *ActionServiceImpl) DeleteAction(
	ctx corectx.Context, cmd itAct.DeleteActionCommand,
) (*itAct.DeleteActionResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete action",
		DbRepoGetter: this.actionRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ActionServiceImpl) ActionExists(
	ctx corectx.Context, query itAct.ActionExistsQuery,
) (*itAct.ActionExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if action exists",
		DbRepoGetter: this.actionRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *ActionServiceImpl) GetAction(
	ctx corectx.Context, query itAct.GetActionQuery,
) (*itAct.GetActionResult, error) {
	return corecrud.GetOne[domain.Action](ctx, corecrud.GetOneParam{
		Action:       "get action",
		DbRepoGetter: this.actionRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ActionServiceImpl) SearchActions(
	ctx corectx.Context, query itAct.SearchActionsQuery,
) (*itAct.SearchActionsResult, error) {
	return corecrud.Search[domain.Action](ctx, corecrud.SearchParam{
		Action:       "search actions",
		DbRepoGetter: this.actionRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *ActionServiceImpl) UpdateAction(
	ctx corectx.Context, cmd itAct.UpdateActionCommand,
) (*itAct.UpdateActionResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Action, *domain.Action]{
		Action:       "update action",
		DbRepoGetter: this.actionRepo,
		Data:         cmd,
	})
}
