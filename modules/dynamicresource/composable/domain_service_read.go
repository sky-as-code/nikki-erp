package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

func (this *DefaultDomainServiceImpl) GetById(
	ctx corectx.Context, params GetByIdQuery,
) (*GetOneResult, error) {
	if cErrs := this.assertActionSupported(CrudActionGetById); cErrs != nil {
		return &GetOneResult{ClientErrors: *cErrs}, nil
	}
	query, err := paramsToGetOneQuery(params)
	if err != nil {
		return nil, errors.Wrap(err, "CrudDomainService.GetById")
	}
	// defaultFields is the *list view* projection a search falls back to (e.g. a user is listed
	// by avatar/name/email/status). Fetching one record by its id is the detail view, so with no
	// explicit selection it returns the whole record instead.
	if len(query.Fields) == 0 {
		query.Fields = columnNames(this.schema)
	}

	result, err := corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[DynamicEntity, *DynamicEntity]{
		Action: this.actionName("get by id"),
		Schema: this.schema,
		GetOneFn: func() (*dyn.OpResult[DynamicEntity], error) {
			return corecrud.GetOne[DynamicEntity](ctx, corecrud.GetOneParam{
				Action:       this.actionName("get by id"),
				DbRepoGetter: this.repository,
				Query:        query,
			})
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "CrudDomainService.GetById")
	}
	return unwrapSingleResult(result), nil
}

// GetOne fetches one record by any unique key present in the query. Every param that names a
// schema column is used as an equality condition, so callers must pass only key fields.
func (this *DefaultDomainServiceImpl) GetOne(
	ctx corectx.Context, params GetOneQuery,
) (*GetOneResult, error) {
	if cErrs := this.assertActionSupported(CrudActionGetByUnique); cErrs != nil {
		return &GetOneResult{ClientErrors: *cErrs}, nil
	}
	fields, err := this.readDesiredFields(params)
	if err != nil {
		return nil, errors.Wrap(err, "CrudDomainService.GetOne")
	}

	graph, cErrs := this.uniqueKeysToGraph(params)
	if cErrs.Count() > 0 {
		return &GetOneResult{ClientErrors: cErrs}, nil
	}

	result, err := corecrud.UiGetOne(ctx, corecrud.UiGetOneParam[DynamicEntity, *DynamicEntity]{
		Action: this.actionName("get one"),
		Schema: this.schema,
		GetOneFn: func() (*dyn.OpResult[DynamicEntity], error) {
			return this.searchSingle(ctx, fields, graph)
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "CrudDomainService.GetOne")
	}
	return unwrapSingleResult(result), nil
}

func (this *DefaultDomainServiceImpl) Search(
	ctx corectx.Context, params SearchQuery, options ...corecrud.ServiceSearchOptions,
) (*SearchResult, error) {
	if cErrs := this.assertActionSupported(CrudActionSearch); cErrs != nil {
		return &SearchResult{ClientErrors: *cErrs}, nil
	}
	opts := firstOrZero(options)
	query, err := paramsToSearchQuery(params)
	if err != nil {
		// A query parameter of the wrong type is the caller's mistake; reporting it as a 500
		// would blame the server for a bad request.
		if cErrs, ok := clientErrorsForDecodeFailure(err); ok {
			return &SearchResult{ClientErrors: cErrs}, nil
		}
		return nil, errors.Wrap(err, "CrudDomainService.Search")
	}

	result, err := corecrud.UiSearch(ctx, corecrud.UiSearchParam[DynamicEntity, *DynamicEntity]{
		Action:        this.actionName("search"),
		DefaultFields: this.defaultFields,
		Schema:        this.schema,
		SearchFn: func(fn corecrud.AfterValidationSuccessFn[dyn.SearchQuery]) (*dyn.OpResult[dyn.PagedResultData[DynamicEntity]], error) {
			return corecrud.Search[DynamicEntity](ctx, corecrud.SearchParam{
				Action:                 this.actionName("search"),
				DbRepoGetter:           this.repository,
				Query:                  query,
				AfterValidationSuccess: chainSearchHooks(fn, opts.AfterValidationSuccess),
			})
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "CrudDomainService.Search")
	}
	return unwrapPagedResult(result), nil
}

// chainSearchHooks runs the UI projection hook first, then the caller's, so a derived service
// can still narrow or rewrite the query after the default field resolution happened.
func chainSearchHooks(
	first corecrud.AfterValidationSuccessFn[dyn.SearchQuery], second corecrud.AfterValidationSuccessFn[dyn.SearchQuery],
) corecrud.AfterValidationSuccessFn[dyn.SearchQuery] {
	if second == nil {
		return first
	}
	if first == nil {
		return second
	}
	return func(ctx corectx.Context, query dyn.SearchQuery) (dyn.SearchQuery, error) {
		query, err := first(ctx, query)
		if err != nil {
			return query, err
		}
		return second(ctx, query)
	}
}

func (this *DefaultDomainServiceImpl) Exists(
	ctx corectx.Context, params ExistsQuery,
) (*ExistsResult, error) {
	if cErrs := this.assertActionSupported(CrudActionExists); cErrs != nil {
		return &ExistsResult{ClientErrors: *cErrs}, nil
	}
	query, err := paramsToExistsQuery(params)
	if err != nil {
		return nil, errors.Wrap(err, "CrudDomainService.Exists")
	}
	result, err := corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       this.actionName("exists"),
		DbRepoGetter: this.repository,
		Query:        query,
	})
	return result, errors.Wrap(err, "CrudDomainService.Exists")
}

// searchSingle runs a one-item search over the given graph and reshapes it into a get-one result.
func (this *DefaultDomainServiceImpl) searchSingle(
	ctx corectx.Context, fields []string, graph *dmodel.SearchGraph,
) (*dyn.OpResult[DynamicEntity], error) {
	searchRes, err := corecrud.Search[DynamicEntity](ctx, corecrud.SearchParam{
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

	result := &dyn.OpResult[DynamicEntity]{
		ClientErrors: searchRes.ClientErrors,
		HasData:      searchRes.HasData,
	}
	if searchRes.HasData {
		result.Data = searchRes.Data.Items[0]
	}
	return result, nil
}
