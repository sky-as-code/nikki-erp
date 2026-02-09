package product

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type ProductService interface {
	CreateProduct(ctx crud.Context, cmd CreateProductCommand) (*CreateProductResult, error)
	UpdateProduct(ctx crud.Context, cmd UpdateProductCommand) (*UpdateProductResult, error)
	DeleteProduct(ctx crud.Context, cmd DeleteProductCommand) (*DeleteProductResult, error)
	GetProductById(ctx crud.Context, query GetProductByIdQuery) (*GetProductByIdResult, error)
	SearchProducts(ctx crud.Context, query SearchProductsQuery) (*SearchProductsResult, error)
}
