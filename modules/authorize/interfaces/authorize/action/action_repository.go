package resource

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type ActionRepository interface {
	Create(ctx context.Context, action domain.Action) (*domain.Action, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.Action, error)
	FindByName(ctx context.Context, param FindByNameParam) (*domain.Action, error)
	Update(ctx context.Context, action domain.Action) (*domain.Action, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.Action], error)
}

type FindByIdParam = GetActionByIdQuery
type FindByNameParam = GetActionByNameCommand

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
