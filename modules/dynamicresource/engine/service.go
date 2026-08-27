package engine

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// NewServiceParam carries what a resource service needs to work.
type NewServiceParam struct {
	Schema     *dmodel.ModelSchema
	Repository it.DynamicResourceRepository

	// DefaultFields is returned by a search that specifies neither fields nor a resolvable
	// view. When empty, every column of the schema is returned.
	DefaultFields []string

	// CrudActions is the engine's action allow-list. Empty means every action is supported.
	CrudActions []it.CrudAction

	// ActionLookup resolves an action definition at call time, which is how the service
	// reaches the validator hooks a module attached with ModifyAction. The service is built
	// before the engine that will own it exists, so the lookup has to be deferred - the same
	// reason invokeComputedFunction defers its own.
	ActionLookup func(actionName string) (it.DynamicActionDefinition, bool)
}

func NewDynamicResourceService(param NewServiceParam) it.DynamicResourceService {
	defaultFields := param.DefaultFields
	if len(defaultFields) == 0 {
		defaultFields = columnNames(param.Schema)
	}
	crudActions := make(map[it.CrudAction]bool, len(param.CrudActions))
	for _, action := range param.CrudActions {
		crudActions[action] = true
	}
	return &DynamicResourceServiceImpl{
		schema:        param.Schema,
		repository:    param.Repository,
		defaultFields: defaultFields,
		crudActions:   crudActions,
		actionLookup:  param.ActionLookup,
	}
}

// DynamicResourceServiceImpl is the schema-agnostic merge of what a feature module
// splits between its application service and its domain service, minus the permission
// checks, which the engine pipeline performs before calling in here.
//
// A module extends it by embedding it in its own service struct and installing that
// struct with Engine.SetResourceService.
type DynamicResourceServiceImpl struct {
	schema        *dmodel.ModelSchema
	repository    it.DynamicResourceRepository
	defaultFields []string

	// crudActions is the engine's action allow-list, empty meaning "all supported".
	crudActions map[it.CrudAction]bool

	// actionLookup reads an action definition, and with it the validator hooks that modules
	// attached via ModifyAction. Nil in a service built outside the registry, which is why
	// every read of it is nil-guarded.
	actionLookup func(actionName string) (it.DynamicActionDefinition, bool)
}

// declaredHooks returns the validator hooks of the named action, or the zero definition when
// there is no lookup or no such action. The definition is the only place hooks live: there is
// no per-call way to supply one, so what runs is answerable by reading the definition alone.
func (this *DynamicResourceServiceImpl) declaredHooks(actionName string) it.DynamicActionDefinition {
	if this.actionLookup == nil {
		return it.DynamicActionDefinition{}
	}
	declared, _ := this.actionLookup(actionName)
	return declared
}

func (this *DynamicResourceServiceImpl) Schema() *dmodel.ModelSchema {
	return this.schema
}

// Repository exposes the repository to embedding services.
func (this *DynamicResourceServiceImpl) Repository() it.DynamicResourceRepository {
	return this.repository
}

func (this *DynamicResourceServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	if cErrs := this.assertActionSupported(it.CrudActionCreate); cErrs != nil {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *cErrs}, nil
	}
	declared := this.declaredHooks(it.ActionCreate)

	// corecrud's create hook takes no foundModel: a create has no stored record to compare
	// against, so the action-level foundModel is always nil here. The variable stays nil when
	// there is no hook, because corecrud skips the step on nil and an always-non-nil closure
	// would turn a no-op into a SetFieldData round trip.
	var validateExtra corecrud.CreateValidateExtraFn[*it.DynamicEntity]
	if declared.ValidateExtra != nil {
		hook := declared.ValidateExtra
		validateExtra = func(ctx corectx.Context, inputModel *it.DynamicEntity, vErrs *ft.ClientErrors) error {
			return hook(ctx, inputModel, nil, vErrs)
		}
	}

	result, err := corecrud.Create[it.DynamicEntity](ctx, corecrud.CreateParam[it.DynamicEntity, *it.DynamicEntity]{
		Action:                 this.actionName("create"),
		BaseRepoGetter:         this.repository,
		Data:                   it.NewDynamicEntityFrom(params),
		BeforeValidation:       declared.BeforeValidation,
		AfterValidationSuccess: declared.AfterValidationSuccess,
		ValidateExtra:          validateExtra,
	})
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.Create")
	}
	return unwrapEntityResult(result), nil
}

