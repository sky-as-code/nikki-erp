package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// The strongly-typed read port other modules bind to.
//
// These wrap the same overrides the engine actions use, so a consumer calling
// SearchProductVariants gets the batched template_* fill for free rather than a second
// implementation of it.

var (
	_ itProduct.ProductVariantDomainService = (*ProductVariantDomainServiceImpl)(nil)
	_ itProduct.ProductTemplateReadService  = (*ProductVariantDomainServiceImpl)(nil)
	_ itProduct.ProductCategoryReadService  = (*ProductVariantDomainServiceImpl)(nil)
)

func (this *ProductVariantDomainServiceImpl) SearchProductVariants(
	ctx corectx.Context, query itProduct.SearchProductVariantsQuery,
) (*itProduct.SearchProductVariantsResult, error) {
	// Routed through this service's own Search so the template_* rewrite and the batched fill
	// apply, rather than going straight to the repository.
	result, err := this.Search(ctx, searchQueryToParams(query))
	if err != nil {
		return nil, errors.Wrap(err, "SearchProductVariants")
	}
	if result.ClientErrors.Count() > 0 {
		return &itProduct.SearchProductVariantsResult{ClientErrors: result.ClientErrors}, nil
	}
	if !result.HasData {
		return emptyPage[models.ProductVariant](query), nil
	}
	return pagedResult(&result.Data, query, models.NewProductVariantFrom), nil
}

func (this *ProductVariantDomainServiceImpl) GetProductVariant(
	ctx corectx.Context, query itProduct.GetProductVariantQuery,
) (*itProduct.GetProductVariantResult, error) {
	if query.Id == "" {
		return &itProduct.GetProductVariantResult{}, nil
	}

	params := dmodel.DynamicFields{models.ProductVariantFieldId: string(query.Id)}
	if len(query.Fields) > 0 {
		params[paramFieldNames] = query.Fields
	}

	result, err := this.GetById(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, "GetProductVariant")
	}
	if result.ClientErrors.Count() > 0 {
		return &itProduct.GetProductVariantResult{ClientErrors: result.ClientErrors}, nil
	}
	if !result.HasData {
		return &itProduct.GetProductVariantResult{}, nil
	}

	return &itProduct.GetProductVariantResult{
		Data:    *models.NewProductVariantFrom(result.Data.Item),
		HasData: true,
	}, nil
}

func (this *ProductVariantDomainServiceImpl) ProductVariantsExist(
	ctx corectx.Context, query itProduct.ProductVariantsExistQuery,
) (*itProduct.ProductVariantsExistResult, error) {
	if len(query.Ids) == 0 {
		return &itProduct.ProductVariantsExistResult{HasData: true}, nil
	}

	engine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return nil, err
	}

	keys := make([]dmodel.DynamicFields, 0, len(query.Ids))
	for _, id := range query.Ids {
		keys = append(keys, dmodel.DynamicFields{models.ProductVariantFieldId: string(id)})
	}

	found, err := engine.ResourceRepository().Exists(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "ProductVariantsExist")
	}
	if found.ClientErrors.Count() > 0 {
		return &itProduct.ProductVariantsExistResult{ClientErrors: found.ClientErrors}, nil
	}

	return &itProduct.ProductVariantsExistResult{
		Data:    existsResultData(query.Ids, found.Data),
		HasData: true,
	}, nil
}

func (this *ProductVariantDomainServiceImpl) SearchTemplates(
	ctx corectx.Context, query itProduct.SearchTemplatesQuery,
) (*itProduct.SearchTemplatesResult, error) {
	rows, err := searchRows(ctx, models.ProductTemplateSchemaName, query, "SearchTemplates")
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return emptyPage[models.ProductTemplate](query), nil
	}
	return pagedResult(rows, query, models.NewProductTemplateFrom), nil
}

