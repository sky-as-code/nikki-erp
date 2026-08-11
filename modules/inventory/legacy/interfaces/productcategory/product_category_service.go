package productcategory

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type ProductCategoryService interface {
	CreateProductCategory(ctx corectx.Context, cmd CreateProductCategoryCommand) (*CreateProductCategoryResult, error)
	DeleteProductCategory(ctx corectx.Context, cmd DeleteProductCategoryCommand) (*DeleteProductCategoryResult, error)
	ProductCategoryExists(ctx corectx.Context, query ProductCategoryExistsQuery) (*ProductCategoryExistsResult, error)
	GetProductCategory(ctx corectx.Context, query GetProductCategoryQuery) (*GetProductCategoryResult, error)
	SearchProductCategories(ctx corectx.Context, query SearchProductCategoriesQuery) (*SearchProductCategoriesResult, error)
	UpdateProductCategory(ctx corectx.Context, cmd UpdateProductCategoryCommand) (*dyn.OpResult[dyn.MutateResultData], error)
}
