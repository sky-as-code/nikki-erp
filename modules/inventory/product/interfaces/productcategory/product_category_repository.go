package productcategory

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

type ProductCategoryRepository interface {
	Create(ctx crud.Context, productCategory *domain.ProductCategory) (*domain.ProductCategory, error)
	Update(ctx crud.Context, productCategory *domain.ProductCategory, prevEtag model.Etag) (*domain.ProductCategory, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*domain.ProductCategory, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.ProductCategory], error)
}

type DeleteParam = DeleteProductCategoryCommand
type FindByIdParam = GetProductCategoryByIdQuery

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
