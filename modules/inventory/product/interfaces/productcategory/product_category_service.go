package productcategory

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type ProductCategoryService interface {
	CreateProductCategory(ctx crud.Context, cmd CreateProductCategoryCommand) (*CreateProductCategoryResult, error)
	UpdateProductCategory(ctx crud.Context, cmd UpdateProductCategoryCommand) (*UpdateProductCategoryResult, error)
	DeleteProductCategory(ctx crud.Context, cmd DeleteProductCategoryCommand) (*DeleteProductCategoryResult, error)
	GetProductCategoryById(ctx crud.Context, query GetProductCategoryByIdQuery) (*GetProductCategoryByIdResult, error)
	SearchProductCategories(ctx crud.Context, query SearchProductCategoriesQuery) (*SearchProductCategoriesResult, error)
}
