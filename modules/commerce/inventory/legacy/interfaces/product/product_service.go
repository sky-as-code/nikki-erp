package product

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type ProductService interface {
	CreateProduct(ctx corectx.Context, cmd CreateProductCommand) (*CreateProductResult, error)
	DeleteProduct(ctx corectx.Context, cmd DeleteProductCommand) (*DeleteProductResult, error)
	GetProduct(ctx corectx.Context, query GetProductQuery) (*GetProductResult, error)
	ProductExists(ctx corectx.Context, query ProductExistsQuery) (*ProductExistsResult, error)
	SearchProducts(ctx corectx.Context, query SearchProductsQuery) (*SearchProductsResult, error)
	SetProductIsArchived(ctx corectx.Context, cmd SetProductIsArchivedCommand) (*SetProductIsArchivedResult, error)
	UpdateProduct(ctx corectx.Context, cmd UpdateProductCommand) (*dyn.OpResult[dyn.MutateResultData], error)
}
