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

	// FieldResolver resolves the field list of a named search view. It is optional:
	// when absent, a search that names a view falls back to DefaultFields.
	FieldResolver corecrud.FieldsResolver

	// DefaultFields is returned by a search that specifies neither fields nor a resolvable
	// view. When empty, every column of the schema is returned.
	DefaultFields []string
}

func NewDynamicResourceService(param NewServiceParam) it.DynamicResourceService {
	defaultFields := param.DefaultFields
	if len(defaultFields) == 0 {
		defaultFields = columnNames(param.Schema)
	}
	return &DynamicResourceServiceImpl{
		schema:        param.Schema,
		repository:    param.Repository,
		fieldResolver: param.FieldResolver,
		defaultFields: defaultFields,
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
	fieldResolver corecrud.FieldsResolver
	defaultFields []string
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
	result, err := corecrud.Create[it.DynamicEntity](ctx, corecrud.CreateParam[it.DynamicEntity, *it.DynamicEntity]{
		Action:         this.actionName("create"),
		BaseRepoGetter: this.repository,
		Data:           it.NewDynamicEntityFrom(params),
	})
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.Create")
	}
	return unwrapEntityResult(result), nil
}

func (this *DynamicResourceServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	result, err := corecrud.Update[it.DynamicEntity](ctx, corecrud.UpdateParam[it.DynamicEntity, *it.DynamicEntity]{
		Action:       this.actionName("update"),
		DbRepoGetter: this.repository,
		Data:         it.NewDynamicEntityFrom(params),
	})
	return result, errors.Wrap(err, "DynamicResourceService.Update")
}

func (this *DynamicResourceServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	result, err := corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       this.actionName("delete"),
		DbRepoGetter: this.repository,
		Cmd:          paramsToDeleteCommand(params),
	})
	return result, errors.Wrap(err, "DynamicResourceService.Delete")
}

func (this *DynamicResourceServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	result, err := corecrud.SetIsArchived(ctx, this.repository, paramsToSetArchivedCommand(params))
	return result, errors.Wrap(err, "DynamicResourceService.SetArchived")
}

func (this *DynamicResourceServiceImpl) GetById(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
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
	query, err := paramsToSearchQuery(params)
	if err != nil {
		return nil, errors.Wrap(err, "DynamicResourceService.Search")
	}

	result, err := corecrud.UiSearch(ctx, corecrud.UiSearchParam[it.DynamicEntity, *it.DynamicEntity]{
		Action:        this.actionName("search"),
		DefaultFields: this.defaultFields,
		FieldResolver: this.fieldResolver,
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
