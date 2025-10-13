package resource

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type ResourceRepository interface {
	Create(ctx crud.Context, resource *domain.Resource) (*domain.Resource, error)
	FindByName(ctx crud.Context, param FindByNameParam) (*domain.Resource, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.Resource, error)
	Update(ctx crud.Context, resource *domain.Resource, prevEtag model.Etag) (*domain.Resource, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.Resource], error)
	Exists(ctx crud.Context, param ExistsParam) (bool, error)
}

type FindByIdParam = GetResourceByIdQuery
type FindByNameParam = GetResourceByNameQuery
type DeleteParam = DeleteResourceHardByNameQuery
type ExistsParam = ExistsResourceQuery

type SearchParam struct {
	Predicate   *orm.Predicate
	Order       []orm.OrderOption
	Page        int
	Size        int
	WithActions bool
}
