package services

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
	itAct "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/resource"
)

func NewActionDomainService(resourceSvc itRes.ResourceDomainService) itAct.ActionDomainService {
	return resourceSvc.(itAct.ActionDomainService)
}

func (this *ResourceDomainServiceImpl) CreateAction(
	ctx corectx.Context, cmd itAct.CreateActionCommand,
) (*itAct.CreateActionResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Action, *domain.Action]{
		Action:         "create action",
		BaseRepoGetter: this.actionRepo,
		Data:           cmd,
		BeforeValidation: func(ctx corectx.Context, inputModel *domain.Action, vErrs *ft.ClientErrors) (*domain.Action, error) {
			err := this.checkResourceExists(ctx, inputModel.MustGetResourceId(), vErrs)
			if err != nil {
				return nil, err
			}
			return inputModel, nil
		},
	})
}

func (this *ResourceDomainServiceImpl) DeleteAction(
	ctx corectx.Context, cmd itAct.DeleteActionCommand,
) (_ *itAct.DeleteActionResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "delete action"); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := cmd.GetSchema().ValidateStruct(cmd)
	if cErrs.Count() > 0 {
		return &itAct.DeleteActionResult{ClientErrors: cErrs}, nil
	}

	cmd = *(sanitized.(*itAct.DeleteActionCommand))

	err = this.checkResourceExists(ctx, cmd.ResourceId, &cErrs)
	if err != nil {
		return
	}
	if cErrs.Count() > 0 {
		return &itAct.DeleteActionResult{ClientErrors: cErrs}, nil
	}

	action := domain.NewAction()
	action.SetId(util.ToPtr(model.Id(cmd.ActionId)))
	action.SetResourceId(&cmd.ResourceId)
	delResult, err := this.actionRepo.DeleteOne(ctx, *action)

	return delResult, err
}

func (this *ResourceDomainServiceImpl) ActionExists(
	ctx corectx.Context, query itAct.ActionExistsQuery,
) (_ *itAct.ActionExistsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "check if action exists"); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itAct.ActionExistsResult{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itAct.ActionExistsQuery))

	err = this.checkResourceExists(ctx, query.ResourceId, &cErrs)
	if err != nil {
		return
	}
	if cErrs.Count() > 0 {
		return &itAct.ActionExistsResult{ClientErrors: cErrs}, nil
	}

	keys := this.existsQueryKeys(query)
	repoOut, err := this.actionRepo.Exists(ctx, keys)
	if err != nil {
		return nil, err
	}
	if len(repoOut.ClientErrors) > 0 {
		return &itAct.ActionExistsResult{ClientErrors: repoOut.ClientErrors}, nil
	}
	data := this.existsResultData(repoOut.Data)
	return &itAct.ActionExistsResult{Data: data, HasData: true}, nil
}

func (this *ResourceDomainServiceImpl) existsQueryKeys(query itAct.ActionExistsQuery) []domain.Action {
	return array.Map(query.ActionIds, func(id model.Id) domain.Action {
		action := domain.NewAction()
		action.SetId(&id)
		action.SetResourceId(&query.ResourceId)
		return *action
	})
}

func (this *ResourceDomainServiceImpl) existsResultData(repo dyn.RepoExistsResult) dyn.ExistsResultData {
	data := dyn.ExistsResultData{
		Existing:    make([]string, 0),
		NotExisting: make([]string, 0),
	}
	data.Existing = array.Map(repo.Existing, func(fields dmodel.DynamicFields) model.Id {
		return *fields.GetModelId(domain.ActionFieldId)
	})
	data.NotExisting = array.Map(repo.NotExisting, func(fields dmodel.DynamicFields) model.Id {
		return *fields.GetModelId(domain.ActionFieldId)
	})
	return data
}

