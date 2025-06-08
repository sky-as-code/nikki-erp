package group

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type GroupRepository interface {
	Create(ctx context.Context, group domain.Group) (*domain.Group, error)
	Delete(ctx context.Context, id model.Id) error
	FindById(ctx context.Context, param FindByIdParam) (*domain.Group, error)
	FindByName(ctx context.Context, name string) (*domain.Group, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, predicate *orm.Predicate, order []orm.OrderOption, opts crud.PagingOptions) (*crud.PagedResult[domain.Group], error)
	Update(ctx context.Context, group domain.Group) (*domain.Group, error)
}

type FindByIdParam = GetGroupByIdQuery
