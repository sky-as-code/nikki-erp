package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type ProductRepository interface {
	Create(ctx crud.Context, product *Product) (*Product, error)
	Update(ctx crud.Context, product *Product, prevEtag model.Etag) (*Product, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*Product, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[Product], error)
}

type ProductService interface {
	CreateProduct(ctx crud.Context, cmd CreateProductCommand) (*CreateProductResult, error)
	UpdateProduct(ctx crud.Context, cmd UpdateProductCommand) (*UpdateProductResult, error)
	DeleteProduct(ctx crud.Context, cmd DeleteProductCommand) (*DeleteProductResult, error)
	GetProductById(ctx crud.Context, query GetProductByIdQuery) (*GetProductByIdResult, error)
	SearchProducts(ctx crud.Context, query SearchProductsQuery) (*SearchProductsResult, error)
}

type DeleteParam = DeleteProductCommand
type FindByIdParam = GetProductByIdQuery
type SearchParam struct {
	Predicate    *orm.Predicate
	Order        []orm.OrderOption
	Page         int
	Size         int
	WithVariants bool
}