func (this *ResourceDomainServiceImpl) GetAction(
	ctx corectx.Context, query itAct.GetActionQuery,
) (_ *dyn.OpResult[domain.Action], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "get action"); e != nil {
			err = e
		}
	}()
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[domain.Action]{ClientErrors: cErrs}, nil
	}
	query = *(sanitized.(*itAct.GetActionQuery))

	err = this.checkResourceExists(ctx, query.ResourceId, &cErrs)
	if err != nil {
		return
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[domain.Action]{ClientErrors: cErrs}, nil
	}

	// Use search graph to leverage the index defined in action_entity.go, `resource_id` comes before `action_code`
	graph := dmodel.NewSearchGraph()
	graph.And(
		*dmodel.NewSearchNode().Condition(dmodel.NewCondition(domain.ActionFieldResourceId, dmodel.Equals, query.ResourceId)),
		*dmodel.NewSearchNode().Condition(dmodel.NewCondition(domain.ActionFieldId, dmodel.Equals, query.ActionId)),
	)
	resSearch, err := this.actionRepo.Search(ctx, dyn.RepoSearchParam{
		Graph:  graph,
		Fields: query.Columns,
		Page:   0,
		Size:   1,
	})
	if err != nil {
		return nil, err
	}
	if resSearch.ClientErrors.Count() > 0 {
		return &dyn.OpResult[domain.Action]{ClientErrors: resSearch.ClientErrors}, nil
	}

	var result dyn.OpResult[domain.Action]
	result.HasData = resSearch.HasData
	if resSearch.HasData {
		result.Data = resSearch.Data.Items[0]
	}
	return &result, err
}

func (this *ResourceDomainServiceImpl) SearchActions(
	ctx corectx.Context, query itAct.SearchActionsQuery,
) (result *itAct.SearchActionsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "search actions"); e != nil {
			err = e
		}
	}()
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itAct.SearchActionsResult{ClientErrors: cErrs}, nil
	}

	query = *(sanitized.(*itAct.SearchActionsQuery))
	err = this.checkResourceExists(ctx, query.ResourceId, &cErrs)
	if err != nil {
		return
	}
	if cErrs.Count() > 0 {
		return &itAct.SearchActionsResult{ClientErrors: cErrs}, nil
	}

	cond := dmodel.NewCondition(domain.ActionFieldResourceId, dmodel.Equals, query.ResourceId)
	graph := dmodel.NewSearchGraph()
	if query.Graph != nil {
		// Add resource_id filter to existing action search graph
		node := query.Graph.ToSearchNode()
		graph.And(
			*dmodel.NewSearchNode().Condition(cond),
			*node,
		)
	} else {
		graph.Condition(cond)
	}
	return corecrud.Search[domain.Action](ctx, corecrud.SearchParam{
		Action:       "search actions",
		DbRepoGetter: this.actionRepo,
		Query: dyn.SearchQuery{
			Fields: query.Columns,
			Graph:  graph,
			Page:   query.Page,
			Size:   query.Size,
		},
	})
}

func (this *ResourceDomainServiceImpl) UpdateAction(
	ctx corectx.Context, cmd itAct.UpdateActionCommand,
) (*itAct.UpdateActionResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Action, *domain.Action]{
		Action:       "update action",
		DbRepoGetter: this.actionRepo,
		Data:         cmd,
		BeforeValidation: func(ctx corectx.Context, inputModel *domain.Action, vErrs *ft.ClientErrors) (*domain.Action, error) {
			err := this.checkResourceExists(ctx, inputModel.MustGetResourceId(), vErrs)
			if err != nil {
				return nil, err
			}
			return inputModel, nil
		},
	})
}

// checkResourceExists encapsulates the resource existence check.
func (this *ResourceDomainServiceImpl) checkResourceExists(ctx corectx.Context, resourceId model.Id, cErrs *ft.ClientErrors) error {
	resource := models.NewResource()
	resource.SetId(&resourceId)
	resrcExists, err := this.resourceRepo.Exists(ctx, []domain.Resource{
		*resource,
	})
	if err != nil {
		return errors.Wrap(err, "check resource exists")
	}
	if resrcExists.ClientErrors.Count() > 0 {
		return errors.Wrap(resrcExists.ClientErrors.ToError(), "check resource exists")
	}
	existing := array.ContainsIf(resrcExists.Data.Existing, func(item dmodel.DynamicFields) bool {
		return *item.GetModelId(domain.ResourceFieldId) == resourceId
	})
	if !existing {
		cErrs.Append(*ft.NewNotFoundError(domain.ActionFieldResourceId))
		return nil
	}
	return nil
}