func (this *ProductVariantDomainServiceImpl) SearchCategories(
	ctx corectx.Context, query itProduct.SearchCategoriesQuery,
) (*itProduct.SearchCategoriesResult, error) {
	rows, err := searchRows(ctx, models.ProductCategorySchemaName, query, "SearchCategories")
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return emptyPage[models.ProductCategory](query), nil
	}
	return pagedResult(rows, query, models.NewProductCategoryFrom), nil
}

// searchQueryToParams converts the typed query into the params the resource service expects, so
// the typed port and the REST surface run through exactly the same code.
func searchQueryToParams(query dyn.SearchQuery) dmodel.DynamicFields {
	params := dmodel.DynamicFields{
		"page": query.Page,
		"size": query.Size,
	}
	if len(query.Fields) > 0 {
		params[paramFieldNames] = query.Fields
	}
	if query.Graph != nil {
		params["graph"] = query.Graph
	}
	if query.Language != nil {
		params["language"] = query.Language
	}
	return params
}

// searchRows runs one page of a search against the named resource's repository.
func searchRows(
	ctx corectx.Context, schemaName string, query dyn.SearchQuery, what string,
) (*dyn.PagedResultData[dmodel.DynamicFields], error) {
	engine, err := engineFor(schemaName)
	if err != nil {
		return nil, err
	}

	// This path goes straight to the repository rather than through crud.Search, so it has to
	// default the locale itself -- otherwise a variant list would sort its LangJson columns by the
	// raw jsonb while every other list in the product sorted by the reader's language.
	language := query.Language
	if language == nil {
		language = dyn.ResolveLocale(ctx)
	}

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Fields:   query.Fields,
		Page:     query.Page,
		Size:     query.Size,
		Graph:    query.Graph,
		Language: language,
	})
	if err != nil {
		return nil, errors.Wrap(err, what)
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), what)
	}
	if !found.HasData {
		return nil, nil
	}
	return &found.Data, nil
}

// emptyPage is still a page: the caller reads Total and Items either way, so this is HasData
// rather than a missing result.
func emptyPage[TModel any](query dyn.SearchQuery) *dyn.OpResult[dyn.PagedResultData[TModel]] {
	return &dyn.OpResult[dyn.PagedResultData[TModel]]{
		Data: dyn.PagedResultData[TModel]{
			Items: []TModel{},
			Page:  query.Page,
			Size:  query.Size,
		},
		HasData: true,
	}
}

// pagedResult wraps the rows of one page in the typed model the port declares.
func pagedResult[TModel any](
	rows *dyn.PagedResultData[dmodel.DynamicFields],
	query dyn.SearchQuery,
	wrap func(dmodel.DynamicFields) *TModel,
) *dyn.OpResult[dyn.PagedResultData[TModel]] {
	if rows == nil {
		return emptyPage[TModel](query)
	}

	items := make([]TModel, 0, len(rows.Items))
	for _, row := range rows.Items {
		items = append(items, *wrap(row))
	}
	return &dyn.OpResult[dyn.PagedResultData[TModel]]{
		Data: dyn.PagedResultData[TModel]{
			Items: items,
			Total: rows.Total,
			Page:  rows.Page,
			Size:  rows.Size,
		},
		HasData: true,
	}
}

// existsResultData splits the requested ids into those that exist and those that do not.
//
// The repository answers with the key maps it matched, not with ids, so the ids are read back out
// of them. Rebuilding the split from `requested` rather than trusting the order keeps the result
// aligned with what the caller asked about.
func existsResultData(requested []model.Id, found dyn.RepoExistsResult) dyn.ExistsResultData {
	existingIds := make(map[model.Id]bool, len(found.Existing))
	for _, keys := range found.Existing {
		if id := keys.GetModelId(models.ProductVariantFieldId); id != nil {
			existingIds[*id] = true
		}
	}

	existing := make([]model.Id, 0, len(requested))
	notExisting := make([]model.Id, 0)
	for _, id := range requested {
		if existingIds[id] {
			existing = append(existing, id)
			continue
		}
		notExisting = append(notExisting, id)
	}
	return dyn.ExistsResultData{Existing: existing, NotExisting: notExisting}
}
