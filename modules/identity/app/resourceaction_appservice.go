package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itAct "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

func NewActionApplicationService(resourceSvc itRes.ResourceAppService) itAct.ActionAppService {
	return resourceSvc.(itAct.ActionAppService)
}

func (this *ResourceApplicationServiceImpl) CreateAction(ctx corectx.Context, cmd itAct.CreateActionCommand) (*itAct.CreateActionResult, error) {
	if cErr := assertPermission(ctx, "manage_actions", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itAct.CreateActionResult{ClientErrors: *cErr}, nil
	}
	return this.actionSvc.CreateAction(ctx, cmd)
}

func (this *ResourceApplicationServiceImpl) DeleteAction(ctx corectx.Context, cmd itAct.DeleteActionCommand) (*itAct.DeleteActionResult, error) {
	if cErr := assertPermission(ctx, "manage_actions", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itAct.DeleteActionResult{ClientErrors: *cErr}, nil
	}
	return this.actionSvc.DeleteAction(ctx, cmd)
}

func (this *ResourceApplicationServiceImpl) ActionExists(ctx corectx.Context, query itAct.ActionExistsQuery) (*itAct.ActionExistsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itAct.ActionExistsResult{ClientErrors: *cErr}, nil
	}
	return this.actionSvc.ActionExists(ctx, query)
}

func (this *ResourceApplicationServiceImpl) GetAction(ctx corectx.Context, query itAct.GetActionQuery) (*itAct.GetActionResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itAct.GetActionResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[domain.Action, *domain.Action]{
		Action: "get action",
		Schema: this.actionRepo.GetBaseRepo().Schema(),
		GetOneFn: func() (*dyn.OpResult[domain.Action], error) {
			return this.actionSvc.GetAction(ctx, query)
		},
	})
}

func (this *ResourceApplicationServiceImpl) SearchActions(ctx corectx.Context, query itAct.SearchActionsQuery) (*itAct.SearchActionsResult, error) {
	if cErr := assertPermission(ctx, "read", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itAct.SearchActionsResult{ClientErrors: *cErr}, nil
	}
	return corecrud.UiSearch(ctx, corecrud.UiSearchParam[domain.Action, *domain.Action]{
		Action:            "search actions",
		FieldResolver:     this.userPrefSvc.(corecrud.FieldsResolver),
		Schema:            this.actionRepo.GetBaseRepo().Schema(),
		DefaultSearchName: "action_list",
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[domain.Action]], error) {
			return this.actionSvc.SearchActions(ctx, query)
		},
	})
}

func (this *ResourceApplicationServiceImpl) UpdateAction(ctx corectx.Context, cmd itAct.UpdateActionCommand) (*itAct.UpdateActionResult, error) {
	if cErr := assertPermission(ctx, "manage_actions", c.ResourceAuthorizationResource, c.ResourceScopeDomain); cErr != nil {
		return &itAct.UpdateActionResult{ClientErrors: *cErr}, nil
	}
	return this.actionSvc.UpdateAction(ctx, cmd)
}
