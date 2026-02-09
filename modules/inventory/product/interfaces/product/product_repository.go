package product

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type ProductRepository interface {
	Create(ctx crud.Context, product *domain.Product) (*domain.Product, error)
	Update(ctx crud.Context, product *domain.Product, prevEtag model.Etag) (*domain.Product, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*domain.Product, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Product], error)
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
