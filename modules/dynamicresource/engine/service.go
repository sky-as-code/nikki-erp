package engine

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
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

	// ActionLookup resolves the action whose hooks an operation runs. Pass the engine's Action
	// method. Leaving it nil runs no hook, which suits a test exercising the plain CRUD.
	ActionLookup ActionLookupFn
}

func NewDynamicResourceService(param NewServiceParam) it.DynamicResourceService {
	defaultFields := param.DefaultFields
	if len(defaultFields) == 0 {
		defaultFields = columnNames(param.Schema)
	}
	return &DynamicResourceServiceImpl{
		schema:        param.Schema,
		repository:    param.Repository,
		defaultFields: defaultFields,
		actionLookup:  param.ActionLookup,
	}
}

// DynamicResourceServiceImpl is the schema-agnostic merge of what a feature module
// splits between its application service and its domain service, minus the permission
// checks, which the engine pipeline performs before calling in here.
//
// Each operation runs the hooks its action declares, resolved through actionLookup by the
// action's name. The name is fixed by the method, never supplied by the caller, so a module
// reaching the service without going through an action gets the same hooks a request does.
//
// A module extends it by embedding it in its own service struct and installing that
// struct with Engine.SetResourceService.
type DynamicResourceServiceImpl struct {
	schema        *dmodel.ModelSchema
	repository    it.DynamicResourceRepository
	defaultFields []string
	actionLookup  ActionLookupFn
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
	definition := this.action(it.ActionCreate)

	result, err := corecrud.Create[it.DynamicEntity](ctx, corecrud.CreateParam[it.DynamicEntity, *it.DynamicEntity]{
		Action:                 this.actionName("create"),
		BaseRepoGetter:         this.repository,
		Data:                   it.NewDynamicEntityFrom(params),
		BeforeValidation:       beforeValidationFn(definition.BeforeValidation),
		ValidateExtra:          createValidateExtraFn(definition.ValidateExtra),
		AfterValidationSuccess: afterValidationFn(definition.AfterValidationSuccess),
	})
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.Create")
	}
	return unwrapEntityResult(result), nil
}

// Update hands its hooks to the crud helper rather than running them here, which is also what
// gives an update guard a real foundModel: the helper reads the stored row to check the etag,
// and passes that same row on. Any KeysToFetch the action declares is therefore redundant on
// update, and is ignored.
func (this *DynamicResourceServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	definition := this.action(it.ActionUpdate)

	result, err := corecrud.Update[it.DynamicEntity](ctx, corecrud.UpdateParam[it.DynamicEntity, *it.DynamicEntity]{
		Action:                 this.actionName("update"),
		DbRepoGetter:           this.repository,
		Data:                   it.NewDynamicEntityFrom(params),
		BeforeValidation:       beforeValidationFn(definition.BeforeValidation),
		ValidateExtra:          updateValidateExtraFn(definition.ValidateExtra),
		AfterValidationSuccess: afterValidationFn(definition.AfterValidationSuccess),
	})
	return result, errors.Wrap(err, "DynamicResourceService.Update")
}

// Delete resolves the record its guard validates against before calling the crud helper: the
// helper's own hook is handed nothing but the key fields, and a delete guard almost always needs
// the row's state to decide whether the delete is allowed at all.
func (this *DynamicResourceServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	definition := this.action(it.ActionDelete)

	params, foundModel, cErrs, err := this.prepare(ctx, definition, params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.Delete")
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
	}

	result, err := corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:                 this.actionName("delete"),
		DbRepoGetter:           this.repository,
		Cmd:                    paramsToDeleteCommand(params),
		ValidateExtra:          deleteValidateExtraFn(definition.ValidateExtra, params, foundModel),
		AfterValidationSuccess: deleteAfterValidationFn(definition.AfterValidationSuccess, params),
	})
	return result, errors.Wrap(err, "DynamicResourceService.Delete")
}

// SetArchived runs its hooks here rather than passing them down: crud.SetIsArchived validates a
// typed command instead of the resource schema, so it has no hook slot to hand them to.
func (this *DynamicResourceServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	params, cErrs, err := this.runHooks(ctx, it.ActionSetArchived, params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.SetArchived")
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
	}

	result, err := corecrud.SetIsArchived(ctx, this.repository, paramsToSetArchivedCommand(params))
	return result, errors.Wrap(err, "DynamicResourceService.SetArchived")
}

func (this *DynamicResourceServiceImpl) GetById(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	params, cErrs, err := this.runHooks(ctx, it.ActionGetById, params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.GetById")
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]{ClientErrors: cErrs}, nil
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
	params, cErrs, err := this.runHooks(ctx, it.ActionGetByUnique, params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.GetOne")
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]{ClientErrors: cErrs}, nil
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
	params, cErrs, err := this.runHooks(ctx, it.ActionSearch, params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.Search")
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{ClientErrors: cErrs}, nil
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
	params, cErrs, err := this.runHooks(ctx, it.ActionExists, params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.Exists")
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.ExistsResultData]{ClientErrors: cErrs}, nil
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