func (this *DynamicResourceServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if cErrs := this.assertActionSupported(it.CrudActionUpdate); cErrs != nil {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *cErrs}, nil
	}
	declared := this.declaredHooks(it.ActionUpdate)

	// No shim on this side: UpdateParam.ValidateExtra *is* ActionValidateExtraFn, and corecrud
	// fetches the stored record itself to fill foundModel.
	result, err := corecrud.Update[it.DynamicEntity](ctx, corecrud.UpdateParam[it.DynamicEntity, *it.DynamicEntity]{
		Action:                 this.actionName("update"),
		DbRepoGetter:           this.repository,
		Data:                   it.NewDynamicEntityFrom(params),
		BeforeValidation:       declared.BeforeValidation,
		AfterValidationSuccess: declared.AfterValidationSuccess,
		ValidateExtra:          declared.ValidateExtra,
	})
	return result, errors.Wrap(err, "DynamicResourceService.Update")
}

func (this *DynamicResourceServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if cErrs := this.assertActionSupported(it.CrudActionDelete); cErrs != nil {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *cErrs}, nil
	}
	declared := this.declaredHooks(it.ActionDelete)

	// corecrud's delete hook sees only the key fields - it never reads the row it is about to
	// remove - but a delete guard's whole job is to inspect the stored record ("only a
	// cancelled order may be deleted"). So the service fetches it here and hands it over as
	// foundModel, which is the contract the entity-typed hook declares.
	//
	// A guard that cannot see the record would fail open or, worse, fail closed on every
	// delete, so a fetch that errors aborts rather than validating against a blank.
	var validateExtra corecrud.DeleteValidateExtraFn
	if declared.ValidateExtra != nil {
		hook := declared.ValidateExtra
		validateExtra = func(ctx corectx.Context, keyFields dmodel.DynamicFields, vErrs *ft.ClientErrors) error {
			foundModel, err := this.findByKeys(ctx, keyFields, vErrs)
			if err != nil {
				return errors.Wrap(err, "Delete.ValidateExtra")
			}
			if vErrs.Count() > 0 {
				return nil
			}
			return hook(ctx, it.NewDynamicEntityFrom(keyFields), foundModel, vErrs)
		}
	}

	result, err := corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:        this.actionName("delete"),
		DbRepoGetter:  this.repository,
		Cmd:           paramsToDeleteCommand(params),
		ValidateExtra: validateExtra,
	})
	return result, errors.Wrap(err, "DynamicResourceService.Delete")
}

func (this *DynamicResourceServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if cErrs := this.assertActionSupported(it.CrudActionSetArchived); cErrs != nil {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *cErrs}, nil
	}
	// corecrud.SetIsArchived takes its arguments positionally and accepts no hooks, so a
	// set_archived ValidateExtra has nowhere to run. Enforced in DefineAction rather than
	// dropped silently - see assertNoUnsupportedHooks.
	result, err := corecrud.SetIsArchived(ctx, this.repository, paramsToSetArchivedCommand(params))
	return result, errors.Wrap(err, "DynamicResourceService.SetArchived")
}

func (this *DynamicResourceServiceImpl) GetById(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	if cErrs := this.assertActionSupported(it.CrudActionGetById); cErrs != nil {
		return &dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]{ClientErrors: *cErrs}, nil
	}
	query, err := paramsToGetOneQuery(params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.GetById")
	}
	// defaultFields is the *list view* projection a search falls back to (e.g. a user is
	// listed by avatar/name/email/status). Fetching one record by its id is the detail
	// view, so with no explicit selection it returns the whole record instead.
	if len(query.Fields) == 0 {
		query.Fields = columnNames(this.schema)
	}

	result, err := corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[it.DynamicEntity, *it.DynamicEntity]{
		Action: this.actionName("get by id"),
		Schema: this.schema,
		GetOneFn: func() (*dyn.OpResult[it.DynamicEntity], error) {
			return corecrud.GetOne[it.DynamicEntity](ctx, corecrud.GetOneParam{
				Action:       this.actionName("get by id"),
				DbRepoGetter: this.repository,
				Query:        query,
			})
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.GetById")
	}
	return unwrapSingleResult(result), nil
}

