package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/productcategory"
)

func NewProductCategoryServiceImpl(
	prodCatRepo it.ProductCategoryRepository,
	cqrsBus cqrs.CqrsBus,
) it.ProductCategoryService {
	return &ProductCategoryServiceImpl{
		prodCatRepo: prodCatRepo,
		cqrsBus:     cqrsBus,
	}
}

type ProductCategoryServiceImpl struct {
	prodCatRepo it.ProductCategoryRepository
	cqrsBus     cqrs.CqrsBus
}

func (s *ProductCategoryServiceImpl) CreateProductCategory(ctx corectx.Context, cmd it.CreateProductCategoryCommand) (*it.CreateProductCategoryResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.ProductCategory, *domain.ProductCategory]{
		Action:         "create product category",
		BaseRepoGetter: s.prodCatRepo,
		Data:           cmd,
	})
}

func (s *ProductCategoryServiceImpl) UpdateProductCategory(ctx corectx.Context, cmd it.UpdateProductCategoryCommand) (*dyn.OpResult[dyn.MutateResultData], error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.ProductCategory, *domain.ProductCategory]{
		Action:       "update product category",
		DbRepoGetter: s.prodCatRepo,
		Data:         cmd,
	})
}

func (s *ProductCategoryServiceImpl) DeleteProductCategory(ctx corectx.Context, cmd it.DeleteProductCategoryCommand) (*it.DeleteProductCategoryResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete product category",
		DbRepoGetter: s.prodCatRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (s *ProductCategoryServiceImpl) GetProductCategory(ctx corectx.Context, query it.GetProductCategoryQuery) (*it.GetProductCategoryResult, error) {
	return corecrud.GetOne[domain.ProductCategory](ctx, corecrud.GetOneParam{
		Action:       "get product category",
		DbRepoGetter: s.prodCatRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (s *ProductCategoryServiceImpl) SearchProductCategories(ctx corectx.Context, query it.SearchProductCategoriesQuery) (*it.SearchProductCategoriesResult, error) {
	return corecrud.Search[domain.ProductCategory](ctx, corecrud.SearchParam{
		Action:       "search product categories",
		DbRepoGetter: s.prodCatRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (s *ProductCategoryServiceImpl) ProductCategoryExists(ctx corectx.Context, query it.ProductCategoryExistsQuery) (*it.ProductCategoryExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "product category exists",
		DbRepoGetter: s.prodCatRepo,
		Query:        dyn.ExistsQuery(query),
	})
}