// GetOne fetches one record by any unique key present in params. Every param that names
// a schema column is used as an equality condition, so callers must pass only key fields.
func (this *DynamicResourceServiceImpl) GetOne(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	if cErrs := this.assertActionSupported(it.CrudActionGetByUnique); cErrs != nil {
		return &dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]{ClientErrors: *cErrs}, nil
	}
	fields, err := this.readDesiredFields(params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.GetOne")
	}

	graph, cErrs := this.uniqueKeysToGraph(params)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]{ClientErrors: cErrs}, nil
	}

	result, err := corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[it.DynamicEntity, *it.DynamicEntity]{
		Action: this.actionName("get one"),
		Schema: this.schema,
		GetOneFn: func() (*dyn.OpResult[it.DynamicEntity], error) {
			return this.searchSingle(ctx, fields, graph)
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.GetOne")
	}
	return unwrapSingleResult(result), nil
}

func (this *DynamicResourceServiceImpl) Search(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	if cErrs := this.assertActionSupported(it.CrudActionSearch); cErrs != nil {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{ClientErrors: *cErrs}, nil
	}
	query, err := paramsToSearchQuery(params)
	if err != nil {
		// A query parameter of the wrong type is the caller's mistake; reporting it as a 500
		// would blame the server for a bad request.
		if cErrs, ok := clientErrorsForDecodeFailure(err); ok {
			return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{ClientErrors: cErrs}, nil
		}
		return nil, errors.Wrap(err, "DynamicResourceService.Search")
	}

	result, err := corecrud.UiSearch(ctx, corecrud.UiSearchParam[it.DynamicEntity, *it.DynamicEntity]{
		Action:        this.actionName("search"),
		DefaultFields: this.defaultFields,
		Schema:        this.schema,
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[it.DynamicEntity]], error) {
			return corecrud.Search[it.DynamicEntity](ctx, corecrud.SearchParam{
				Action:                 this.actionName("search"),
				DbRepoGetter:           this.repository,
				Query:                  query,
				AfterValidationSuccess: fn,
			})
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.Search")
	}
	return unwrapPagedResult(result), nil
}

func (this *DynamicResourceServiceImpl) Exists(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.ExistsResultData], error) {
	if cErrs := this.assertActionSupported(it.CrudActionExists); cErrs != nil {
		return &dyn.OpResult[dyn.ExistsResultData]{ClientErrors: *cErrs}, nil
	}
	query, err := paramsToExistsQuery(params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.Exists")
	}
	result, err := corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       this.actionName("exists"),
		DbRepoGetter: this.repository,
		Query:        query,
	})
	return result, errors.Wrap(err, "DynamicResourceService.Exists")
}

// searchSingle runs a one-item search over the given graph and reshapes it into a get-one result.
func (this *DynamicResourceServiceImpl) searchSingle(
	ctx corectx.Context, fields []string, graph *dmodel.SearchGraph,
) (*dyn.OpResult[it.DynamicEntity], error) {
	searchRes, err := corecrud.Search[it.DynamicEntity](ctx, corecrud.SearchParam{
		Action:       this.actionName("get one"),
		DbRepoGetter: this.repository,
		Query: dyn.SearchQuery{
			Fields: fields,
			Graph:  graph,
			Page:   0,
			Size:   1,
		},
	})
	if err != nil {
		return nil, err
	}

	result := &dyn.OpResult[it.DynamicEntity]{
		ClientErrors: searchRes.ClientErrors,
		HasData:      searchRes.HasData,
	}
	if searchRes.HasData {
		result.Data = searchRes.Data.Items[0]
	}
	return result, nil
}

func (this *DynamicResourceServiceImpl) actionName(verb string) string {
	return verb + " " + this.schema.Name()
}
